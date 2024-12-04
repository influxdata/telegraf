package internal

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const maxDecompressionSize = 1024

func TestGzipEncodeDecode(t *testing.T) {
	enc, err := NewGzipEncoder()
	require.NoError(t, err)
	dec := NewGzipDecoder(WithMaxDecompressionSize(maxDecompressionSize))

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestGzipReuse(t *testing.T) {
	enc, err := NewGzipEncoder()
	require.NoError(t, err)
	dec := NewGzipDecoder(WithMaxDecompressionSize(maxDecompressionSize))

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))

	payload, err = enc.Encode([]byte("doody"))
	require.NoError(t, err)

	actual, err = dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "doody", string(actual))
}

func TestZlibEncodeDecode(t *testing.T) {
	enc, err := NewZlibEncoder()
	require.NoError(t, err)
	dec := NewZlibDecoder(WithMaxDecompressionSize(maxDecompressionSize))

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestZlibEncodeDecodeWithTooLargeMessage(t *testing.T) {
	enc, err := NewZlibEncoder()
	require.NoError(t, err)
	dec := NewZlibDecoder(WithMaxDecompressionSize(3))

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	_, err = dec.Decode(payload)
	require.ErrorContains(t, err, "size of decoded data exceeds allowed size 3")
}

func TestZstdEncodeDecode(t *testing.T) {
	enc, err := NewZstdEncoder()
	require.NoError(t, err)
	dec, err := NewZstdDecoder(WithMaxDecompressionSize(maxDecompressionSize))
	require.NoError(t, err)

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestZstdReuse(t *testing.T) {
	enc, err := NewZstdEncoder()
	require.NoError(t, err)
	dec, err := NewZstdDecoder(WithMaxDecompressionSize(maxDecompressionSize))
	require.NoError(t, err)

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))

	payload, err = enc.Encode([]byte("doody"))
	require.NoError(t, err)

	actual, err = dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "doody", string(actual))
}

func TestIdentityEncodeDecode(t *testing.T) {
	dec := NewIdentityDecoder(WithMaxDecompressionSize(maxDecompressionSize))
	enc, err := NewIdentityEncoder()
	require.NoError(t, err)

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestStreamIdentityDecode(t *testing.T) {
	var r bytes.Buffer
	n, err := r.WriteString("howdy")
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
		validLevels []int
		errormsg    string
	}{
		{
			algorithm:   "gzip",
			validLevels: []int{0, 1, 9},
			errormsg:    "invalid compression level",
		},
		{
			algorithm:   "zlib",
			validLevels: []int{0, 1, 9},
			errormsg:    "invalid compression level",
		},
		{
			algorithm:   "zstd",
			validLevels: []int{1, 3, 7, 11},
			errormsg:    "invalid compression level",
		},
		{
			algorithm: "identity",
			errormsg:  "does not support options",
		},
	}

	for _, tt := range tests {
		// Check default i.e. without specifying level
		t.Run(tt.algorithm+" default", func(t *testing.T) {
			enc, err := NewContentEncoder(tt.algorithm)
			require.NoError(t, err)
			require.NotNil(t, enc)
		})

		// Check invalid level
		t.Run(tt.algorithm+" invalid", func(t *testing.T) {
			_, err := NewContentEncoder(tt.algorithm, WithCompressionLevel(12))
			require.ErrorContains(t, err, tt.errormsg)
		})

		// Check known levels 0..9
		for level := 0; level < 10; level++ {
			name := fmt.Sprintf("%s level %d", tt.algorithm, level)
			t.Run(name, func(t *testing.T) {
				var valid bool
				for _, l := range tt.validLevels {
					if l == level {
						valid = true
						break
					}
				}

				enc, err := NewContentEncoder(tt.algorithm, WithCompressionLevel(level))
				if valid {
					require.NoError(t, err)
					require.NotNil(t, enc)
				} else {
					require.ErrorContains(t, err, tt.errormsg)
				}
			})
		}
	}
}

func BenchmarkGzipEncode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder(WithMaxDecompressionSize(dataLen))
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
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
	dec := NewGzipDecoder(WithMaxDecompressionSize(dataLen))
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkGzipEncodeDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder(WithMaxDecompressionSize(dataLen))
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkGzipEncodeBig(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 1024*1024))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder(WithMaxDecompressionSize(dataLen))
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
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
	dec := NewGzipDecoder(WithMaxDecompressionSize(dataLen))
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkGzipEncodeDecodeBig(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 1024*1024))
	dataLen := int64(len(data)) + 1

	enc, err := NewGzipEncoder()
	require.NoError(b, err)
	dec := NewGzipDecoder(WithMaxDecompressionSize(dataLen))
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkZstdEncode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewZstdEncoder()
	require.NoError(b, err)
	dec, err := NewZstdDecoder(WithMaxDecompressionSize(dataLen))
	require.NoError(b, err)
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err := enc.Encode(data)
		require.NoError(b, err)
	}
}

func BenchmarkZstdDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewZstdEncoder()
	require.NoError(b, err)
	dec, err := NewZstdDecoder(WithMaxDecompressionSize(dataLen))
	require.NoError(b, err)
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkZstdEncodeDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewZstdEncoder()
	require.NoError(b, err)
	dec, err := NewZstdDecoder(WithMaxDecompressionSize(dataLen))
	require.NoError(b, err)
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkZstdEncodeBig(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 1024*1024))
	dataLen := int64(len(data)) + 1

	enc, err := NewZstdEncoder()
	require.NoError(b, err)
	dec, err := NewZstdDecoder(WithMaxDecompressionSize(dataLen))
	require.NoError(b, err)
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err := enc.Encode(data)
		require.NoError(b, err)
	}
}

func BenchmarkZstdDecodeBig(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 1024*1024))
	dataLen := int64(len(data)) + 1

	enc, err := NewZstdEncoder()
	require.NoError(b, err)
	dec, err := NewZstdDecoder(WithMaxDecompressionSize(dataLen))
	require.NoError(b, err)
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkZstdEncodeDecodeBig(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 1024*1024))
	dataLen := int64(len(data)) + 1

	enc, err := NewZstdEncoder()
	require.NoError(b, err)
	dec, err := NewZstdDecoder(WithMaxDecompressionSize(dataLen))
	require.NoError(b, err)
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkZlibEncode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewZlibEncoder()
	require.NoError(b, err)
	dec := NewZlibDecoder(WithMaxDecompressionSize(dataLen))
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
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
	dec := NewZlibDecoder(WithMaxDecompressionSize(dataLen))
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkZlibEncodeDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	enc, err := NewZlibEncoder()
	require.NoError(b, err)
	dec := NewZlibDecoder(WithMaxDecompressionSize(dataLen))
	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}

func BenchmarkIdentityEncodeDecode(b *testing.B) {
	data := []byte(strings.Repeat("-howdy stranger-", 64))
	dataLen := int64(len(data)) + 1

	dec := NewIdentityDecoder(WithMaxDecompressionSize(dataLen))
	enc, err := NewIdentityEncoder()
	require.NoError(b, err)

	payload, err := enc.Encode(data)
	require.NoError(b, err)
	actual, err := dec.Decode(payload)
	require.NoError(b, err)
	require.Equal(b, data, actual)

	for n := 0; n < b.N; n++ {
		payload, err := enc.Encode(data)
		require.NoError(b, err)

		_, err = dec.Decode(payload)
		require.NoError(b, err)
	}
}
