package kube_inventory

import (
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	cli := &client{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   []telegraf.Metric
		hasError bool
		include  []string
		exclude  []string
	}{
		{
			name: "no service",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/service/": &corev1.ServiceList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect service",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/service/": &corev1.ServiceList{
						Items: []corev1.Service{
							{
								Spec: corev1.ServiceSpec{
									Ports: []corev1.ServicePort{
										{
											Port: 8080,
											TargetPort: intstr.IntOrString{
												IntVal: 1234,
											},
											Name:     "diagnostic",
											Protocol: "TCP",
										},
									},
									ExternalIPs: []string{"1.0.0.127"},
									ClusterIP:   "127.0.0.1",
									Selector: map[string]string{
										"select1": "s1",
										"select2": "s2",
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "checker",
									CreationTimestamp: metav1.Time{Time: now},
								},
							},
						},
					},
				},
			},

			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_service",
					map[string]string{
						"service_name":     "checker",
						"namespace":        "ns1",
						"port_name":        "diagnostic",
						"port_protocol":    "TCP",
						"cluster_ip":       "127.0.0.1",
						"ip":               "1.0.0.127",
						"selector_select1": "s1",
						"selector_select2": "s2",
					},
					map[string]interface{}{
						"port":        int32(8080),
						"target_port": int32(1234),
						"generation":  int64(12),
						"created":     now.UnixNano(),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
	}

	for _, v := range tests {
		ks := &KubernetesInventory{
			client: cli,
		}
		ks.SelectorInclude = v.include
		ks.SelectorExclude = v.exclude
		require.NoError(t, ks.createSelectorFilters())
		acc := new(testutil.Accumulator)
		for _, service := range ((v.handler.responseMap["/service/"]).(*corev1.ServiceList)).Items {
			ks.gatherService(service, acc)
		}

		err := acc.FirstError()
		if v.hasError {
			require.Errorf(t, err, "%s failed, should have error", v.name)
			continue
		}

		// No error case
		require.NoErrorf(t, err, "%s failed, err: %v", v.name, err)

		require.Len(t, acc.Metrics, len(v.output))
		testutil.RequireMetricsEqual(t, acc.GetTelegrafMetrics(), v.output, testutil.IgnoreTime())
	}
}

func TestServiceSelectorFilter(t *testing.T) {
	cli := &client{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	responseMap := map[string]interface{}{
		"/service/": &corev1.ServiceList{
			Items: []corev1.Service{
				{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Port: 8080,
								TargetPort: intstr.IntOrString{
									IntVal: 1234,
								},
								Name:     "diagnostic",
								Protocol: "TCP",
							},
						},
						ExternalIPs: []string{"1.0.0.127"},
						ClusterIP:   "127.0.0.1",
						Selector: map[string]string{
							"select1": "s1",
							"select2": "s2",
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						Generation:        12,
						Namespace:         "ns1",
						Name:              "checker",
						CreationTimestamp: metav1.Time{Time: now},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		handler  *mockHandler
		hasError bool
		include  []string
		exclude  []string
		expected map[string]string
	}{
		{
			name: "nil filters equals all selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  nil,
			exclude:  nil,
			expected: map[string]string{
				"selector_select1": "s1",
				"selector_select2": "s2",
			},
		},
		{
			name: "empty filters equals all selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{},
			exclude:  []string{},
			expected: map[string]string{
				"selector_select1": "s1",
				"selector_select2": "s2",
			},
		},
		{
			name: "include filter equals only include-matched selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{"select1"},
			exclude:  []string{},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
		{
			name: "exclude filter equals only non-excluded selectors (overrides include filter)",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{},
			exclude:  []string{"select2"},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
		{
			name: "include glob filter equals only include-matched selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{"*1"},
			exclude:  []string{},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
		{
			name: "exclude glob filter equals only non-excluded selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{},
			exclude:  []string{"*2"},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
		{
			name: "exclude glob filter equals only non-excluded selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{},
			exclude:  []string{"*2"},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
	}
	for _, v := range tests {
		ks := &KubernetesInventory{
			client: cli,
		}
		ks.SelectorInclude = v.include
		ks.SelectorExclude = v.exclude
		require.NoError(t, ks.createSelectorFilters())
		acc := new(testutil.Accumulator)
		for _, service := range ((v.handler.responseMap["/service/"]).(*corev1.ServiceList)).Items {
			ks.gatherService(service, acc)
		}

		// Grab selector tags
		actual := map[string]string{}
		for _, metric := range acc.Metrics {
			for key, val := range metric.Tags {
				if strings.Contains(key, "selector_") {
					actual[key] = val
				}
			}
		}

		require.Equalf(t, v.expected, actual,
			"actual selector tags (%v) do not match expected selector tags (%v)", actual, v.expected)
	}
}
