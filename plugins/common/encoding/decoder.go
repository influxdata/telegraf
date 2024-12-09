package encoding

import (
	"errors"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

// NewDecoder returns an x/text Decoder for the specified text encoding.  The
// Decoder converts a character encoding into utf-8 bytes.  If a BOM is found
// it will be converted into an utf-8 BOM, you can use
// github.com/dimchansky/utfbom to strip the BOM.
//
// The "none" or "" encoding will pass through bytes unchecked.  Use the utf-8
// encoding if you want invalid bytes replaced using the unicode
// replacement character.
//
// Detection of utf-16 endianness using the BOM is not currently provided due
// to the tail input plugins requirement to be able to start at the middle or
// end of the file.
func NewDecoder(enc string) (*Decoder, error) {
	switch enc {
	case "utf-8":
		return createDecoder(unicode.UTF8.NewDecoder()), nil
	case "utf-16le":
		return createDecoder(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()), nil
	case "utf-16be":
		return createDecoder(unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()), nil
	case "none", "":
		return createDecoder(encoding.Nop.NewDecoder()), nil
	}
	return nil, errors.New("unknown character encoding")
}
