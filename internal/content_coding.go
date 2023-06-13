package internal

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/pgzip"
)

const DefaultMaxDecompressionSize = 500 * 1024 * 1024 //500MB

type CompressionLevel int

const (
	None CompressionLevel = iota
	Default
	BestSpeed
	BestCompression
)

func ToCompressionLevel(level string) (CompressionLevel, error) {
	switch level {
	case "", "default":
		return Default, nil
	case "none":
		return None, nil
	case "best speed":
		return BestSpeed, nil
	case "best compression":
		return BestCompression, nil
	}
	return -1, errors.New("invalid compression level")
}

type encoderConfig struct {
	level CompressionLevel
}

// EncodingOption provide methods to change the encoding from the standard
// configuration.
type EncodingOption func(*encoderConfig)

func WithCompressionLevel(level CompressionLevel) EncodingOption {
	return func(cfg *encoderConfig) {
		cfg.level = level
	}
}

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
	z           *pgzip.Reader
	endOfStream bool
}

func NewGzipReader(r io.Reader) (io.Reader, error) {
	// We need a read that implements ByteReader in order to line up the next
	// stream.
	br := bufio.NewReader(r)

	// Reads the first gzip stream header.
	z, err := pgzip.NewReader(br)
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
func NewContentEncoder(encoding string, options ...EncodingOption) (ContentEncoder, error) {
	switch encoding {
	case "gzip":
		return NewGzipEncoder(options...)
	case "zlib":
		return NewZlibEncoder(options...)
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

func (a *AutoDecoder) Decode(data []byte, maxDecompressionSize int64) ([]byte, error) {
	if a.encoding == "gzip" {
		return a.gzip.Decode(data, maxDecompressionSize)
	}
	return a.identity.Decode(data, maxDecompressionSize)
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
	pwriter *pgzip.Writer
	writer  *gzip.Writer
	buf     *bytes.Buffer
}

func NewGzipEncoder(options ...EncodingOption) (*GzipEncoder, error) {
	cfg := encoderConfig{level: Default}
	for _, o := range options {
		o(&cfg)
	}

	var plevel int
	switch cfg.level {
	case None:
		plevel = pgzip.NoCompression
	case Default:
		plevel = pgzip.DefaultCompression
	case BestSpeed:
		plevel = pgzip.BestSpeed
	case BestCompression:
		plevel = pgzip.BestCompression
	default:
		return nil, fmt.Errorf("invalid compression level %d", cfg.level)
	}
	var buf bytes.Buffer
	pw, err := pgzip.NewWriterLevel(&buf, plevel)
	if err != nil {
		return nil, err
	}

	var level int
	switch cfg.level {
	case None:
		level = gzip.NoCompression
	case Default:
		level = gzip.DefaultCompression
	case BestSpeed:
		level = gzip.BestSpeed
	case BestCompression:
		level = gzip.BestCompression
	default:
		return nil, fmt.Errorf("invalid compression level %d", cfg.level)
	}
	w, err := gzip.NewWriterLevel(&buf, level)
	return &GzipEncoder{
		pwriter: pw,
		writer:  w,
		buf:     &buf,
	}, err
}

func (e *GzipEncoder) Encode(data []byte) ([]byte, error) {
	// Parallel Gzip is only faster for larger data chunks. According to the
	// project's documentation the trade-off size is at about 1MB, so we switch
	// to parallel Gzip if the data is larger and run the built-in version
	// otherwise.
	if len(data) > 1024*1024 {
		return e.encodeBig(data)
	}
	return e.encodeSmall(data)
}

func (e *GzipEncoder) encodeSmall(data []byte) ([]byte, error) {
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

func (e *GzipEncoder) encodeBig(data []byte) ([]byte, error) {
	e.buf.Reset()
	e.pwriter.Reset(e.buf)

	_, err := e.pwriter.Write(data)
	if err != nil {
		return nil, err
	}
	err = e.pwriter.Close()
	if err != nil {
		return nil, err
	}
	return e.buf.Bytes(), nil
}

type ZlibEncoder struct {
	writer *zlib.Writer
	buf    *bytes.Buffer
}

func NewZlibEncoder(options ...EncodingOption) (*ZlibEncoder, error) {
	cfg := encoderConfig{level: Default}
	for _, o := range options {
		o(&cfg)
	}

	var level int
	switch cfg.level {
	case None:
		level = zlib.NoCompression
	case Default:
		level = zlib.DefaultCompression
	case BestSpeed:
		level = zlib.BestSpeed
	case BestCompression:
		level = zlib.BestCompression
	default:
		return nil, fmt.Errorf("invalid compression level %d", cfg.level)
	}

	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, level)
	return &ZlibEncoder{
		writer: w,
		buf:    &buf,
	}, err
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
	Decode([]byte, int64) ([]byte, error)
}

// GzipDecoder decompresses buffers with gzip compression.
type GzipDecoder struct {
	preader *pgzip.Reader
	reader  *gzip.Reader
	buf     *bytes.Buffer
}

func NewGzipDecoder() *GzipDecoder {
	return &GzipDecoder{
		preader: new(pgzip.Reader),
		reader:  new(gzip.Reader),
		buf:     new(bytes.Buffer),
	}
}

func (*GzipDecoder) SetEncoding(string) {}

func (d *GzipDecoder) Decode(data []byte, maxDecompressionSize int64) ([]byte, error) {
	// Parallel Gzip is only faster for larger data chunks. According to the
	// project's documentation the trade-off size is at about 1MB, so we switch
	// to parallel Gzip if the data is larger and run the built-in version
	// otherwise.
	if len(data) > 1024*1024 {
		return d.decodeBig(data, maxDecompressionSize)
	}
	return d.decodeSmall(data, maxDecompressionSize)
}

func (d *GzipDecoder) decodeSmall(data []byte, maxDecompressionSize int64) ([]byte, error) {
	err := d.reader.Reset(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	d.buf.Reset()

	n, err := io.CopyN(d.buf, d.reader, maxDecompressionSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	} else if n == maxDecompressionSize {
		return nil, fmt.Errorf("size of decoded data exceeds allowed size %d", maxDecompressionSize)
	}

	err = d.reader.Close()
	if err != nil {
		return nil, err
	}
	return d.buf.Bytes(), nil
}

func (d *GzipDecoder) decodeBig(data []byte, maxDecompressionSize int64) ([]byte, error) {
	err := d.preader.Reset(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	d.buf.Reset()

	n, err := io.CopyN(d.buf, d.preader, maxDecompressionSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	} else if n == maxDecompressionSize {
		return nil, fmt.Errorf("size of decoded data exceeds allowed size %d", maxDecompressionSize)
	}

	err = d.preader.Close()
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

func (d *ZlibDecoder) Decode(data []byte, maxDecompressionSize int64) ([]byte, error) {
	d.buf.Reset()

	b := bytes.NewBuffer(data)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}

	n, err := io.CopyN(d.buf, r, maxDecompressionSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	} else if n == maxDecompressionSize {
		return nil, fmt.Errorf("size of decoded data exceeds allowed size %d", maxDecompressionSize)
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

func (*IdentityDecoder) Decode(data []byte, maxDecompressionSize int64) ([]byte, error) {
	size := int64(len(data))
	if size > maxDecompressionSize {
		return nil, fmt.Errorf("size of decoded data: %d exceeds allowed size %d", size, maxDecompressionSize)
	}
	return data, nil
}
