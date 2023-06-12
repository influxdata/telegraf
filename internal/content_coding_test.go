package internal

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const maxDecompressionSize = 1024

func TestCompressionLevelValidationErrors(t *testing.T) {
	var errorTests = []struct {
		name      string
		algorithm string
		level     int
		expected  string
	}{
		{"wrong-algorithm", "asda", 1, "invalid value for content_encoding"},
		{"wrong-level", "gzip", 4, "unsupported compression level provided: 4. only [-2 -1 1 9] are supported"},
		{"wrong-level-and-algorithm", "asdas", 15, "invalid value for content_encoding"},
	}
	var successTests = []struct {
		name      string
		algorithm string
		level     int
	}{
		{"disabled", "", 0},
		{"default", "gzip", -1},
		{"enabled-0", "", 0},
		{"enabled-9", "gzip", 9},
		{"enabled-default", "zlib", -1},
		{"enabled-2", "gzip", -2},
		{"enabled-1", "gzip", 1},
	}
	for _, tt := range successTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewContentEncoder(tt.algorithm, EncoderCompressionLevel(tt.level))
			require.NoError(t, err)
		})
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewContentEncoder(tt.algorithm, EncoderCompressionLevel(tt.level))
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

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
