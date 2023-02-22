package internal

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"io"
)

// NewStreamContentDecoder returns a reader that will decode the stream
// according to the encoding type.
func NewStreamContentDecoder(encoding string, r io.Reader) (io.Reader, error) {
	switch encoding {
	case "gzip":
		return NewGzipReader(r)
	case "identity", "":
		return r, nil
	default:
		return nil, errors.New("invalid value for content_encoding")
	}
}

// GzipReader is similar to gzip.Reader but reads only a single gzip stream per read.
type GzipReader struct {
	r           io.Reader
	z           *gzip.Reader
	endOfStream bool
}

func NewGzipReader(r io.Reader) (io.Reader, error) {
	// We need a read that implements ByteReader in order to line up the next
	// stream.
	br := bufio.NewReader(r)

	// Reads the first gzip stream header.
	z, err := gzip.NewReader(br)
	if err != nil {
		return nil, err
	}

	// Prevent future calls to Read from reading the following gzip header.
	z.Multistream(false)

	return &GzipReader{r: br, z: z}, nil
}

func (r *GzipReader) Read(b []byte) (int, error) {
	if r.endOfStream {
		// Reads the next gzip header and prepares for the next stream.
		err := r.z.Reset(r.r)
		if err != nil {
			return 0, err
		}
		r.z.Multistream(false)
		r.endOfStream = false
	}

	n, err := r.z.Read(b)

	// Since multistream is disabled, io.EOF indicates the end of the gzip
	// sequence.  On the next read we must read the next gzip header.
	if errors.Is(err, io.EOF) {
		r.endOfStream = true
		return n, nil
	}
	return n, err
}

// NewContentEncoder returns a ContentEncoder for the encoding type.
func NewContentEncoder(encoding string) (ContentEncoder, error) {
	switch encoding {
	case "gzip":
		return NewGzipEncoder(), nil
	case "zlib":
		return NewZlibEncoder(), nil
	case "identity", "":
		return NewIdentityEncoder(), nil
	default:
		return nil, errors.New("invalid value for content_encoding")
	}
}

type AutoDecoder struct {
	encoding string
	gzip     *GzipDecoder
	identity *IdentityDecoder
}

func (a *AutoDecoder) SetEncoding(encoding string) {
	a.encoding = encoding
}

func (a *AutoDecoder) Decode(data []byte) ([]byte, error) {
	if a.encoding == "gzip" {
		return a.gzip.Decode(data)
	}
	return a.identity.Decode(data)
}

func NewAutoContentDecoder() *AutoDecoder {
	var a AutoDecoder

	a.identity = NewIdentityDecoder()
	a.gzip = NewGzipDecoder()
	return &a
}

// NewContentDecoder returns a ContentDecoder for the encoding type.
func NewContentDecoder(encoding string) (ContentDecoder, error) {
	switch encoding {
	case "gzip":
		return NewGzipDecoder(), nil
	case "zlib":
		return NewZlibDecoder(), nil
	case "identity", "":
		return NewIdentityDecoder(), nil
	case "auto":
		return NewAutoContentDecoder(), nil
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

func NewGzipEncoder() *GzipEncoder {
	var buf bytes.Buffer
	return &GzipEncoder{
		writer: gzip.NewWriter(&buf),
		buf:    &buf,
	}
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

type ZlibEncoder struct {
	writer *zlib.Writer
	buf    *bytes.Buffer
}

func NewZlibEncoder() *ZlibEncoder {
	var buf bytes.Buffer
	return &ZlibEncoder{
		writer: zlib.NewWriter(&buf),
		buf:    &buf,
	}
}

func (e *ZlibEncoder) Encode(data []byte) ([]byte, error) {
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
	SetEncoding(string)
	Decode([]byte) ([]byte, error)
}

// GzipDecoder decompresses buffers with gzip compression.
type GzipDecoder struct {
	reader *gzip.Reader
	buf    *bytes.Buffer
}

func NewGzipDecoder() *GzipDecoder {
	return &GzipDecoder{
		reader: new(gzip.Reader),
		buf:    new(bytes.Buffer),
	}
}

func (*GzipDecoder) SetEncoding(string) {}

func (d *GzipDecoder) Decode(data []byte) ([]byte, error) {
	err := d.reader.Reset(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	d.buf.Reset()

	_, err = d.buf.ReadFrom(d.reader)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	err = d.reader.Close()
	if err != nil {
		return nil, err
	}
	return d.buf.Bytes(), nil
}

type ZlibDecoder struct {
	buf *bytes.Buffer
}

func NewZlibDecoder() *ZlibDecoder {
	return &ZlibDecoder{
		buf: new(bytes.Buffer),
	}
}

func (*ZlibDecoder) SetEncoding(string) {}

func (d *ZlibDecoder) Decode(data []byte) ([]byte, error) {
	d.buf.Reset()

	b := bytes.NewBuffer(data)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(d.buf, r)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	err = r.Close()
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

func (*IdentityDecoder) SetEncoding(string) {}

func (*IdentityDecoder) Decode(data []byte) ([]byte, error) {
	return data, nil
}
