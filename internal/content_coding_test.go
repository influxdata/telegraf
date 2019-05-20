package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGzipEncodeDecode(t *testing.T) {
	enc, err := NewGzipEncoder()
	require.NoError(t, err)
	dec, err := NewGzipDecoder()
	require.NoError(t, err)

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}

func TestGzipReuse(t *testing.T) {
	enc, err := NewGzipEncoder()
	require.NoError(t, err)
	dec, err := NewGzipDecoder()
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
	enc := NewIdentityEncoder()
	dec := NewIdentityDecoder()

	payload, err := enc.Encode([]byte("howdy"))
	require.NoError(t, err)

	actual, err := dec.Decode(payload)
	require.NoError(t, err)

	require.Equal(t, "howdy", string(actual))
}
