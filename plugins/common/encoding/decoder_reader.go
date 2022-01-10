package encoding

import (
	"errors"
	"io"

	"golang.org/x/text/transform"
)

// Other than resetting r.err and r.transformComplete in Read() this
// was copied from x/text

func newDecoder(t transform.Transformer) *Decoder {
	return &Decoder{Transformer: t}
}

// A Decoder converts bytes to UTF-8. It implements transform.Transformer.
//
// Transforming source bytes that are not of that encoding will not result in an
// error per se. Each byte that cannot be transcoded will be represented in the
// output by the UTF-8 encoding of '\uFFFD', the replacement rune.
type Decoder struct {
	transform.Transformer

	// This forces external creators of Decoders to use names in struct
	// initializers, allowing for future extensibility without having to break
	// code.
	_ struct{}
}

// Bytes converts the given encoded bytes to UTF-8. It returns the converted
// bytes or nil, err if any error occurred.
func (d *Decoder) Bytes(b []byte) ([]byte, error) {
	b, _, err := transform.Bytes(d, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// String converts the given encoded string to UTF-8. It returns the converted
// string or "", err if any error occurred.
func (d *Decoder) String(s string) (string, error) {
	s, _, err := transform.String(d, s)
	if err != nil {
		return "", err
	}
	return s, nil
}

// Reader wraps another Reader to decode its bytes.
//
// The Decoder may not be used for any other operation as long as the returned
// Reader is in use.
func (d *Decoder) Reader(r io.Reader) io.Reader {
	return NewReader(r, d)
}

// Reader wraps another io.Reader by transforming the bytes read.
type Reader struct {
	r   io.Reader
	t   transform.Transformer
	err error

	// dst[dst0:dst1] contains bytes that have been transformed by t but
	// not yet copied out via Read.
	dst        []byte
	dst0, dst1 int

	// src[src0:src1] contains bytes that have been read from r but not
	// yet transformed through t.
	src        []byte
	src0, src1 int

	// transformComplete is whether the transformation is complete,
	// regardless of whether or not it was successful.
	transformComplete bool
}

var (
	// ErrShortDst means that the destination buffer was too short to
	// receive all of the transformed bytes.
	ErrShortDst = errors.New("transform: short destination buffer")

	// ErrShortSrc means that the source buffer has insufficient data to
	// complete the transformation.
	ErrShortSrc = errors.New("transform: short source buffer")

	// errInconsistentByteCount means that Transform returned success (nil
	// error) but also returned nSrc inconsistent with the src argument.
	errInconsistentByteCount = errors.New("transform: inconsistent byte count returned")
)

const defaultBufSize = 4096

// NewReader returns a new Reader that wraps r by transforming the bytes read
// via t. It calls Reset on t.
func NewReader(r io.Reader, t transform.Transformer) *Reader {
	t.Reset()
	return &Reader{
		r:   r,
		t:   t,
		dst: make([]byte, defaultBufSize),
		src: make([]byte, defaultBufSize),
	}
}

// Read implements the io.Reader interface.
func (r *Reader) Read(p []byte) (int, error) {
	// Clear previous errors so a Read can be performed even if the last call
	// returned EOF.
	r.err = nil
	r.transformComplete = false

	n := 0
	for {
		// Copy out any transformed bytes and return the final error if we are done.
		if r.dst0 != r.dst1 {
			n = copy(p, r.dst[r.dst0:r.dst1])
			r.dst0 += n
			if r.dst0 == r.dst1 && r.transformComplete {
				return n, r.err
			}
			return n, nil
		} else if r.transformComplete {
			return 0, r.err
		}

		// Try to transform some source bytes, or to flush the transformer if we
		// are out of source bytes. We do this even if r.r.Read returned an error.
		// As the io.Reader documentation says, "process the n > 0 bytes returned
		// before considering the error".
		if r.src0 != r.src1 || r.err != nil {
			var err error
			r.dst0 = 0
			r.dst1, n, err = r.t.Transform(r.dst, r.src[r.src0:r.src1], r.err == io.EOF)
			r.src0 += n

			switch {
			case err == nil:
				if r.src0 != r.src1 {
					r.err = errInconsistentByteCount
				}
				// The Transform call was successful; we are complete if we
				// cannot read more bytes into src.
				r.transformComplete = r.err != nil
				continue
			case err == ErrShortDst && (r.dst1 != 0 || n != 0):
				// Make room in dst by copying out, and try again.
				continue
			case err == ErrShortSrc && r.src1-r.src0 != len(r.src) && r.err == nil:
				// Read more bytes into src via the code below, and try again.
			default:
				r.transformComplete = true
				// The reader error (r.err) takes precedence over the
				// transformer error (err) unless r.err is nil or io.EOF.
				if r.err == nil || r.err == io.EOF {
					r.err = err
				}
				continue
			}
		}

		// Move any untransformed source bytes to the start of the buffer
		// and read more bytes.
		if r.src0 != 0 {
			r.src0, r.src1 = 0, copy(r.src, r.src[r.src0:r.src1])
		}
		n, r.err = r.r.Read(r.src[r.src1:])
		r.src1 += n
	}
}
