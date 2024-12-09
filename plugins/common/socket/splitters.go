package socket

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

type lengthFieldSpec struct {
	Offset       int64  `toml:"offset"`
	Bytes        int64  `toml:"bytes"`
	Endianness   string `toml:"endianness"`
	HeaderLength int64  `toml:"header_length"`
	converter    func([]byte) int
}

type SplitConfig struct {
	SplittingStrategy    string          `toml:"splitting_strategy"`
	SplittingDelimiter   string          `toml:"splitting_delimiter"`
	SplittingLength      int             `toml:"splitting_length"`
	SplittingLengthField lengthFieldSpec `toml:"splitting_length_field"`
}

func (cfg *SplitConfig) NewSplitter() (bufio.SplitFunc, error) {
	switch cfg.SplittingStrategy {
	case "", "newline":
		return bufio.ScanLines, nil
	case "null":
		return scanNull, nil
	case "delimiter":
		re := regexp.MustCompile(`(\s*0?x)`)
		d := re.ReplaceAllString(strings.ToLower(cfg.SplittingDelimiter), "")
		delimiter, err := hex.DecodeString(d)
		if err != nil {
			return nil, fmt.Errorf("decoding delimiter failed: %w", err)
		}
		return createScanDelimiter(delimiter), nil
	case "fixed length":
		return createScanFixedLength(cfg.SplittingLength), nil
	case "variable length":
		// Create the converter function
		var order binary.ByteOrder
		switch strings.ToLower(cfg.SplittingLengthField.Endianness) {
		case "", "be":
			order = binary.BigEndian
		case "le":
			order = binary.LittleEndian
		default:
			return nil, fmt.Errorf("invalid 'endianness' %q", cfg.SplittingLengthField.Endianness)
		}

		switch cfg.SplittingLengthField.Bytes {
		case 1:
			cfg.SplittingLengthField.converter = func(b []byte) int {
				return int(b[0])
			}
		case 2:
			cfg.SplittingLengthField.converter = func(b []byte) int {
				return int(order.Uint16(b))
			}
		case 4:
			cfg.SplittingLengthField.converter = func(b []byte) int {
				return int(order.Uint32(b))
			}
		case 8:
			cfg.SplittingLengthField.converter = func(b []byte) int {
				return int(order.Uint64(b))
			}
		default:
			cfg.SplittingLengthField.converter = func(b []byte) int {
				buf := make([]byte, 8)
				start := 0
				if order == binary.BigEndian {
					start = 8 - len(b)
				}
				for i := 0; i < len(b); i++ {
					buf[start+i] = b[i]
				}
				return int(order.Uint64(buf))
			}
		}

		// Check if we have enough bytes in the header
		return createScanVariableLength(cfg.SplittingLengthField), nil
	}

	return nil, fmt.Errorf("unknown 'splitting_strategy' %q", cfg.SplittingStrategy)
}

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
