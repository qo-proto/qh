package qh

import (
	"fmt"
	"strings"
)

func (r *Request) AnnotateWireFormat(data []byte) string {
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
	remaining := len(data) - offset
	if remaining < 0 {
		remaining = 0
	}
	if hostLen > uint64(remaining) {
		hostLen = uint64(remaining)
	}
	annotateString(&sb, data, &offset, int(hostLen), "Host") //nolint:gosec // bounds checked above

	pathLen := annotateVarint(&sb, data, &offset, "Path length")
	remaining = len(data) - offset
	if remaining < 0 {
		remaining = 0
	}
	if pathLen > uint64(remaining) {
		pathLen = uint64(remaining)
	}
	annotateString(&sb, data, &offset, int(pathLen), "Path") //nolint:gosec // bounds checked above

	headersLen := annotateVarint(&sb, data, &offset, "Headers length")
	remaining = len(data) - offset
	if remaining < 0 {
		remaining = 0
	}
	if headersLen > uint64(remaining) {
		headersLen = uint64(remaining)
	}
	headersEndOffset := offset + int(headersLen) //nolint:gosec // bounds checked above
	annotateHeaders(&sb, data, &offset, headersEndOffset, true)
	offset = headersEndOffset

	bodyLen := annotateVarint(&sb, data, &offset, "Body length")
	if bodyLen > 0 && offset+int(bodyLen) <= len(data) {
		bodyPreview := string(data[offset : offset+int(bodyLen)])
		if len(bodyPreview) > 50 {
			bodyPreview = bodyPreview[:50] + "..."
		}
		sb.WriteString(fmt.Sprintf("    (body data)                  Body: %s\n", bodyPreview))
		offset += int(bodyLen)
	}

	fmt.Fprintf(&sb, "    (parsed %d / %d bytes)\n", offset, len(data))

	return sb.String()
}

func (r *Response) AnnotateWireFormat(data []byte) string {
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
	headersEndOffset := offset + int(headersLen)
	annotateHeaders(&sb, data, &offset, headersEndOffset, false)
	offset = headersEndOffset

	bodyLen := annotateVarint(&sb, data, &offset, "Body length")
	if bodyLen > 0 && offset+int(bodyLen) <= len(data) {
		bodyPreview := string(data[offset : offset+int(bodyLen)])
		if len(bodyPreview) > 50 {
			bodyPreview = bodyPreview[:50] + "..."
		}
		sb.WriteString(fmt.Sprintf("    (body data)                  Body: %s\n", bodyPreview))
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
	for i := n * 5; i < 28; i++ {
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

	sb.WriteString("    ")
	writeHex(sb, data[*offset:*offset+length])
	for i := length * 5; i < 28; i++ {
		sb.WriteByte(' ')
	}
	fmt.Fprintf(sb, " %s: %s\n", label, value)

	*offset += length
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
