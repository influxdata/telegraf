package kube_inventory

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/influxdata/telegraf/testutil"
)

func TestEndpoint(t *testing.T) {
	cli := &client{}

	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   *testutil.Accumulator
		hasError bool
	}{
		{
			name: "no endpoints",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &v1.EndpointsList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect ready endpoints",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &v1.EndpointsList{
						Items: []*v1.Endpoints{
							{
								Subsets: []*v1.EndpointSubset{
									{
										Addresses: []*v1.EndpointAddress{
											{
												Hostname: toStrPtr("storage-6"),
												NodeName: toStrPtr("b.storage.internal"),
												TargetRef: &v1.ObjectReference{
													Kind: toStrPtr("pod"),
													Name: toStrPtr("storage-6"),
												},
											},
										},
										Ports: []*v1.EndpointPort{
											{
												Name:     toStrPtr("server"),
												Protocol: toStrPtr("TCP"),
												Port:     toInt32Ptr(8080),
											},
										},
									},
								},
								Metadata: &metav1.ObjectMeta{
									Generation:        toInt64Ptr(12),
									Namespace:         toStrPtr("ns1"),
									Name:              toStrPtr("storage"),
									CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(now.Unix())},
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
							"ready":      true,
							"port":       int32(8080),
							"generation": int64(12),
							"created":    now.UnixNano(),
						},
						Tags: map[string]string{
							"endpoint_name": "storage",
							"namespace":     "ns1",
							"hostname":      "storage-6",
							"node_name":     "b.storage.internal",
							"port_name":     "server",
							"port_protocol": "TCP",
							"pod":           "storage-6",
						},
					},
				},
			},
			hasError: false,
		},
		{
			name: "collect notready endpoints",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/endpoints/": &v1.EndpointsList{
						Items: []*v1.Endpoints{
							{
								Subsets: []*v1.EndpointSubset{
									{
										NotReadyAddresses: []*v1.EndpointAddress{
											{
												Hostname: toStrPtr("storage-6"),
												NodeName: toStrPtr("b.storage.internal"),
												TargetRef: &v1.ObjectReference{
													Kind: toStrPtr("pod"),
													Name: toStrPtr("storage-6"),
												},
											},
										},
										Ports: []*v1.EndpointPort{
											{
												Name:     toStrPtr("server"),
												Protocol: toStrPtr("TCP"),
												Port:     toInt32Ptr(8080),
											},
										},
									},
								},
								Metadata: &metav1.ObjectMeta{
									Generation:        toInt64Ptr(12),
									Namespace:         toStrPtr("ns1"),
									Name:              toStrPtr("storage"),
									CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(now.Unix())},
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
							"ready":      false,
							"port":       int32(8080),
							"generation": int64(12),
							"created":    now.UnixNano(),
						},
						Tags: map[string]string{
							"endpoint_name": "storage",
							"namespace":     "ns1",
							"hostname":      "storage-6",
							"node_name":     "b.storage.internal",
							"port_name":     "server",
							"port_protocol": "TCP",
							"pod":           "storage-6",
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
		acc := new(testutil.Accumulator)
		for _, endpoint := range ((v.handler.responseMap["/endpoints/"]).(*v1.EndpointsList)).Items {
			err := ks.gatherEndpoint(*endpoint, acc)
			if err != nil {
				t.Errorf("Failed to gather endpoint - %s", err.Error())
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
