package internal

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

const maxDecompressionSize = 1024

func TestGzipEncodeDecode(t *testing.T) {
	enc := NewGzipEncoder()
	dec := NewGzipDecoder()

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload, maxDecompressionSize)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestGzipReuse(t *testing.T) {
	enc := NewGzipEncoder()
	dec := NewGzipDecoder()

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload, maxDecompressionSize)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))

	payload, err = enc.Encode([]byte("doody"))
	require.NoError(t, err)

	actual, err = dec.Decode(payload, maxDecompressionSize)
	require.NoError(t, err)

	require.Equal(t, "doody", string(actual))
}

func TestZlibEncodeDecode(t *testing.T) {
	enc := NewZlibEncoder()
	dec := NewZlibDecoder()

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload, maxDecompressionSize)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestZlibEncodeDecodeWithTooLargeMessage(t *testing.T) {
	enc := NewZlibEncoder()
	dec := NewZlibDecoder()

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	_, err = dec.Decode(payload, 3)
	require.ErrorContains(t, err, "size of decoded data exceeds allowed size 3")
}

func TestIdentityEncodeDecode(t *testing.T) {
	enc := NewIdentityEncoder()
	dec := NewIdentityDecoder()

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload, maxDecompressionSize)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestStreamIdentityDecode(t *testing.T) {
	var r bytes.Buffer
	n, err := r.Write([]byte("howdy"))
	require.NoError(t, err)
	require.Equal(t, 5, n)

	dec, err := NewStreamContentDecoder("identity", &r)
	require.NoError(t, err)

	data, err := io.ReadAll(dec)
	require.NoError(t, err)

	require.Equal(t, []byte("howdy"), data)
}

func TestStreamGzipDecode(t *testing.T) {
	enc := NewGzipEncoder()
	written, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	w := bytes.NewBuffer(written)

	dec, err := NewStreamContentDecoder("gzip", w)
	require.NoError(t, err)

	b := make([]byte, 10)
	n, err := dec.Read(b)
	require.NoError(t, err)
	require.Equal(t, 5, n)

	require.Equal(t, []byte("howdy"), b[:n])
}
