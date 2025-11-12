package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/hex"
	"fmt"
	"unsafe"

	"github.com/tbocek/qotp"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
)

const Version = "1.0.0"

// Helper to read uint48 from buffer (not exported in qotp package)
func uint48(b []byte) uint64 {
	return uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40
}

// Helper to write uint64 to buffer
func putUint64(b []byte, v uint64) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

// Helper to write uint48 to buffer
func putUint48(b []byte, v uint64) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
}

// Global storage for shared secrets keyed by connection ID
var sharedSecrets = make(map[uint64][]byte)
var currentEpochs = make(map[uint64]uint64)

//export SetSharedSecret
func SetSharedSecret(connId C.ulonglong, secret *C.char, secretLen C.int) {
	secretBytes := C.GoBytes(unsafe.Pointer(secret), secretLen)
	sharedSecrets[uint64(connId)] = secretBytes
}

//export SetSharedSecretHex
func SetSharedSecretHex(connId C.ulonglong, secretHex *C.char) C.int {
	secretHexStr := C.GoString(secretHex)
	secretBytes, err := hex.DecodeString(secretHexStr)
	if err != nil {
		fmt.Printf("[qotp_crypto] ERROR: Invalid hex string for connection %d\n", connId)
		return -1
	}
	sharedSecrets[uint64(connId)] = secretBytes
	fmt.Printf("[qotp_crypto] Key loaded for connection 0x%x (%d bytes)\n", connId, len(secretBytes))
	return 0
}

//export GetVersion
func GetVersion() *C.char {
	return C.CString(Version)
}

//export GetLoadedKeyCount
func GetLoadedKeyCount() C.int {
	return C.int(len(sharedSecrets))
}

//export GetLoadedKeys
func GetLoadedKeys(output *C.ulonglong, maxCount C.int) C.int {
	count := 0
	for connId := range sharedSecrets {
		if count >= int(maxCount) {
			break
		}
		*(*C.ulonglong)(unsafe.Pointer(uintptr(unsafe.Pointer(output)) + uintptr(count*8))) = C.ulonglong(connId)
		count++
	}
	return C.int(count)
}

//export GetConnectionId
func GetConnectionId(data *C.char, dataLen C.int) C.ulonglong {
	if dataLen < qotp.HeaderSize+qotp.ConnIdSize {
		return 0
	}

	dataBytes := C.GoBytes(unsafe.Pointer(data), dataLen)

	// Connection ID is at bytes 1-8 (little-endian)
	connId := uint64(dataBytes[1]) |
		uint64(dataBytes[2])<<8 |
		uint64(dataBytes[3])<<16 |
		uint64(dataBytes[4])<<24 |
		uint64(dataBytes[5])<<32 |
		uint64(dataBytes[6])<<40 |
		uint64(dataBytes[7])<<48 |
		uint64(dataBytes[8])<<56

	return C.ulonglong(connId)
}

//export GetMessageType
func GetMessageType(data *C.char, dataLen C.int) C.int {
	if dataLen < 1 {
		return -1
	}

	dataBytes := C.GoBytes(unsafe.Pointer(data), dataLen)
	msgType := (dataBytes[0] >> 5) & 0x07

	return C.int(msgType)
}

//export DecryptDataPacket
func DecryptDataPacket(
	encryptedData *C.char,
	encryptedLen C.int,
	connId C.ulonglong,
	isSender C.int,
	epoch C.ulonglong,
	output *C.char,
	outputMaxLen C.int) C.int {

	// Get shared secret for this connection
	secret, ok := sharedSecrets[uint64(connId)]
	if !ok {
		return -1 // No shared secret for this connection
	}

	// Convert encrypted data to Go slice
	encBytes := C.GoBytes(unsafe.Pointer(encryptedData), encryptedLen)

	// Call the actual decryption function
	decrypted, err := decryptDataForPcap(encBytes, isSender != 0, uint64(epoch), secret, uint64(connId))
	if err != nil {
		return -2 // Decryption failed
	}

	if len(decrypted) > int(outputMaxLen) {
		return -3 // Output buffer too small
	}

	// Copy decrypted data to output buffer
	for i := 0; i < len(decrypted); i++ {
		*(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(output)) + uintptr(i))) = C.char(decrypted[i])
	}

	return C.int(len(decrypted))
}

// decryptDataForPcap is adapted from DecryptDataForPcap in crypto.go
func decryptDataForPcap(encryptedPortion []byte, isSenderOnInit bool, epoch uint64, sharedSecret []byte, connId uint64) ([]byte, error) {
	// Reconstruct the AAD (header) as it was during encryption for a Data packet.
	header := make([]byte, qotp.HeaderSize+qotp.ConnIdSize)
	header[0] = (uint8(qotp.Data) << 5) | qotp.CryptoVersion
	putUint64(header[qotp.HeaderSize:], connId)

	// For offline decryption, we must call chainedDecrypt with the offline flag.
	_, _, payload, err := chainedDecrypt(isSenderOnInit, epoch, sharedSecret, header, encryptedPortion, true)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

// chainedDecrypt from crypto.go - adapted for pcap decryption
func chainedDecrypt(isSender bool, epochCrypt uint64, sharedSecret []byte, header []byte, encData []byte, isOfflineDecrypt bool) (
	snConn uint64, currentEpochCrypt uint64, packetData []byte, err error) {

	if len(encData) < qotp.SnSize {
		return 0, 0, nil, fmt.Errorf("encrypted data too short")
	}

	snConnBytes := make([]byte, qotp.SnSize)
	encryptedSN := encData[:qotp.SnSize]
	ciphertextWithNonce := encData[qotp.SnSize:]

	snConnBytes, err = openNoVerify(sharedSecret, ciphertextWithNonce, encryptedSN, snConnBytes)
	if err != nil {
		return 0, 0, nil, err
	}
	snConn = uint48(snConnBytes)

	nonceDet := make([]byte, chacha20poly1305.NonceSize)

	epochs := []uint64{epochCrypt}
	// Only try previous epoch if > 0
	if epochCrypt > 0 {
		epochs = append(epochs, epochCrypt-1)
	}
	epochs = append(epochs, epochCrypt+1)

	aead, err := chacha20poly1305.New(sharedSecret)
	if err != nil {
		return 0, 0, nil, err
	}
	putUint48(nonceDet[6:], snConn)

	for _, epochTry := range epochs {
		putUint48(nonceDet, epochTry)

		// The logic for the sender bit is inverted between live decryption and offline pcap decryption.
		if isOfflineDecrypt {
			if isSender {
				nonceDet[0] = nonceDet[0] | 0x80 // bit set
			} else {
				nonceDet[0] = nonceDet[0] &^ 0x80 // bit clear
			}
		} else {
			if isSender {
				nonceDet[0] = nonceDet[0] &^ 0x80 // bit clear
			} else {
				nonceDet[0] = nonceDet[0] | 0x80 // bit set
			}
		}

		packetData, err = aead.Open(nil, nonceDet, ciphertextWithNonce, header)
		if err == nil {
			return snConn, epochTry, packetData, nil
		}
	}
	return 0, 0, nil, err
}

// openNoVerify from crypto.go
func openNoVerify(sharedSecret []byte, nonce []byte, encoded []byte, snSer []byte) ([]byte, error) {
	// The nonce for the SN is the first 24 bytes of the main ciphertext
	snNonce := nonce[:chacha20poly1305.NonceSizeX]

	s, err := chacha20.NewUnauthenticatedCipher(sharedSecret, snNonce)
	if err != nil {
		return nil, err
	}
	s.SetCounter(1) // Per the spec, skip the first 32 bytes of the keystream for SN encryption.

	// Decrypt the ciphertext
	s.XORKeyStream(snSer, encoded)

	return snSer, nil
}

func main() {
	// Required for buildmode=c-shared
}
