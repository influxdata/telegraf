package kube_inventory

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/stretchr/testify/require"
)

type mockHandler struct {
	responseMap map[string]interface{}
}

func toStrPtr(s string) *string {
	return &s
}

func toInt32Ptr(i int32) *int32 {
	return &i
}

func toBoolPtr(b bool) *bool {
	return &b
}

func TestNewClient(t *testing.T) {
	_, err := newClient("https://127.0.0.1:443/", "default", "abc123", time.Second, tls.ClientConfig{})
	require.NoErrorf(t, err, "Failed to create new client - %v", err)
}
