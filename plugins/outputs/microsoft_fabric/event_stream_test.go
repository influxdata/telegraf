package microsoft_fabric

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsEventstreamEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "endpoint prefix",
			endpoint: "Endpoint=sb://example.com",
		},
		{
			name:     "case insensitive prefix",
			endpoint: "Endpoint=sb://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, isEventstreamEndpoint(tt.endpoint))
		})
	}
}

func TestIsNotEventstreamEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "Invalid prefix",
			endpoint: "invalid=https://example.com",
		},
		{
			name:     "Empty string",
			endpoint: "",
		},
		{
			name:     "Just URL",
			endpoint: "https://example.com",
		},
		{
			name:     "eventhouse endpoint",
			endpoint: "data source=https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.False(t, isEventstreamEndpoint(tt.endpoint))
		})
	}
}
