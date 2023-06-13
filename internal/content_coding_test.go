package internal

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const maxDecompressionSize = 1024

func TestGzipEncodeDecode(t *testing.T) {
	enc, err := NewGzipEncoder()
	require.NoError(t, err)
	dec := NewGzipDecoder()

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload, maxDecompressionSize)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestGzipReuse(t *testing.T) {
	enc, err := NewGzipEncoder()
	require.NoError(t, err)
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
	enc, err := NewZlibEncoder()
	require.NoError(t, err)
	dec := NewZlibDecoder()

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload, maxDecompressionSize)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestZlibEncodeDecodeWithTooLargeMessage(t *testing.T) {
	enc, err := NewZlibEncoder()
	require.NoError(t, err)
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
	enc, err := NewGzipEncoder()
	require.NoError(t, err)
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

func TestCompressionLevel(t *testing.T) {
	tests := []struct {
		algorithm   string
		compression string
		errormsg    string
	}{
		{
			algorithm:   "gzip",
			compression: "default",
		},
		{
			algorithm:   "gzip",
			compression: "none",
		},
		{
			algorithm:   "gzip",
			compression: "best compression",
		},
		{
			algorithm:   "gzip",
			compression: "best speed",
		},
		{
			algorithm:   "gzip",
			compression: "invalid",
			errormsg:    "invalid compression level",
		},
		{
			algorithm:   "zlib",
			compression: "default",
		},
		{
			algorithm:   "zlib",
			compression: "none",
		},
		{
			algorithm:   "zlib",
			compression: "best compression",
		},
		{
			algorithm:   "zlib",
			compression: "best speed",
		},
		{
			algorithm:   "zlib",
			compression: "invalid",
			errormsg:    "invalid compression level",
		},
		{
			algorithm:   "identity",
			compression: "default",
		},
		{
			algorithm:   "identity",
			compression: "none",
		},
		{
			algorithm:   "identity",
			compression: "best compression",
		},
		{
			algorithm:   "identity",
			compression: "best speed",
		},
		{
			algorithm:   "identity",
			compression: "invalid",
			errormsg:    "invalid compression level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.algorithm+" "+tt.compression, func(t *testing.T) {
			level, err := ToCompressionLevel(tt.compression)
			if tt.errormsg != "" {
				require.ErrorContains(t, err, tt.errormsg)
				return
			}
			require.NoError(t, err)

			enc, err := NewContentEncoder(tt.algorithm, WithCompressionLevel(level))
			require.NoError(t, err)
			require.NotNil(t, enc)
		})
	}
}

func BenchmarkGzipEncode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder()
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err := enc.Encode(data)
		require.NoError(b, err)
	}
}

func BenchmarkGzipDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder()
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err = dec.Decode(payload, dataLen)
		require.NoError(b, err)
	}
}

func BenchmarkGzipEncodeDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder()
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload, dataLen)
		require.NoError(b, err)
	}
}

func BenchmarkGzipEncodeBig(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 1024*1024))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder()
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err := enc.Encode(data)
		require.NoError(b, err)
	}
}

func BenchmarkGzipDecodeBig(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 1024*1024))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder()
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err = dec.Decode(payload, dataLen)
		require.NoError(b, err)
	}
}

func BenchmarkGzipEncodeDecodeBig(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 1024*1024))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder()
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload, dataLen)
		require.NoError(b, err)
	}
}

func BenchmarkZlibEncode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewZlibEncoder()
	require.NoError(b, err)
	dec := NewZlibDecoder()
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err := enc.Encode(data)
		require.NoError(b, err)
	}
}

func BenchmarkZlibDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewZlibEncoder()
	require.NoError(b, err)
	dec := NewZlibDecoder()
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err = dec.Decode(payload, dataLen)
		require.NoError(b, err)
	}
}

func BenchmarkZlibEncodeDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewZlibEncoder()
	require.NoError(b, err)
	dec := NewZlibDecoder()
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload, dataLen)
		require.NoError(b, err)
	}
}

func BenchmarkIdentityEncodeDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc := NewIdentityEncoder()
	dec := NewIdentityDecoder()

	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload, dataLen)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload, dataLen)
		require.NoError(b, err)
	}
}
