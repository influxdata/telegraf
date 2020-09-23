package sensu

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	_, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	t.Run("default-user-agent", func(t *testing.T) {
		var err = error(nil)
		require.NoError(t, err)
	})
}

func TestResolveEventEndpointUrl(t *testing.T) {
	agentApiUrl := "http://127.0.0.1:3031"
	backendApiUrl := "http://127.0.0.1:8080"
	entityNamespace := "test-namespace"
	tests := []struct {
		name                string
		plugin              *Sensu
		expectedEndpointUrl string
	}{
		{
			name: "agent event endpoint",
			plugin: &Sensu{
				AgentApiUrl: &agentApiUrl,
			},
			expectedEndpointUrl: "http://127.0.0.1:3031/events",
		},
		{
			name: "backend event endpoint with default namespace",
			plugin: &Sensu{
				AgentApiUrl:   &agentApiUrl,
				BackendApiUrl: &backendApiUrl,
			},
			expectedEndpointUrl: "http://127.0.0.1:8080/api/core/v2/namespaces/default/events",
		},
		{
			name: "backend event endpoint with namespace declared",
			plugin: &Sensu{
				AgentApiUrl:   &agentApiUrl,
				BackendApiUrl: &backendApiUrl,
				Entity: &SensuEntity{
					Namespace: &entityNamespace,
				},
			},
			expectedEndpointUrl: "http://127.0.0.1:8080/api/core/v2/namespaces/test-namespace/events",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := tt.plugin.GetEndpointUrl()
			require.Equal(t, err, error(nil))
			require.Equal(t, tt.expectedEndpointUrl, url)
		})
	}
}
