package socket_listener

import (
	"bufio"
	"bytes"
)

func scanNull(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func createScanDelimiter(delimiter []byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, delimiter); i >= 0 {
			return i + len(delimiter), data[:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
}

func createScanFixedLength(length int) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if len(data) >= length {
			return length, data[:length], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
}

func createScanVariableLength(spec lengthFieldSpec) bufio.SplitFunc {
	minlen := int(spec.Offset)
	minlen += int(spec.Bytes)
	headerLen := int(spec.HeaderLength)

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		dataLen := len(data)
		if dataLen >= minlen {
			// Extract the length field and convert it to a number
			lf := data[spec.Offset : spec.Offset+spec.Bytes]
			length := spec.converter(lf)
			start := headerLen
			end := length + headerLen
			// If we have enough data return it without the header
			if end <= dataLen {
				return end, data[start:end], nil
			}
		}
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
}
