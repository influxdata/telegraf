package kube_inventory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestEndpoint(t *testing.T) {
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   []telegraf.Metric
		hasError bool
	}{
		{
			name: "no endpoints",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &discoveryv1.EndpointSliceList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect ready endpoints",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &discoveryv1.EndpointSliceList{
						Items: []discoveryv1.EndpointSlice{
							{
								Endpoints: []discoveryv1.Endpoint{
									{
										Hostname: toPtr("storage-6"),
										NodeName: toPtr("b.storage.internal"),
										TargetRef: &corev1.ObjectReference{
											Kind: "pod",
											Name: "storage-6",
										},
										Conditions: discoveryv1.EndpointConditions{
											Ready: toPtr(true),
										},
									},
								},
								Ports: []discoveryv1.EndpointPort{
									{
										Name:     toPtr("server"),
										Protocol: toPtr(corev1.Protocol("TCP")),
										Port:     toPtr(int32(8080)),
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "storage",
									CreationTimestamp: metav1.Time{Time: now},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_endpoint",
					map[string]string{
						"endpoint_name": "storage",
						"namespace":     "ns1",
						"hostname":      "storage-6",
						"node_name":     "b.storage.internal",
						"port_name":     "server",
						"port_protocol": "TCP",
						"pod":           "storage-6",
					},
					map[string]interface{}{
						"ready":      true,
						"port":       int32(8080),
						"generation": int64(12),
						"created":    now.UnixNano(),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
		{
			name: "collect notready endpoints",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &discoveryv1.EndpointSliceList{
						Items: []discoveryv1.EndpointSlice{
							{
								Endpoints: []discoveryv1.Endpoint{
									{
										Hostname: toPtr("storage-6"),
										NodeName: toPtr("b.storage.internal"),
										TargetRef: &corev1.ObjectReference{
											Kind: "pod",
											Name: "storage-6",
										},
										Conditions: discoveryv1.EndpointConditions{
											Ready: toPtr(false),
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "storage",
									CreationTimestamp: metav1.Time{Time: now},
								},
								Ports: []discoveryv1.EndpointPort{
									{
										Name:     toPtr("server"),
										Protocol: toPtr(corev1.Protocol("TCP")),
										Port:     toPtr(int32(8080)),
									},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_endpoint",
					map[string]string{
						"endpoint_name": "storage",
						"namespace":     "ns1",
						"hostname":      "storage-6",
						"node_name":     "b.storage.internal",
						"port_name":     "server",
						"port_protocol": "TCP",
						"pod":           "storage-6",
					},
					map[string]interface{}{
						"ready":      false,
						"port":       int32(8080),
						"generation": int64(12),
						"created":    now.UnixNano(),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
		{
			name: "endpoints missing node_name",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &discoveryv1.EndpointSliceList{
						Items: []discoveryv1.EndpointSlice{
							{
								Endpoints: []discoveryv1.Endpoint{
									{
										Hostname: toPtr("storage-6"),
										TargetRef: &corev1.ObjectReference{
											Kind: "pod",
											Name: "storage-6",
										},
										Conditions: discoveryv1.EndpointConditions{
											Ready: toPtr(false),
										},
									},
									{
										Hostname: toPtr("storage-12"),
										TargetRef: &corev1.ObjectReference{
											Kind: "pod",
											Name: "storage-12",
										},
										Conditions: discoveryv1.EndpointConditions{
											Ready: toPtr(true),
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "storage",
									CreationTimestamp: metav1.Time{Time: now},
								},
								Ports: []discoveryv1.EndpointPort{
									{
										Name:     toPtr("server"),
										Protocol: toPtr(corev1.Protocol("TCP")),
										Port:     toPtr(int32(8080)),
									},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_endpoint",
					map[string]string{
						"endpoint_name": "storage",
						"namespace":     "ns1",
						"hostname":      "storage-6",
						"port_name":     "server",
						"port_protocol": "TCP",
						"pod":           "storage-6",
					},
					map[string]interface{}{
						"ready":      false,
						"port":       int32(8080),
						"generation": int64(12),
						"created":    now.UnixNano(),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"kubernetes_endpoint",
					map[string]string{
						"endpoint_name": "storage",
						"namespace":     "ns1",
						"hostname":      "storage-12",
						"port_name":     "server",
						"port_protocol": "TCP",
						"pod":           "storage-12",
					},
					map[string]interface{}{
						"ready":      true,
						"port":       int32(8080),
						"generation": int64(12),
						"created":    now.UnixNano(),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
		{
			name: "endpoints null",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &discoveryv1.EndpointSliceList{
						Items: []discoveryv1.EndpointSlice{
							{
								Endpoints: nil,
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "storage",
									CreationTimestamp: metav1.Time{Time: now},
								},
								Ports: []discoveryv1.EndpointPort{
									{
										Name:     toPtr("server"),
										Protocol: toPtr(corev1.Protocol("TCP")),
										Port:     toPtr(int32(8080)),
									},
								},
							},
						},
					},
				},
			},
			output:   make([]telegraf.Metric, 0),
			hasError: false,
		},
		{
			name: "default port name",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &discoveryv1.EndpointSliceList{
						Items: []discoveryv1.EndpointSlice{
							{
								Endpoints: []discoveryv1.Endpoint{
									{
										Hostname: toPtr("storage-6"),
										TargetRef: &corev1.ObjectReference{
											Kind: "pod",
											Name: "storage-6",
										},
										Conditions: discoveryv1.EndpointConditions{
											Ready: toPtr(false),
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "storage",
									CreationTimestamp: metav1.Time{Time: now},
								},
								Ports: []discoveryv1.EndpointPort{
									{
										Name:     toPtr(""),
										Protocol: toPtr(corev1.Protocol("TCP")),
										Port:     toPtr(int32(8080)),
									},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_endpoint",
					map[string]string{
						"endpoint_name": "storage",
						"namespace":     "ns1",
						"hostname":      "storage-6",
						"port_name":     "",
						"port_protocol": "TCP",
						"pod":           "storage-6",
					},
					map[string]interface{}{
						"ready":      false,
						"port":       int32(8080),
						"generation": int64(12),
						"created":    now.UnixNano(),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
		{
			name: "ports null",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &discoveryv1.EndpointSliceList{
						Items: []discoveryv1.EndpointSlice{
							{
								Endpoints: []discoveryv1.Endpoint{
									{
										Hostname: toPtr("storage-6"),
										TargetRef: &corev1.ObjectReference{
											Kind: "pod",
											Name: "storage-6",
										},
										Conditions: discoveryv1.EndpointConditions{
											Ready: toPtr(false),
										},
									},
									{
										Hostname: toPtr("storage-12"),
										TargetRef: &corev1.ObjectReference{
											Kind: "pod",
											Name: "storage-12",
										},
										Conditions: discoveryv1.EndpointConditions{
											Ready: toPtr(true),
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "storage",
									CreationTimestamp: metav1.Time{Time: now},
								},
								Ports: nil,
							},
						},
					},
				},
			},
			output:   make([]telegraf.Metric, 0),
			hasError: false,
		},
		{
			name: "endpoints and ports null",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &discoveryv1.EndpointSliceList{
						Items: []discoveryv1.EndpointSlice{
							{
								Endpoints: nil,
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "storage",
									CreationTimestamp: metav1.Time{Time: now},
								},
								Ports: nil,
							},
						},
					},
				},
			},
			output:   make([]telegraf.Metric, 0),
			hasError: false,
		},
		{
			name: "empty conditions",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &discoveryv1.EndpointSliceList{
						Items: []discoveryv1.EndpointSlice{
							{
								Endpoints: []discoveryv1.Endpoint{
									{
										Hostname: toPtr("storage-6"),
										TargetRef: &corev1.ObjectReference{
											Kind: "pod",
											Name: "storage-6",
										},
										Conditions: discoveryv1.EndpointConditions{},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "storage",
									CreationTimestamp: metav1.Time{Time: now},
								},
								Ports: []discoveryv1.EndpointPort{
									{
										Name:     toPtr(""),
										Protocol: toPtr(corev1.Protocol("TCP")),
										Port:     toPtr(int32(8080)),
									},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_endpoint",
					map[string]string{
						"endpoint_name": "storage",
						"namespace":     "ns1",
						"hostname":      "storage-6",
						"port_name":     "",
						"port_protocol": "TCP",
						"pod":           "storage-6",
					},
					map[string]interface{}{
						"ready":      true,
						"port":       int32(8080),
						"generation": int64(12),
						"created":    now.UnixNano(),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
	}

	for _, v := range tests {
		acc := new(testutil.Accumulator)
		for _, endpoint := range ((v.handler.responseMap["/endpoints/"]).(*discoveryv1.EndpointSliceList)).Items {
			gatherEndpoint(endpoint, acc)
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
