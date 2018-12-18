package kube_lite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ericchiang/k8s/apis/resource"

	"github.com/influxdata/telegraf/testutil"
)

func TestNode(t *testing.T) {
	cli := &client{
		httpClient: &http.Client{Transport: &http.Transport{}},
		semaphore:  make(chan struct{}, 1),
	}
	now := time.Now()
	//started := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 1, 36, 0, now.Location())
	created := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-2, 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   *testutil.Accumulator
		hasError bool
	}{
		{
			name: "no nodes",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/nodes/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "collect nodes",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/nodes/": &v1.NodeList{
						Items: []*v1.Node{
							{
								Status: &v1.NodeStatus{
									NodeInfo: &v1.NodeSystemInfo{
										KernelVersion:           toStrPtr("4.14.48-coreos-r2"),
										OsImage:                 toStrPtr("Container Linux by CoreOS 1745.7.0 (Rhyolite)"),
										ContainerRuntimeVersion: toStrPtr("docker://18.3.1"),
										KubeletVersion:          toStrPtr("v1.10.3"),
										KubeProxyVersion:        toStrPtr("v1.10.3"),
									},
									Phase: toStrPtr("Running"),
									Capacity: map[string]*resource.Quantity{
										"cpu_cores":               &resource.Quantity{String_: toStrPtr("16")},
										"ephemeral_storage_bytes": &resource.Quantity{String_: toStrPtr("49536401408")},
										"hugepages_1Gi_bytes":     &resource.Quantity{String_: toStrPtr("0")},
										"hugepages_2Mi_bytes":     &resource.Quantity{String_: toStrPtr("0")},
										"memory_bytes":            &resource.Quantity{String_: toStrPtr("128837615616")},
										"pods":                    &resource.Quantity{String_: toStrPtr("110")},
									},
									Allocatable: map[string]*resource.Quantity{
										"cpu_cores":               &resource.Quantity{String_: toStrPtr("16")},
										"ephemeral_storage_bytes": &resource.Quantity{String_: toStrPtr("44582761194")},
										"hugepages_1Gi_bytes":     &resource.Quantity{String_: toStrPtr("0")},
										"hugepages_2Mi_bytes":     &resource.Quantity{String_: toStrPtr("0")},
										"memory_bytes":            &resource.Quantity{String_: toStrPtr("128732758016")},
										"pods":                    &resource.Quantity{String_: toStrPtr("110")},
									},
									Conditions: []*v1.NodeCondition{
										{Type: toStrPtr("Ready"), Status: toStrPtr("true"), LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(now.Unix())}},
										{Type: toStrPtr("OutOfDisk"), Status: toStrPtr("false"), LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(created.Unix())}},
									},
								},
								Spec: &v1.NodeSpec{
									ProviderID: toStrPtr("aws:///us-east-1c/i-0c00"),
									Taints: []*v1.Taint{
										{
											Key:    toStrPtr("k1"),
											Value:  toStrPtr("v1"),
											Effect: toStrPtr("NoExecute"),
										},
										{
											Key:    toStrPtr("k2"),
											Value:  toStrPtr("v2"),
											Effect: toStrPtr("NoSchedule"),
										},
									},
								},
								Metadata: &metav1.ObjectMeta{
									Generation: toInt64Ptr(int64(11232)),
									Namespace:  toStrPtr("ns1"),
									Name:       toStrPtr("node1"),
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(created.Unix())},
								},
							},
						},
					},
				},
			},
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					{
						Measurement: nodeMeasurement,
						Fields: map[string]interface{}{
							// "created":                                    created.Unix(),
							// "status_capacity_cpu_cores":                  int64(16),
							// "status_capacity_ephemeral_storage_bytes":    int64(49536401408),
							// "status_capacity_hugepages_1Gi_bytes":        int64(0),
							// "status_capacity_hugepages_2Mi_bytes":        int64(0),
							// "status_capacity_memory_bytes":               int64(128837615616),
							"status_capacity_pods": int64(110),
							// "status_allocatable_cpu_cores":               int64(16),
							// "status_allocatable_ephemeral_storage_bytes": int64(44582761194),
							// "status_allocatable_hugepages_1Gi_bytes":     int64(0),
							// "status_allocatable_hugepages_2Mi_bytes":     int64(0),
							// "status_allocatable_memory_bytes":            int64(128732758016),
							// "status_allocatable_pods":                    int64(110),
						},
						Tags: map[string]string{
							"name": "node1",
							// "label_lab1":                "v1",
							// "label_lab2":                "v2",
							// "kernel_version":            "4.14.48-coreos-r2",
							// "os_image":                  "Container Linux by CoreOS 1745.7.0 (Rhyolite)",
							// "container_runtime_version": "docker://18.3.1",
							// "kubelet_version":           "v1.10.3",
							// "kubeproxy_version":         "v1.10.3",
							// "status_phase":              "running",
							// "provider_id":               "aws:///us-east-1c/i-0c00",
							// "spec_unschedulable":        "false",
						},
					},
				},
			},
			hasError: false,
		},
	}
	for _, v := range tests {
		ts := httptest.NewServer(v.handler)
		defer ts.Close()

		cli.baseURL = ts.URL
		ks := &KubernetesState{
			client: cli,
		}
		acc := new(testutil.Accumulator)
		registerNodeCollector(context.Background(), acc, ks)
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
				measurement := v.output.Metrics[i].Measurement
				var keyTag string
				switch measurement {
				case nodeMeasurement:
					keyTag = "node"
				}
				var j int
				for j = range acc.Metrics {
					if acc.Metrics[j].Measurement == measurement &&
						acc.Metrics[j].Tags[keyTag] == v.output.Metrics[i].Tags[keyTag] {
						break
					}
				}

				for k, m := range v.output.Metrics[i].Tags {
					if acc.Metrics[j].Tags[k] != m {
						t.Fatalf("%s: tag %s metrics unmatch Expected %s, got %s, measurement %s, j %d\n", v.name, k, m, acc.Metrics[j].Tags[k], measurement, j)
					}
				}
				for k, m := range v.output.Metrics[i].Fields {
					if acc.Metrics[j].Fields[k] != m {
						t.Fatalf("%s: field %s metrics unmatch Expected %v(%T), got %v(%T), measurement %s, j %d\n", v.name, k, m, m, acc.Metrics[j].Fields[k], acc.Metrics[i].Fields[k], measurement, j)
					}
				}
			}
		}

	}
}

func toStrPtr(s string) *string {
	return &s
}

func toInt32Ptr(i int32) *int32 {
	return &i
}

func toInt64Ptr(i int64) *int64 {
	return &i
}

func toBoolPtr(b bool) *bool {
	return &b
}
