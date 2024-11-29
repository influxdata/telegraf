package internal

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/pgzip"
)

const defaultMaxDecompressionSize int64 = 500 * 1024 * 1024 // 500MB

// DecodingOption provide methods to change the decoding from the standard configuration.
type DecodingOption func(*decoderConfig)

type decoderConfig struct {
	maxDecompressionSize int64
}

func WithMaxDecompressionSize(maxDecompressionSize int64) DecodingOption {
	return func(cfg *decoderConfig) {
		cfg.maxDecompressionSize = maxDecompressionSize
	}
}

type encoderConfig struct {
	level int
}

// EncodingOption provide methods to change the encoding from the standard configuration.
type EncodingOption func(*encoderConfig)

func WithCompressionLevel(level int) EncodingOption {
	return func(cfg *encoderConfig) {
		cfg.level = level
	}
}

// NewStreamContentDecoder returns a reader that will decode the stream according to the encoding type.
func NewStreamContentDecoder(encoding string, r io.Reader) (io.Reader, error) {
	switch encoding {
	case "gzip":
		return newGzipReader(r)
	case "identity", "":
		return r, nil
	default:
		return nil, errors.New("invalid value for content_encoding")
	}
}

// gzipReader is similar to gzip.Reader but reads only a single gzip stream per read.
type gzipReader struct {
	r           io.Reader
	z           *pgzip.Reader
	endOfStream bool
}

func newGzipReader(r io.Reader) (io.Reader, error) {
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

	return &gzipReader{r: br, z: z}, nil
}

func (r *gzipReader) Read(b []byte) (int, error) {
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
		return newGzipEncoder(options...)
	case "identity", "":
		return newIdentityEncoder(options...)
	case "zlib":
		return newZlibEncoder(options...)
	case "zstd":
		return newZstdEncoder(options...)
	default:
		return nil, errors.New("invalid value for content_encoding")
	}
}

type autoDecoder struct {
	encoding string
	gzip     *gzipDecoder
	identity *identityDecoder
}

func (a *autoDecoder) SetEncoding(encoding string) {
	a.encoding = encoding
}

func (a *autoDecoder) Decode(data []byte) ([]byte, error) {
	if a.encoding == "gzip" {
		return a.gzip.Decode(data)
	}
	return a.identity.Decode(data)
}

func newAutoContentDecoder(options ...DecodingOption) *autoDecoder {
	var a autoDecoder

	a.identity = newIdentityDecoder(options...)
	a.gzip = newGzipDecoder(options...)
	return &a
}

// NewContentDecoder returns a ContentDecoder for the encoding type.
func NewContentDecoder(encoding string, options ...DecodingOption) (ContentDecoder, error) {
	switch encoding {
	case "auto":
		return newAutoContentDecoder(options...), nil
	case "gzip":
		return newGzipDecoder(options...), nil
	case "identity", "":
		return newIdentityDecoder(options...), nil
	case "zlib":
		return newZlibDecoder(options...), nil
	case "zstd":
		return newZstdDecoder(options...)
	default:
		return nil, errors.New("invalid value for content_encoding")
	}
}

// ContentEncoder applies a wrapper encoding to byte buffers.
type ContentEncoder interface {
	Encode([]byte) ([]byte, error)
}

// gzipEncoder compresses the buffer using gzip at the default level.
type gzipEncoder struct {
	pwriter *pgzip.Writer
	writer  *gzip.Writer
	buf     *bytes.Buffer
}

func newGzipEncoder(options ...EncodingOption) (*gzipEncoder, error) {
	cfg := encoderConfig{level: gzip.DefaultCompression}
	for _, o := range options {
		o(&cfg)
	}

	// Check if the compression level is supported
	switch cfg.level {
	case gzip.NoCompression, gzip.DefaultCompression, gzip.BestSpeed, gzip.BestCompression:
		// Do nothing as those are valid levels
	default:
		return nil, errors.New("invalid compression level, only 0, 1 and 9 are supported")
	}

	var buf bytes.Buffer
	pw, err := pgzip.NewWriterLevel(&buf, cfg.level)
	if err != nil {
		return nil, err
	}

	w, err := gzip.NewWriterLevel(&buf, cfg.level)
	return &gzipEncoder{
		pwriter: pw,
		writer:  w,
		buf:     &buf,
	}, err
}

func (e *gzipEncoder) Encode(data []byte) ([]byte, error) {
	// Parallel Gzip is only faster for larger data chunks. According to the
	// project's documentation the trade-off size is at about 1MB, so we switch
	// to parallel Gzip if the data is larger and run the built-in version
	// otherwise.
	if len(data) > 1024*1024 {
		return e.encodeBig(data)
	}
	return e.encodeSmall(data)
}

func (e *gzipEncoder) encodeSmall(data []byte) ([]byte, error) {
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

func (e *gzipEncoder) encodeBig(data []byte) ([]byte, error) {
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

type zlibEncoder struct {
	writer *zlib.Writer
	buf    *bytes.Buffer
}

func newZlibEncoder(options ...EncodingOption) (*zlibEncoder, error) {
	cfg := encoderConfig{level: zlib.DefaultCompression}
	for _, o := range options {
		o(&cfg)
	}

	switch cfg.level {
	case zlib.NoCompression, zlib.DefaultCompression, zlib.BestSpeed, zlib.BestCompression:
		// Do nothing as those are valid levels
	default:
		return nil, errors.New("invalid compression level, only 0, 1 and 9 are supported")
	}

	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, cfg.level)
	return &zlibEncoder{
		writer: w,
		buf:    &buf,
	}, err
}

func (e *zlibEncoder) Encode(data []byte) ([]byte, error) {
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

type zstdEncoder struct {
	encoder *zstd.Encoder
}

func newZstdEncoder(options ...EncodingOption) (*zstdEncoder, error) {
	cfg := encoderConfig{level: 3}
	for _, o := range options {
		o(&cfg)
	}

	// Map the levels
	var level zstd.EncoderLevel
	switch cfg.level {
	case 1:
		level = zstd.SpeedFastest
	case 3:
		level = zstd.SpeedDefault
	case 7:
		level = zstd.SpeedBetterCompression
	case 11:
		level = zstd.SpeedBestCompression
	default:
		return nil, errors.New("invalid compression level, only 1, 3, 7 and 11 are supported")
	}

	e, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(level))
	return &zstdEncoder{
		encoder: e,
	}, err
}

func (e *zstdEncoder) Encode(data []byte) ([]byte, error) {
	return e.encoder.EncodeAll(data, make([]byte, 0, len(data))), nil
}

// identityEncoder is a null encoder that applies no transformation.
type identityEncoder struct{}

func newIdentityEncoder(options ...EncodingOption) (*identityEncoder, error) {
	if len(options) > 0 {
		return nil, errors.New("identity encoder does not support options")
	}

	return &identityEncoder{}, nil
}

func (*identityEncoder) Encode(data []byte) ([]byte, error) {
	return data, nil
}

// ContentDecoder removes a wrapper encoding from byte buffers.
type ContentDecoder interface {
	SetEncoding(string)
	Decode([]byte) ([]byte, error)
}

// gzipDecoder decompresses buffers with gzip compression.
type gzipDecoder struct {
	preader              *pgzip.Reader
	reader               *gzip.Reader
	buf                  *bytes.Buffer
	maxDecompressionSize int64
}

func newGzipDecoder(options ...DecodingOption) *gzipDecoder {
	cfg := decoderConfig{maxDecompressionSize: defaultMaxDecompressionSize}
	for _, o := range options {
		o(&cfg)
	}

	return &gzipDecoder{
		preader:              new(pgzip.Reader),
		reader:               new(gzip.Reader),
		buf:                  new(bytes.Buffer),
		maxDecompressionSize: cfg.maxDecompressionSize,
	}
}

func (*gzipDecoder) SetEncoding(string) {}

func (d *gzipDecoder) Decode(data []byte) ([]byte, error) {
	// Parallel Gzip is only faster for larger data chunks. According to the
	// project's documentation the trade-off size is at about 1MB, so we switch
	// to parallel Gzip if the data is larger and run the built-in version
	// otherwise.
	if len(data) > 1024*1024 {
		return d.decodeBig(data)
	}
	return d.decodeSmall(data)
}

func (d *gzipDecoder) decodeSmall(data []byte) ([]byte, error) {
	err := d.reader.Reset(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	d.buf.Reset()

	n, err := io.CopyN(d.buf, d.reader, d.maxDecompressionSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	} else if n == d.maxDecompressionSize {
		return nil, fmt.Errorf("size of decoded data exceeds allowed size %d", d.maxDecompressionSize)
	}

	err = d.reader.Close()
	if err != nil {
		return nil, err
	}
	return d.buf.Bytes(), nil
}

func (d *gzipDecoder) decodeBig(data []byte) ([]byte, error) {
	err := d.preader.Reset(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	d.buf.Reset()

	n, err := io.CopyN(d.buf, d.preader, d.maxDecompressionSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	} else if n == d.maxDecompressionSize {
		return nil, fmt.Errorf("size of decoded data exceeds allowed size %d", d.maxDecompressionSize)
	}

	err = d.preader.Close()
	if err != nil {
		return nil, err
	}
	return d.buf.Bytes(), nil
}

type zlibDecoder struct {
	buf                  *bytes.Buffer
	maxDecompressionSize int64
}

func newZlibDecoder(options ...DecodingOption) *zlibDecoder {
	cfg := decoderConfig{maxDecompressionSize: defaultMaxDecompressionSize}
	for _, o := range options {
		o(&cfg)
	}

	return &zlibDecoder{
		buf:                  new(bytes.Buffer),
		maxDecompressionSize: cfg.maxDecompressionSize,
	}
}

func (*zlibDecoder) SetEncoding(string) {}

func (d *zlibDecoder) Decode(data []byte) ([]byte, error) {
	d.buf.Reset()

	b := bytes.NewBuffer(data)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}

	n, err := io.CopyN(d.buf, r, d.maxDecompressionSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	} else if n == d.maxDecompressionSize {
		return nil, fmt.Errorf("size of decoded data exceeds allowed size %d", d.maxDecompressionSize)
	}

	err = r.Close()
	if err != nil {
		return nil, err
	}
	return d.buf.Bytes(), nil
}

type zstdDecoder struct {
	decoder *zstd.Decoder
}

func newZstdDecoder(options ...DecodingOption) (*zstdDecoder, error) {
	cfg := decoderConfig{maxDecompressionSize: defaultMaxDecompressionSize}
	for _, o := range options {
		o(&cfg)
	}

	d, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(0), zstd.WithDecoderMaxWindow(uint64(cfg.maxDecompressionSize)))
	return &zstdDecoder{
		decoder: d,
	}, err
}

func (*zstdDecoder) SetEncoding(string) {}

func (d *zstdDecoder) Decode(data []byte) ([]byte, error) {
	return d.decoder.DecodeAll(data, nil)
}

// identityDecoder is a null decoder that returns the input.
type identityDecoder struct {
}

func newIdentityDecoder(_ ...DecodingOption) *identityDecoder {
	return &identityDecoder{}
}

func (*identityDecoder) SetEncoding(string) {}

func (*identityDecoder) Decode(data []byte) ([]byte, error) {
	return data, nil
}
