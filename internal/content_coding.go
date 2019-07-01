package internal

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"

	"github.com/golang/snappy"
)

// NewContentEncoder returns a ContentEncoder for the encoding type.
func NewContentEncoder(encoding string) (ContentEncoder, error) {
	switch encoding {
	case "gzip":
		return NewGzipEncoder()
	case "snappy":
		return NewSnappyEncoder()
	case "identity", "":
		return NewIdentityEncoder(), nil
	default:
		return nil, errors.New("invalid value for content_encoding")
	}
}

// NewContentDecoder returns a ContentDecoder for the encoding type.
func NewContentDecoder(encoding string) (ContentDecoder, error) {
	switch encoding {
	case "gzip":
		return NewGzipDecoder()
	case "snappy":
		return NewSnappyDecoder()
	case "identity", "":
		return NewIdentityDecoder(), nil
	default:
		return nil, errors.New("invalid value for content_encoding")
	}
}

// ContentEncoder applies a wrapper encoding to byte buffers.
type ContentEncoder interface {
	Encode([]byte) ([]byte, error)
}

// GzipEncoder compresses the buffer using gzip at the default level.
type GzipEncoder struct {
	writer *gzip.Writer
	buf    *bytes.Buffer
}

func NewGzipEncoder() (*GzipEncoder, error) {
	var buf bytes.Buffer
	return &GzipEncoder{
		writer: gzip.NewWriter(&buf),
		buf:    &buf,
	}, nil
}

// SetLevel will change the gzip encoder compression level
// See https://golang.org/pkg/compress/gzip/#pkg-constants
// or a number between 0 and 9
// 0 being no compression
// 9 being best but slowest compression
// -1 is used to reset back to the default level
func (e *GzipEncoder) SetLevel(lvl int) error {
	gzw, err := gzip.NewWriterLevel(e.buf, lvl)
	if err != nil {
		return nil
	}
	e.writer = gzw
	return nil
}

// Encode will take data passed in and encode it with GZip
func (e *GzipEncoder) Encode(data []byte) ([]byte, error) {
	e.buf.Reset()
	e.writer.Reset(e.buf)

	_, err := e.writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = e.writer.Close()
	if err != nil {
		return nil, err
	}
	return e.buf.Bytes(), nil
}

// SnappyEncoder compresses and decompresses the buffer using google's snappy encryption.
type SnappyEncoder struct{}

// NewSnappyEncoder returns a new snappy encoder that can encode []bytes to
// google snappy []bytes.
func NewSnappyEncoder() (*SnappyEncoder, error) {
	return &SnappyEncoder{}, nil
}

// NewSnappyDecoder returns a new snappy dencoder that can dencode []bytes previously encoded to
// []bytes.
func NewSnappyDecoder() (*SnappyEncoder, error) {
	return &SnappyEncoder{}, nil
}

// Encode take all data given to it and encodes it.
// Snappy will never return an error other than nil but returns nil to
// satisfy the Encode interface here.
func (e *SnappyEncoder) Encode(data []byte) ([]byte, error) {
	return snappy.Encode(nil, data), nil
}

// Decode takes the passed in data and decodes it to a []byte.
// It can return an error if the data was encoded incorrectly.
func (e *SnappyEncoder) Decode(data []byte) ([]byte, error) {
	return snappy.Decode(nil, data)
}

// IdentityEncoder is a null encoder that applies no transformation.
type IdentityEncoder struct{}

func NewIdentityEncoder() *IdentityEncoder {
	return &IdentityEncoder{}
}

func (*IdentityEncoder) Encode(data []byte) ([]byte, error) {
	return data, nil
}

// ContentDecoder removes a wrapper encoding from byte buffers.
type ContentDecoder interface {
	Decode([]byte) ([]byte, error)
}

// GzipDecoder decompresses buffers with gzip compression.
type GzipDecoder struct {
	reader *gzip.Reader
	buf    *bytes.Buffer
}

func NewGzipDecoder() (*GzipDecoder, error) {
	return &GzipDecoder{
		reader: new(gzip.Reader),
		buf:    new(bytes.Buffer),
	}, nil
}

func (d *GzipDecoder) Decode(data []byte) ([]byte, error) {
	d.reader.Reset(bytes.NewBuffer(data))
	d.buf.Reset()

	_, err := d.buf.ReadFrom(d.reader)
	if err != nil && err != io.EOF {
		return nil, err
	}
	err = d.reader.Close()
	if err != nil {
		return nil, err
	}
	return d.buf.Bytes(), nil
}

// IdentityDecoder is a null decoder that returns the input.
type IdentityDecoder struct{}

func NewIdentityDecoder() *IdentityDecoder {
	return &IdentityDecoder{}
}

func (*IdentityDecoder) Decode(data []byte) ([]byte, error) {
	return data, nil
}
