//nolint:gosec // G115: Ignore integer overflow warnings for this file
package qh

import (
	"fmt"
	"strings"
)

const (
	hexPaddingWidth        = 28  // Width to pad hex output for alignment with labels
	bodyPreviewMaxLength   = 50  // Maximum length of body preview before truncation
	stringInlineThreshold  = 20  // Maximum string length to display inline (vs multiline)
	maxHexBytesDisplay     = 128 // Maximum hex bytes to display for long strings
	hexChunkSize           = 16  // Number of bytes per line in hex dump
	stringPreviewMaxLength = 100 // Maximum length of decoded string preview
)

func DebugRequest(data []byte) string {
	if len(data) == 0 {
		return "    (empty)\n"
	}

	var sb strings.Builder
	offset := 0

	// First byte: Version + Method
	if offset < len(data) {
		firstByte := data[offset]
		version := firstByte >> 6
		method := Method((firstByte >> 3) & 0b00000111)
		sb.WriteString(
			fmt.Sprintf(
				"    \\x%02x                           First byte (Version=%d, Method=%s)\n",
				firstByte,
				version,
				method.String(),
			),
		)
		offset++
	}

	hostLen := annotateVarint(&sb, data, &offset, "Host length")
	hostLen = min(hostLen, uint64(len(data)-offset))
	annotateString(&sb, data, &offset, int(hostLen), "Host")

	pathLen := annotateVarint(&sb, data, &offset, "Path length")
	pathLen = min(pathLen, uint64(len(data)-offset))
	annotateString(&sb, data, &offset, int(pathLen), "Path")

	headersLen := annotateVarint(&sb, data, &offset, "Headers length")
	headersEndOffset := min(offset+int(headersLen), len(data))
	annotateHeaders(&sb, data, &offset, headersEndOffset, true)
	offset = headersEndOffset

	bodyLen := annotateVarint(&sb, data, &offset, "Body length")
	if bodyLen > 0 && offset+int(bodyLen) <= len(data) {
		bodyPreview := string(data[offset : offset+int(bodyLen)])
		if len(bodyPreview) > bodyPreviewMaxLength {
			bodyPreview = bodyPreview[:bodyPreviewMaxLength] + "..."
		}
		fmt.Fprintf(&sb, "    (body data)                  Body: %s\n", bodyPreview)
		offset += int(bodyLen)
	}

	fmt.Fprintf(&sb, "    (parsed %d / %d bytes)\n", offset, len(data))

	return sb.String()
}

func DebugResponse(data []byte) string {
	if len(data) == 0 {
		return "    (empty)\n"
	}

	var sb strings.Builder
	offset := 0

	// First byte: Version + Status
	if offset < len(data) {
		firstByte := data[offset]
		version := firstByte >> 6
		statusCompact := firstByte & 0b00111111
		statusDecoded := DecodeStatusCode(statusCompact)
		sb.WriteString(
			fmt.Sprintf(
				"    \\x%02x                           First byte (Version=%d, Status=%d)\n",
				firstByte,
				version,
				statusDecoded,
			),
		)
		offset++
	}

	headersLen := annotateVarint(&sb, data, &offset, "Headers length")
	headersEndOffset := min(offset+int(headersLen), len(data))
	annotateHeaders(&sb, data, &offset, headersEndOffset, false)
	offset = headersEndOffset

	bodyLen := annotateVarint(&sb, data, &offset, "Body length")
	if bodyLen > 0 && offset+int(bodyLen) <= len(data) {
		bodyPreview := string(data[offset : offset+int(bodyLen)])
		if len(bodyPreview) > bodyPreviewMaxLength {
			bodyPreview = bodyPreview[:bodyPreviewMaxLength] + "..."
		}
		fmt.Fprintf(&sb, "    (body data)                  Body: %s\n", bodyPreview)
		offset += int(bodyLen)
	}

	fmt.Fprintf(&sb, "    (parsed %d / %d bytes)\n", offset, len(data))

	return sb.String()
}

func writeHex(sb *strings.Builder, data []byte) {
	for i, b := range data {
		if i > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprintf(sb, "\\x%02x", b)
	}
}

func annotateVarint(sb *strings.Builder, data []byte, offset *int, label string) uint64 {
	if *offset >= len(data) {
		return 0
	}
	value, n, _ := ReadUvarint(data, *offset)

	sb.WriteString("    ")
	writeHex(sb, data[*offset:*offset+n])
	for i := n * 5; i < hexPaddingWidth; i++ {
		sb.WriteByte(' ')
	}
	fmt.Fprintf(sb, " %s: %d\n", label, value)

	*offset += n
	return value
}

func annotateString(sb *strings.Builder, data []byte, offset *int, length int, label string) {
	if *offset+length > len(data) {
		return
	}
	value := string(data[*offset : *offset+length])
	hexData := data[*offset : *offset+length]

	if length <= stringInlineThreshold {
		annotateShortString(sb, hexData, value, length, label)
	} else {
		annotateLongString(sb, hexData, value, length, label)
	}

	*offset += length
}

func annotateShortString(sb *strings.Builder, hexData []byte, value string, length int, label string) {
	sb.WriteString("    ")
	writeHex(sb, hexData)
	for i := length * 5; i < hexPaddingWidth; i++ {
		sb.WriteByte(' ')
	}
	fmt.Fprintf(sb, " %s: %s\n", label, value)
}

func annotateLongString(sb *strings.Builder, hexData []byte, value string, length int, label string) {
	displayLength := length
	truncated := false
	if displayLength > maxHexBytesDisplay {
		displayLength = maxHexBytesDisplay
		truncated = true
	}

	fmt.Fprintf(sb, "    %s:\n", label)

	// Print hex in chunks
	for i := 0; i < displayLength; i += hexChunkSize {
		sb.WriteString("      ")
		end := i + hexChunkSize
		if end > displayLength {
			end = displayLength
		}
		writeHex(sb, hexData[i:end])
		sb.WriteString("\n")
	}

	if truncated {
		fmt.Fprintf(sb, "      ... (%d more bytes)\n", length-maxHexBytesDisplay)
	}

	if len(value) > stringPreviewMaxLength {
		fmt.Fprintf(sb, "      → %s...\n", value[:stringPreviewMaxLength])
	} else {
		fmt.Fprintf(sb, "      → %s\n", value)
	}
}

//nolint:nestif // acceptable, maybe fix later
func annotateHeaders(
	sb *strings.Builder,
	data []byte,
	offset *int,
	endOffset int,
	isRequest bool,
) {
	for *offset < endOffset && *offset < len(data) {
		headerID := data[*offset]

		if headerID == 0x00 {
			sb.WriteString("    \\x00                           Custom header\n")
			*offset++
			if *offset >= len(data) {
				break
			}

			keyLen := annotateVarint(sb, data, offset, "Key length")
			annotateString(sb, data, offset, int(keyLen), "Key")
			valueLen := annotateVarint(sb, data, offset, "Value length")
			annotateString(sb, data, offset, int(valueLen), "Value")
		} else {
			var headerName string
			var hasValue bool
			if isRequest {
				if entry, ok := requestHeaderStaticTable[headerID]; ok {
					headerName = entry.name
					hasValue = entry.value == ""
				}
			} else {
				if entry, ok := responseHeaderStaticTable[headerID]; ok {
					headerName = entry.name
					hasValue = entry.value == ""
				}
			}

			if headerName != "" {
				fmt.Fprintf(sb, "    \\x%02x                           Header ID (%s)\n", headerID, headerName)
			} else {
				fmt.Fprintf(sb, "    \\x%02x                           Header ID (unknown)\n", headerID)
			}
			*offset++

			// Only read value if it's Format 2 (name-only header)
			if hasValue && *offset < len(data) {
				valueLen := annotateVarint(sb, data, offset, "Value length")
				annotateString(sb, data, offset, int(valueLen), "Value")
			}
		}
	}
}
