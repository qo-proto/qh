//nolint:gosec // G115: Ignore integer overflow warnings for this file
package qh

import (
	"fmt"
	"strings"
)

const (
	customHeaderID         = 0x00 // Header ID indicating a custom header
	stringInlineThreshold  = 15   // Maximum string length to display inline (vs multiline)
	maxHexBytesDisplay     = 128  // Maximum hex bytes to display for long strings
	hexChunkSize           = 16   // Number of bytes per line in hex dump
	stringPreviewMaxLength = 100  // Maximum length of decoded string preview
	offsetColumnWidth      = 8    // Width of "0xXXXX  " (6 chars + 2 spaces)
	bytesColumnWidth       = 49   // Width of the hex bytes column (fits ~15 bytes with spacing)
	continuationIndent     = 10   // Indentation for continuation lines (aligns under first hex byte)
	nestedFieldIndent      = "  " // Indentation for nested fields (custom headers, header values)
)

func DebugRequest(data []byte) string {
	if len(data) == 0 {
		return "(empty)\n"
	}

	var sb strings.Builder
	offset := 0

	sb.WriteString("OFFSET  BYTES                                            DESCRIPTION\n")

	// First byte: Version + Method
	if offset < len(data) {
		firstByte := data[offset]
		version := firstByte >> versionBitShift
		method := Method((firstByte >> methodBitShift) & methodMask)
		writeTableRow(&sb, offset, data[offset:offset+1],
			fmt.Sprintf("First byte (Version=%d, Method=%s)", version, method.String()))
		offset++
	}

	hostLen := annotateVarint(&sb, data, &offset, "Host length")
	hostLen = min(hostLen, uint64(len(data)-offset))
	annotateString(&sb, data, &offset, int(hostLen), "Host")

	pathLen := annotateVarint(&sb, data, &offset, "Path length")
	pathLen = min(pathLen, uint64(len(data)-offset))
	annotateString(&sb, data, &offset, int(pathLen), "Path")

	headersLen := annotateVarint(&sb, data, &offset, "Headers length")
	sb.WriteString("\n") // Blank line before headers section
	headersEndOffset := min(offset+int(headersLen), len(data))
	annotateHeaders(&sb, data, &offset, headersEndOffset, true)
	offset = headersEndOffset

	sb.WriteString("\n") // Blank line before body
	annotateVarint(&sb, data, &offset, "Body length")

	sb.WriteString("\n")
	fmt.Fprintf(&sb, "Summary: parsed %d / %d bytes\n", offset, len(data))

	return sb.String()
}

func DebugResponse(data []byte) string {
	if len(data) == 0 {
		return "(empty)\n"
	}

	var sb strings.Builder
	offset := 0

	sb.WriteString("OFFSET  BYTES                                            DESCRIPTION\n")

	// First byte: Version + Status
	if offset < len(data) {
		firstByte := data[offset]
		version := firstByte >> versionBitShift
		statusCompact := firstByte & statusCodeMask
		statusDecoded := DecodeStatusCode(statusCompact)
		writeTableRow(&sb, offset, data[offset:offset+1],
			fmt.Sprintf("First byte (Version=%d, Status=%d)", version, statusDecoded))
		offset++
	}

	headersLen := annotateVarint(&sb, data, &offset, "Headers length")
	sb.WriteString("\n") // Blank line before headers section
	headersEndOffset := min(offset+int(headersLen), len(data))
	annotateHeaders(&sb, data, &offset, headersEndOffset, false)
	offset = headersEndOffset

	sb.WriteString("\n") // Blank line before body
	annotateVarint(&sb, data, &offset, "Body length")

	sb.WriteString("\n")
	fmt.Fprintf(&sb, "Summary: parsed %d / %d bytes\n", offset, len(data))

	return sb.String()
}

func writeTableRow(sb *strings.Builder, offset int, bytes []byte, description string) {
	fmt.Fprintf(sb, "0x%04x  ", offset)

	hexStr := formatHex(bytes)
	sb.WriteString(hexStr)

	if len(hexStr) < bytesColumnWidth {
		for i := len(hexStr); i < bytesColumnWidth; i++ {
			sb.WriteByte(' ')
		}
	} else {
		sb.WriteByte(' ') // Just one space if overflow
	}

	sb.WriteString(description)
	sb.WriteString("\n")
}

func writeTableRowMultiline(sb *strings.Builder, offset int, bytes []byte, description string) {
	if len(bytes) <= hexChunkSize {
		writeTableRow(sb, offset, bytes, description)
		return
	}

	// Limit bytes to display
	displayBytes := bytes
	truncated := false
	if len(bytes) > maxHexBytesDisplay {
		displayBytes = bytes[:maxHexBytesDisplay]
		truncated = true
	}

	// First line with description
	firstChunk := displayBytes[:hexChunkSize]
	writeTableRow(sb, offset, firstChunk, description)

	// Continuation lines (aligned under first hex byte)
	for i := hexChunkSize; i < len(displayBytes); i += hexChunkSize {
		end := min(i+hexChunkSize, len(displayBytes))
		chunk := displayBytes[i:end]
		hexStr := formatHex(chunk)

		fmt.Fprintf(sb, "%s%s\n", strings.Repeat(" ", continuationIndent), hexStr)
	}

	if truncated {
		descriptionColumnStart := offsetColumnWidth + bytesColumnWidth
		fmt.Fprintf(sb, "%s[%d bytes total; %d shown]\n",
			strings.Repeat(" ", descriptionColumnStart), len(bytes), len(displayBytes))
	}
}

func formatHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, b := range data {
		if i > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprintf(&sb, "%02x", b)
	}
	return sb.String()
}

func annotateVarint(sb *strings.Builder, data []byte, offset *int, label string) uint64 {
	if *offset >= len(data) {
		return 0
	}
	value, n, _ := ReadUvarint(data, *offset)

	writeTableRow(sb, *offset, data[*offset:*offset+n], fmt.Sprintf("%s: %d", label, value))

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
		// Short string: inline on one line
		writeTableRow(sb, *offset, hexData, fmt.Sprintf("%s: %s", label, value))
	} else {
		// Long string: multiline with truncation info
		displayValue := value
		suffix := ""
		if len(value) > stringPreviewMaxLength {
			displayValue = value[:stringPreviewMaxLength]
			suffix = fmt.Sprintf("... [%d bytes total, showing %d]", len(value), stringPreviewMaxLength)
		}
		writeTableRowMultiline(sb, *offset, hexData, fmt.Sprintf("%s: %s%s", label, displayValue, suffix))
	}

	*offset += length
}

func annotateHeaders(
	sb *strings.Builder,
	data []byte,
	offset *int,
	endOffset int,
	isRequest bool,
) {
	for *offset < endOffset && *offset < len(data) {
		headerID := data[*offset]

		if headerID == customHeaderID {
			annotateCustomHeader(sb, data, offset)
		} else {
			annotateStaticTableHeader(sb, data, offset, headerID, isRequest)
		}
	}
}

func annotateCustomHeader(sb *strings.Builder, data []byte, offset *int) {
	writeTableRow(sb, *offset, []byte{data[*offset]}, "Custom header")
	*offset++
	if *offset >= len(data) {
		return
	}

	keyLen := annotateVarint(sb, data, offset, nestedFieldIndent+"Key length")
	annotateString(sb, data, offset, int(keyLen), nestedFieldIndent+"Key")
	valueLen := annotateVarint(sb, data, offset, nestedFieldIndent+"Value length")
	annotateString(sb, data, offset, int(valueLen), nestedFieldIndent+"Value")
}

func annotateStaticTableHeader(sb *strings.Builder, data []byte, offset *int, headerID byte, isRequest bool) {
	headerName, headerValue, valueFollows := lookupHeaderInStaticTable(headerID, isRequest)

	if headerName != "" {
		if valueFollows {
			// Format 2: name-only in static table, value bytes follow
			writeTableRow(sb, *offset, []byte{headerID}, fmt.Sprintf("Header ID (%s)", headerName))
		} else {
			// Format 1: complete name+value pair in static table
			writeTableRow(sb, *offset, []byte{headerID}, fmt.Sprintf("Header ID (%s: %s)", headerName, headerValue))
		}
	} else {
		writeTableRow(sb, *offset, []byte{headerID}, "Header ID (unknown)")
	}
	*offset++

	// Read value bytes if Format 2
	if valueFollows && *offset < len(data) {
		valueLen := annotateVarint(sb, data, offset, nestedFieldIndent+"Value length")
		annotateString(sb, data, offset, int(valueLen), nestedFieldIndent+"Value")
	}
}

func lookupHeaderInStaticTable(headerID byte, isRequest bool) (string, string, bool) {
	if isRequest {
		if entry, ok := requestHeaderStaticTable[headerID]; ok {
			// If entry.value is empty, the value bytes follow in the wire format (Format 2)
			// If entry.value is filled, it's a complete pair (Format 1)
			return entry.name, entry.value, entry.value == ""
		}
	} else {
		if entry, ok := responseHeaderStaticTable[headerID]; ok {
			return entry.name, entry.value, entry.value == ""
		}
	}
	return "", "", false
}
