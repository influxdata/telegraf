package jolokia2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeReadResponses_Jolokia1xTarget(t *testing.T) {
	// Jolokia 1.x returns target directly in request.target
	jresponses := []jolokiaResponse{
		{
			Request: jolokiaResponseRequest{
				jolokiaRequest: jolokiaRequest{
					Type:      "read",
					Mbean:     "java.lang:type=Runtime",
					Attribute: "Uptime",
					Target: &jolokiaTarget{
						URL: "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi",
					},
				},
			},
			Value:  1214083,
			Status: 200,
		},
	}

	responses := makeReadResponses(jresponses)

	require.Len(t, responses, 1)
	require.Equal(t, "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi", responses[0].RequestTarget)
	require.Equal(t, "java.lang:type=Runtime", responses[0].RequestMbean)
	require.Equal(t, 200, responses[0].Status)
}

func TestMakeReadResponses_Jolokia2xTarget(t *testing.T) {
	// Jolokia 2.x returns target in request.options.target
	jresponses := []jolokiaResponse{
		{
			Request: jolokiaResponseRequest{
				jolokiaRequest: jolokiaRequest{
					Type:      "read",
					Mbean:     "java.lang:type=Runtime",
					Attribute: "Uptime",
				},
				Options: &jolokiaOptions{
					Target: &jolokiaTarget{
						URL: "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi",
					},
				},
			},
			Value:  1214083,
			Status: 200,
		},
	}

	responses := makeReadResponses(jresponses)

	require.Len(t, responses, 1)
	require.Equal(t, "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi", responses[0].RequestTarget)
	require.Equal(t, "java.lang:type=Runtime", responses[0].RequestMbean)
	require.Equal(t, 200, responses[0].Status)
}

func TestMakeReadResponses_NoTarget(t *testing.T) {
	// No target (direct agent connection, not proxy)
	jresponses := []jolokiaResponse{
		{
			Request: jolokiaResponseRequest{
				jolokiaRequest: jolokiaRequest{
					Type:      "read",
					Mbean:     "java.lang:type=Runtime",
					Attribute: "Uptime",
				},
			},
			Value:  1214083,
			Status: 200,
		},
	}

	responses := makeReadResponses(jresponses)

	require.Len(t, responses, 1)
	require.Empty(t, responses[0].RequestTarget)
	require.Equal(t, "java.lang:type=Runtime", responses[0].RequestMbean)
}

func TestMakeReadResponses_Jolokia1xTakesPrecedence(t *testing.T) {
	// If both locations have a target, Jolokia 1.x location takes precedence
	jresponses := []jolokiaResponse{
		{
			Request: jolokiaResponseRequest{
				jolokiaRequest: jolokiaRequest{
					Type:      "read",
					Mbean:     "java.lang:type=Runtime",
					Attribute: "Uptime",
					Target: &jolokiaTarget{
						URL: "service:jmx:rmi:///jndi/rmi://target1:9010/jmxrmi",
					},
				},
				Options: &jolokiaOptions{
					Target: &jolokiaTarget{
						URL: "service:jmx:rmi:///jndi/rmi://target2:9010/jmxrmi",
					},
				},
			},
			Value:  1214083,
			Status: 200,
		},
	}

	responses := makeReadResponses(jresponses)

	require.Len(t, responses, 1)
	// Jolokia 1.x location (request.target) takes precedence
	require.Equal(t, "service:jmx:rmi:///jndi/rmi://target1:9010/jmxrmi", responses[0].RequestTarget)
}

func TestMakeReadResponses_MultipleTargets(t *testing.T) {
	// Multiple responses with different targets (Jolokia 2.x format)
	jresponses := []jolokiaResponse{
		{
			Request: jolokiaResponseRequest{
				jolokiaRequest: jolokiaRequest{
					Type:      "read",
					Mbean:     "java.lang:type=Runtime",
					Attribute: "Uptime",
				},
				Options: &jolokiaOptions{
					Target: &jolokiaTarget{
						URL: "service:jmx:rmi:///jndi/rmi://host1:9010/jmxrmi",
					},
				},
			},
			Value:  1000,
			Status: 200,
		},
		{
			Request: jolokiaResponseRequest{
				jolokiaRequest: jolokiaRequest{
					Type:      "read",
					Mbean:     "java.lang:type=Runtime",
					Attribute: "Uptime",
				},
				Options: &jolokiaOptions{
					Target: &jolokiaTarget{
						URL: "service:jmx:rmi:///jndi/rmi://host2:9010/jmxrmi",
					},
				},
			},
			Value:  2000,
			Status: 200,
		},
	}

	responses := makeReadResponses(jresponses)

	require.Len(t, responses, 2)
	require.Equal(t, "service:jmx:rmi:///jndi/rmi://host1:9010/jmxrmi", responses[0].RequestTarget)
	require.Equal(t, "service:jmx:rmi:///jndi/rmi://host2:9010/jmxrmi", responses[1].RequestTarget)
}
