package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/tbocek/qotp"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run ./cmd/qotp-decrypt <keylog_file> <input_pcap> <output_pcap>")
		os.Exit(1)
	}

	keyLogFile := os.Args[1]
	inputFile := os.Args[2]
	outputFile := os.Args[3]

	slog.Info("Starting QOTP decryption", "keylog", keyLogFile, "input", inputFile, "output", outputFile)

	// 1. Load session secrets from the keylog file
	secrets, err := loadSecrets(keyLogFile)
	if err != nil {
		slog.Error("Failed to load keylog file", "error", err)
		os.Exit(1)
	}
	slog.Info("Successfully loaded secrets", "count", len(secrets))

	// 2. Open the input pcap file
	handle, err := pcap.OpenOffline(inputFile)
	if err != nil {
		slog.Error("Failed to open input pcap file", "error", err)
		os.Exit(1)
	}
	defer handle.Close()

	// 3. Create the output pcap file
	f, err := os.Create(outputFile)
	if err != nil {
		slog.Error("Failed to create output pcap file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	w := pcapgo.NewWriter(f)
	// Set a generous snapshot length for the output file. 65535 is a common standard.
	const snapshotLength = 65535
	if err := w.WriteFileHeader(snapshotLength, handle.LinkType()); err != nil {
		slog.Error("Failed to write pcap header", "error", err)
		os.Exit(1)
	}

	// 4. Process each packet
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer == nil {
			// Not a UDP packet, skip
			continue
		}

		udp, _ := udpLayer.(*layers.UDP)
		isQOTP := udp.DstPort == 8090 || udp.SrcPort == 8090

		// Not a QOTP packet, skip
		if !isQOTP {
			continue
		}

		// Check if it's a QOTP Data packet (type 4)
		msgType := qotp.CryptoMsgType(udp.Payload[0] >> 5)
		if msgType != qotp.Data {
			if err := w.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
				slog.Error("Failed to write non-Data packet", "error", err)
			}
			continue
		}

		connID := qotp.Uint64(udp.Payload[1:9])
		secret, ok := secrets[connID]
		if !ok {
			// No secret for this connection, write original packet
			if err := w.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
				slog.Error("Failed to write packet with no secret", "error", err)
			}
			slog.Info("No Secret for this connection")
			continue
		}

		// We have a secret, let's decrypt!
		// We need to determine the epoch. For simplicity, we'll assume epoch 0.
		// A more advanced tool could track epochs.
		// The 'isSenderOnInit' flag is true if the packet is from the original sender (client).
		// The server is the one listening on the well-known port.
		isFromServer := udp.SrcPort == 8090
		isSenderOnInit := !isFromServer

		// Pass only the encrypted part of the payload, which starts after the QOTP header and Connection ID.
		encryptedPortion := udp.Payload[qotp.HeaderSize+qotp.ConnIdSize:]
		decryptedPayload, err := qotp.DecryptDataForPcap(encryptedPortion, isSenderOnInit, 0, secret, connID)
		if err != nil {
			// Decryption failed, write original packet
			if err := w.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
				slog.Error("Failed to write original packet after decryption failure", "error", err)
			}
			slog.Error("Decryption failed", "error", err)
			continue
		}

		// Log a preview of the decrypted payload
		previewLen := min(64, len(decryptedPayload))
		previewBytes := decryptedPayload[:previewLen]
		slog.Info("Successfully decrypted payload", "size", len(decryptedPayload),
			"preview_str", string(previewBytes), "preview_hex", hex.EncodeToString(previewBytes))

		// To ensure the decrypted payload is written, we create a new UDP layer
		// and serialize it with the lower-level layers from the original packet.
		newUDPLayer := &layers.UDP{
			SrcPort: udp.SrcPort,
			DstPort: udp.DstPort,
		}

		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ip, ok := ipLayer.(*layers.IPv4); ok {
			newUDPLayer.SetNetworkLayerForChecksum(ip)
		}

		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}

		// Build the list of layers to serialize. We must use the concrete, serializable
		// layer types, not the generic interfaces returned by packet.LinkLayer(), etc.
		layersToSerialize := []gopacket.SerializableLayer{}
		for _, layer := range packet.Layers() {
			// We only want to re-serialize the layers *below* the UDP layer.
			if layer.LayerType() == layers.LayerTypeUDP {
				break
			}
			if serializable, ok := layer.(gopacket.SerializableLayer); ok {
				layersToSerialize = append(layersToSerialize, serializable)
			}
		}
		// Append our new UDP layer and its decrypted payload
		// The payload must be added as a gopacket.Payload layer after the UDP layer.
		layersToSerialize = append(layersToSerialize, newUDPLayer, gopacket.Payload(decryptedPayload))

		err = gopacket.SerializeLayers(buf, opts, layersToSerialize...)
		if err != nil {
			slog.Error("Failed to serialize packet layers", "error", err)
			// Write the original packet on serialization failure to avoid losing it
			if err := w.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
				slog.Error("Failed to write original packet after serialization failure", "error", err)
			}
			continue
		}

		// Create a new CaptureInfo with the correct, updated length.
		newCI := packet.Metadata().CaptureInfo
		newCI.CaptureLength = len(buf.Bytes())
		newCI.Length = len(buf.Bytes())

		if err := w.WritePacket(newCI, buf.Bytes()); err != nil {
			slog.Error("Failed to write modified packet", "error", err)
		}
	}

	slog.Info("Decryption complete", "output_file", outputFile)
}

func loadSecrets(filename string) (map[uint64][]byte, error) {
	secrets := make(map[uint64][]byte)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 3 || parts[0] != "QOTP_SHARED_SECRET" {
			continue
		}
		connID, err := strconv.ParseUint(parts[1], 16, 64)
		if err != nil {
			continue
		}
		secret, err := hex.DecodeString(parts[2])
		if err != nil {
			continue
		}
		secrets[connID] = secret
	}
	return secrets, scanner.Err()
}
