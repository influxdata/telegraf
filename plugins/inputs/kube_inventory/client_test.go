package kube_inventory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

type mockHandler struct {
	responseMap map[string]interface{}
}

func toPtr[T any](v T) *T {
	return &v
}

func TestNewClient(t *testing.T) {
	_, err := newClient("https://127.0.0.1:443/", "default", "", "abc123", time.Second, tls.ClientConfig{})
	require.NoErrorf(t, err, "Failed to create new client: %v", err)

	_, err = newClient("https://127.0.0.1:443/", "default", "nonexistantFile", "", time.Second, tls.ClientConfig{})
	require.Errorf(t, err, "Failed to read token file \"file\": open file: no such file or directory: %v", err)
}
