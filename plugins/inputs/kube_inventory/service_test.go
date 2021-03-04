package kube_inventory

import (
	"reflect"

	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"strings"
)

func TestService(t *testing.T) {
	cli := &client{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   *testutil.Accumulator
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

			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					{
						Fields: map[string]interface{}{
							"port":        int32(8080),
							"target_port": int32(1234),
							"generation":  int64(12),
							"created":     time.Unix(int64(now.Second()), int64(now.Nanosecond())).UnixNano(),
						},
						Tags: map[string]string{
							"service_name":     "checker",
							"namespace":        "ns1",
							"port_name":        "diagnostic",
							"port_protocol":    "TCP",
							"cluster_ip":       "127.0.0.1",
							"ip":               "1.0.0.127",
							"selector_select1": "s1",
							"selector_select2": "s2",
						},
					},
				},
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
		ks.createSelectorFilters()
		acc := new(testutil.Accumulator)
		for _, service := range ((v.handler.responseMap["/service/"]).(*corev1.ServiceList)).Items {
			err := ks.gatherService(service, acc)
			if err != nil {
				t.Errorf("Failed to gather service - %s", err.Error())
			}
		}

		err := acc.FirstError()
		if err == nil && v.hasError {
			t.Fatalf("%s failed, should have error", v.name)
		} else if err != nil && !v.hasError {
			t.Fatalf("%s failed, err: %v", v.name, err)
		}
		if v.output == nil && len(acc.Metrics) > 0 {
			t.Fatalf("%s: collected extra data", v.name)
		} else if v.output != nil && len(v.output.Metrics) > 0 {
			for i := range v.output.Metrics {
				for k, m := range v.output.Metrics[i].Tags {
					if acc.Metrics[i].Tags[k] != m {
						t.Fatalf("%s: tag %s metrics unmatch Expected %s, got '%v'\n", v.name, k, m, acc.Metrics[i].Tags[k])
					}
				}
				for k, m := range v.output.Metrics[i].Fields {
					if acc.Metrics[i].Fields[k] != m {
						t.Fatalf("%s: field %s metrics unmatch Expected %v(%T), got %v(%T)\n", v.name, k, m, m, acc.Metrics[i].Fields[k], acc.Metrics[i].Fields[k])
					}
				}
			}
		}
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
		ks.createSelectorFilters()
		acc := new(testutil.Accumulator)
		for _, service := range ((v.handler.responseMap["/service/"]).(*corev1.ServiceList)).Items {
			err := ks.gatherService(service, acc)
			if err != nil {
				t.Errorf("Failed to gather service - %s", err.Error())
			}
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

		if !reflect.DeepEqual(v.expected, actual) {
			t.Fatalf("actual selector tags (%v) do not match expected selector tags (%v)", actual, v.expected)
		}
	}
}
