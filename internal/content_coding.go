package internal

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
)

// NewContentEncoder returns a ContentEncoder for the encoding type.
func NewContentEncoder(encoding string) (ContentEncoder, error) {
	switch encoding {
	case "gzip":
		return NewGzipEncoder()

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
