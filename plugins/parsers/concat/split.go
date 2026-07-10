package concat

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
)

type Splitter struct {
	ConstBytes         map[string]string `toml:"const_bytes"`
	HeaderLength       int               `toml:"header_length"`
	MessageLengthField LengthFieldConfig `toml:"message_length_field"`
	Checksum           ChecksumConfig    `toml:"checksum"`
	s                  bufio.SplitFunc

	// Resume cursor within the current message; see the scan loop in Init.
	scanPos int
}

type ChecksumConfig struct {
	Strategy string `toml:"strategy"`
	Range    []int  `toml:"range"`
	Offset   int    `toml:"offset"`
	Bytes    int    `toml:"bytes"`
}

type LengthFieldConfig struct {
	Offset     int    `toml:"offset"`
	Bytes      int    `toml:"bytes"`
	Endianness string `toml:"endianness"`
	MaxLength  int    `toml:"max_length"`
}

func (s *Splitter) Init() error {
	// Parse the constant bytes into an offset -> value map
	constBytes := make(map[int]byte, len(s.ConstBytes))
	for k, v := range s.ConstBytes {
		offset, err := strconv.Atoi(k)
		if err != nil {
			return fmt.Errorf("invalid const byte offset %q: %w", k, err)
		}
		if offset < 0 {
			return fmt.Errorf("const byte offset %d must not be negative", offset)
		}
		if s.HeaderLength > 0 && offset >= s.HeaderLength {
			return fmt.Errorf("const byte offset %d is past the header length %d", offset, s.HeaderLength)
		}
		// Decode the hex value, accepting an optional "0x"/"x" prefix.
		h := strings.ToLower(v)
		h = strings.TrimPrefix(h, "0x")
		h = strings.TrimPrefix(h, "x")
		b, err := hex.DecodeString(h)
		if err != nil {
			return fmt.Errorf("invalid const byte value %q at offset %d: %w", v, offset, err)
		}
		if len(b) != 1 {
			return fmt.Errorf("const byte value %q at offset %d is not a single byte", v, offset)
		}
		constBytes[offset] = b[0]
	}

	lengthConv, err := makeLengthConverter(s.MessageLengthField.Bytes, s.MessageLengthField.Endianness)
	if err != nil {
		return err
	}
	lf := s.MessageLengthField

	checksumFn, err := makeChecksum(s.Checksum.Strategy, s.Checksum.Bytes)
	if err != nil {
		return err
	}
	cs := s.Checksum
	if checksumFn != nil {
		if len(cs.Range) != 2 {
			return fmt.Errorf("checksum requires a range [start, end]")
		}
		if cs.Range[0] < 0 || cs.Range[1] < cs.Range[0] {
			return fmt.Errorf("invalid checksum range %v", cs.Range)
		}
	}

	// Without a length field, boundaries are found by scanning for the next
	// header, so we need something to recognize a header by.
	if lf.Bytes <= 0 && len(constBytes) == 0 && checksumFn == nil {
		return fmt.Errorf("cannot determine message boundaries: configure a length field, constant bytes, or a checksum")
	}

	// Bytes needed before a header can be validated at a given position.
	minlen := 0
	if lf.Bytes > 0 {
		minlen = max(minlen, lf.Offset+lf.Bytes)
	}
	for offset := range constBytes {
		minlen = max(minlen, offset+1)
	}
	if checksumFn != nil {
		minlen = max(minlen, cs.Offset+cs.Bytes)
		minlen = max(minlen, cs.Range[1])
	}

	headerLen := s.HeaderLength

	// headerAt reports whether a valid header starts at position p. The caller
	// must ensure at least p+minlen bytes are available.
	headerAt := func(buf []byte, p int) bool {
		for offset, b := range constBytes {
			if buf[p+offset] != b {
				return false
			}
		}
		if checksumFn != nil {
			sum := checksumFn(buf[p+cs.Range[0] : p+cs.Range[1]])
			if !bytes.Equal(sum, buf[p+cs.Offset:p+cs.Offset+cs.Bytes]) {
				return false
			}
		}
		return true
	}

	// scanPos is 0 at the start of a new message (header not yet validated) and
	// >0 once validated, when we are only looking for the end. It is reset to 0
	// on every message boundary.
	s.scanPos = 0

	if lf.Bytes > 0 {
		// Length mode: the length field gives the message end.
		s.s = func(data []byte, _ bool) (advance int, token []byte, err error) {
			dataLen := len(data)
			if dataLen == 0 || dataLen < minlen {
				return 0, nil, nil
			}

			// Validate the header only at the start of a new message; resync by
			// one byte on mismatch.
			if s.scanPos == 0 {
				if !headerAt(data, 0) {
					return 1, nil, nil
				}
				s.scanPos = 1
			}

			end := lengthConv(data[lf.Offset : lf.Offset+lf.Bytes])
			if end <= 0 || (lf.MaxLength > 0 && end > lf.MaxLength) {
				s.scanPos = 0
				return 1, nil, nil
			}
			// A valid header before the claimed end means the length is bogus
			// (desync or truncation); discard up to that header.
			p := s.scanPos
			for ; p < end && p+minlen <= dataLen; p++ {
				if headerAt(data, p) {
					s.scanPos = 0
					return p, nil, nil
				}
			}
			if end > dataLen {
				s.scanPos = p
				return 0, nil, nil
			}
			s.scanPos = 0
			if end > headerLen {
				return end, data[headerLen:end], nil
			}
			// The length field is too small to include any payload; discard it.
			return end, nil, nil
		}
		return nil
	}

	// Delimiter mode: the message ends at the next valid header.
	s.s = func(data []byte, _ bool) (advance int, token []byte, err error) {
		dataLen := len(data)
		if dataLen == 0 || dataLen < minlen {
			return 0, nil, nil
		}

		// Validate the header only at the start of a new message; resync by
		// one byte on mismatch.
		if s.scanPos == 0 {
			if !headerAt(data, 0) {
				return 1, nil, nil
			}
			s.scanPos = 1
		}

		// A header can only start after the current message's header, so never
		// look before headerLen. Resume from scanPos so each position is
		// examined once.
		p := s.scanPos
		if p < headerLen {
			p = headerLen
		}
		for ; p+minlen <= dataLen; p++ {
			if headerAt(data, p) {
				s.scanPos = 0
				return p, data[headerLen:p], nil
			}
		}
		s.scanPos = p
		return 0, nil, nil
	}

	return nil
}

// Split implements bufio.SplitFunc, returning one message with the header
// stripped. An (advance>0, nil) return means bytes were skipped to
// resynchronize without producing a message.
func (s *Splitter) Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if s.s == nil {
		return 0, nil, nil
	}
	return s.s(data, atEOF)
}

// makeLengthConverter returns a function decoding a big- or little-endian
// unsigned integer of the given width. It returns nil when no length field is
// configured (nbytes <= 0).
func makeLengthConverter(nbytes int, endianness string) (func([]byte) int, error) {
	if nbytes <= 0 {
		return nil, nil
	}
	var order binary.ByteOrder
	switch strings.ToLower(endianness) {
	case "", "be":
		order = binary.BigEndian
	case "le":
		order = binary.LittleEndian
	default:
		return nil, fmt.Errorf("invalid endianness %q", endianness)
	}
	switch nbytes {
	case 1:
		return func(b []byte) int { return int(b[0]) }, nil
	case 2:
		return func(b []byte) int { return int(order.Uint16(b)) }, nil
	case 4:
		return func(b []byte) int { return int(order.Uint32(b)) }, nil
	case 8:
		return func(b []byte) int { return int(order.Uint64(b)) }, nil
	default:
		return nil, fmt.Errorf("invalid length field width %d", nbytes)
	}
}

// makeChecksum returns a function computing the checksum of the given data
// using the requested strategy, producing a digest of width bytes. It returns
// nil when no checksum is configured.
func makeChecksum(strategy string, width int) (func([]byte) []byte, error) {
	switch strings.ToLower(strategy) {
	case "", "none":
		return nil, nil
	case "crc32":
		if width != 4 {
			return nil, fmt.Errorf("crc32 checksum requires 4 bytes, got %d", width)
		}
		return func(data []byte) []byte {
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, crc32.ChecksumIEEE(data))
			return b
		}, nil
	case "md5":
		if width != md5.Size {
			return nil, fmt.Errorf("md5 checksum requires %d bytes, got %d", md5.Size, width)
		}
		return func(data []byte) []byte {
			sum := md5.Sum(data)
			return sum[:]
		}, nil
	case "sha256":
		if width != sha256.Size {
			return nil, fmt.Errorf("sha256 checksum requires %d bytes, got %d", sha256.Size, width)
		}
		return func(data []byte) []byte {
			sum := sha256.Sum256(data)
			return sum[:]
		}, nil
	case "xor":
		if width <= 0 {
			return nil, fmt.Errorf("xor checksum requires a positive byte width")
		}
		return func(data []byte) []byte {
			sum := make([]byte, width)
			for i, b := range data {
				sum[i%width] ^= b
			}
			return sum
		}, nil
	default:
		return nil, fmt.Errorf("unknown checksum strategy %q", strategy)
	}
}
