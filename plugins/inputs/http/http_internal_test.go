package http

import (
	"compress/gzip"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMakeRequestBodyReaderEmptyBody(t *testing.T) {
	body := makeRequestBodyReader("", "")
	require.Nil(t, body)
}

func TestMakeRequestBodyReaderNoEncoding(t *testing.T) {
	body := makeRequestBodyReader("", "payload")
	require.NotNil(t, body)
	t.Cleanup(func() { _ = body.Close() })

	actual, err := io.ReadAll(body)
	require.NoError(t, err)
	require.Equal(t, []byte("payload"), actual)
}

func TestMakeRequestBodyReaderGzip(t *testing.T) {
	body := makeRequestBodyReader("gzip", "payload")
	require.NotNil(t, body)
	t.Cleanup(func() { _ = body.Close() })

	reader, err := gzip.NewReader(body)
	require.NoError(t, err)
	t.Cleanup(func() { _ = reader.Close() })

	actual, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, []byte("payload"), actual)
}

func TestGatherURLEarlyFailureWithGzipBody(t *testing.T) {
	h := &HTTP{
		Method:          "BAD METHOD",
		Body:            "payload",
		ContentEncoding: "gzip",
	}

	done := make(chan error, 1)
	go func() {
		done <- h.gatherURL(nil, "http://example.com")
	}()

	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("gatherURL timed out on early failure")
	}
}
