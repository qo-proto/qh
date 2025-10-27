package main

import (
	"bufio"
	"encoding/hex"
	"errors"
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

var errUsage = errors.New("invalid arguments")

// Decryptor holds the state for the decryption process.
type Decryptor struct {
	secrets    map[uint64][]byte
	pcapWriter *pcapgo.Writer
}

// NewDecryptor creates and initializes a new Decryptor.
func NewDecryptor(keyLogFile string) (*Decryptor, error) {
	secrets, err := loadSecrets(keyLogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load keylog file: %w", err)
	}
	slog.Info("Successfully loaded secrets", "count", len(secrets))
	return &Decryptor{secrets: secrets}, nil
}

// ProcessPacket handles the decryption and writing of a single packet.
func (d *Decryptor) ProcessPacket(packet gopacket.Packet) error {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return d.pcapWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
	}
	udp, _ := udpLayer.(*layers.UDP)

	// Check if it's a QOTP Data packet that we can decrypt.
	isQOTP := udp.DstPort == 8090 || udp.SrcPort == 8090
	isDataPacket := len(udp.Payload) > 0 && qotp.CryptoMsgType(udp.Payload[0]>>5) == qotp.Data

	if !isQOTP || !isDataPacket {
		return d.pcapWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
	}

	connID := qotp.Uint64(udp.Payload[1:9])
	secret, ok := d.secrets[connID]
	if !ok {
		slog.Debug("No secret found for connection, writing original packet", "connID", connID)
		return d.pcapWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
	}

	// Decrypt the payload.
	isFromServer := udp.SrcPort == 8090
	encryptedPortion := udp.Payload[qotp.HeaderSize+qotp.ConnIdSize:]
	decryptedPayload, err := qotp.DecryptDataForPcap(encryptedPortion, !isFromServer, 0, secret, connID)
	if err != nil {
		slog.Error("Decryption failed, writing original packet", "error", err)
		return d.pcapWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
	}

	slog.Info("Successfully decrypted payload", "size", len(decryptedPayload))

	// Re-serialize the packet with the decrypted payload.
	return d.writeDecryptedPacket(packet, udp, decryptedPayload)
}

// writeDecryptedPacket re-serializes the packet with the new payload.
func (d *Decryptor) writeDecryptedPacket(packet gopacket.Packet, originalUDP *layers.UDP, payload []byte) error {
	newUDPLayer := &layers.UDP{
		SrcPort: originalUDP.SrcPort,
		DstPort: originalUDP.DstPort,
	}

	if ipLayer, ok := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4); ok {
		if err := newUDPLayer.SetNetworkLayerForChecksum(ipLayer); err != nil {
			slog.Warn("Failed to set network layer for checksum", "error", err)
		}
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}

	layersToSerialize := []gopacket.SerializableLayer{}
	for _, layer := range packet.Layers() {
		if layer.LayerType() == layers.LayerTypeUDP {
			break // Stop at the UDP layer.
		}
		if serializable, ok := layer.(gopacket.SerializableLayer); ok {
			layersToSerialize = append(layersToSerialize, serializable)
		}
	}
	layersToSerialize = append(layersToSerialize, newUDPLayer, gopacket.Payload(payload))

	if err := gopacket.SerializeLayers(buf, opts, layersToSerialize...); err != nil {
		slog.Error("Failed to serialize packet, writing original", "error", err)
		return d.pcapWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
	}

	newCI := packet.Metadata().CaptureInfo
	newCI.CaptureLength = len(buf.Bytes())
	newCI.Length = len(buf.Bytes())

	return d.pcapWriter.WritePacket(newCI, buf.Bytes())
}

// run is the main application logic, moved out of main to allow for clean error handling.
func run() error {
	if len(os.Args) != 4 {
		return errUsage
	}
	keyLogFile, inputFile, outputFile := os.Args[1], os.Args[2], os.Args[3]
	slog.Info("Starting QOTP decryption", "keylog", keyLogFile, "input", inputFile, "output", outputFile)

	decryptor, err := NewDecryptor(keyLogFile)
	if err != nil {
		return err
	}

	handle, err := pcap.OpenOffline(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input pcap: %w", err)
	}
	defer handle.Close()

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output pcap: %w", err)
	}
	defer f.Close()

	decryptor.pcapWriter = pcapgo.NewWriter(f)
	const snapshotLength = 65535
	if err := decryptor.pcapWriter.WriteFileHeader(snapshotLength, handle.LinkType()); err != nil {
		return fmt.Errorf("failed to write pcap header: %w", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if err := decryptor.ProcessPacket(packet); err != nil {
			slog.Error("Failed to process and write packet", "error", err)
		}
	}

	slog.Info("Decryption complete", "output_file", outputFile)
	return nil
}

func main() {
	if err := run(); err != nil {
		if errors.Is(err, errUsage) {
			_, _ = fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/qotp-decrypt <keylog_file> <input_pcap> <output_pcap>")
		} else {
			slog.Error("Application failed", "error", err)
		}
		os.Exit(1)
	}
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
			slog.Warn("Skipping invalid connection ID in keylog", "value", parts[1], "error", err)
			continue
		}
		secret, err := hex.DecodeString(parts[2])
		if err != nil {
			slog.Warn("Skipping invalid secret in keylog", "connID", connID, "error", err)
			continue
		}
		secrets[connID] = secret
	}
	return secrets, scanner.Err()
}
