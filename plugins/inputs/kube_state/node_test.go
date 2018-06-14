package kube_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
						Items: []v1.Node{
							{
								Status: v1.NodeStatus{
									NodeInfo: v1.NodeSystemInfo{
										KernelVersion:           "4.14.48-coreos-r2",
										OSImage:                 "Container Linux by CoreOS 1745.7.0 (Rhyolite)",
										ContainerRuntimeVersion: "docker://18.3.1",
										KubeletVersion:          "v1.10.3",
										KubeProxyVersion:        "v1.10.3",
									},
									Phase: v1.NodeRunning,
									Capacity: v1.ResourceList{
										v1.ResourceCPU:                   resource.MustParse("16"),
										v1.ResourceEphemeralStorage:      resource.MustParse("48375392Ki"),
										v1.ResourceName("hugepages-1Gi"): resource.MustParse("0"),
										v1.ResourceName("hugepages-2Mi"): resource.MustParse("0"),
										v1.ResourceMemory:                resource.MustParse("125817984Ki"),
										v1.ResourcePods:                  resource.MustParse("110"),
									},
									Allocatable: v1.ResourceList{
										v1.ResourceCPU:                   resource.MustParse("16"),
										v1.ResourceEphemeralStorage:      resource.MustParse("44582761194"),
										v1.ResourceName("hugepages-1Gi"): resource.MustParse("0"),
										v1.ResourceName("hugepages-2Mi"): resource.MustParse("0"),
										v1.ResourceMemory:                resource.MustParse("125715584Ki"),
										v1.ResourcePods:                  resource.MustParse("110"),
									},
									Conditions: []v1.NodeCondition{
										{Type: v1.NodeReady, Status: v1.ConditionTrue, LastTransitionTime: metav1.Time{Time: now}},
										{Type: v1.NodeOutOfDisk, Status: v1.ConditionFalse, LastTransitionTime: metav1.Time{Time: created}},
									},
								},
								Spec: v1.NodeSpec{
									ProviderID:    "aws:///us-east-1c/i-0c00",
									Unschedulable: false,
									Taints: []v1.Taint{
										{
											Key:    "k1",
											Value:  "v1",
											Effect: v1.TaintEffectNoExecute,
										},
										{
											Key:    "k2",
											Value:  "v2",
											Effect: v1.TaintEffectNoSchedule,
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation: int64(11232),
									Namespace:  "ns1",
									Name:       "node1",
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: metav1.Time{Time: created},
								},
							},
						},
					},
				},
			},
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					{
						Measurement: nodeStatusConditionsMeasurement,
						Fields: map[string]interface{}{
							"gauge": 1,
						},
						Tags: map[string]string{
							"node":      "node1",
							"condition": "ready",
							"status":    "true",
						},
					},
					{
						Measurement: nodeStatusConditionsMeasurement,
						Fields: map[string]interface{}{
							"gauge": 1,
						},
						Tags: map[string]string{
							"node":      "node1",
							"condition": "outofdisk",
							"status":    "false",
						},
					},
					{
						Measurement: nodeTaintMeasurement,
						Fields: map[string]interface{}{
							"gauge": 1,
						},
						Tags: map[string]string{
							"node":   "node1",
							"key":    "k1",
							"value":  "v1",
							"effect": "NoExecute",
						},
					},
					{
						Measurement: nodeTaintMeasurement,
						Fields: map[string]interface{}{
							"gauge": 1,
						},
						Tags: map[string]string{
							"node":   "node1",
							"key":    "k2",
							"value":  "v2",
							"effect": "NoSchedule",
						},
					},
					{
						Measurement: nodeMeasurement,
						Fields: map[string]interface{}{
							"created":                                    created.Unix(),
							"status_capacity_cpu_cores":                  int64(16),
							"status_capacity_ephemeral_storage_bytes":    int64(49536401408),
							"status_capacity_hugepages_1Gi_bytes":        int64(0),
							"status_capacity_hugepages_2Mi_bytes":        int64(0),
							"status_capacity_memory_bytes":               int64(128837615616),
							"status_capacity_pods":                       int64(110),
							"status_allocatable_cpu_cores":               int64(16),
							"status_allocatable_ephemeral_storage_bytes": int64(44582761194),
							"status_allocatable_hugepages_1Gi_bytes":     int64(0),
							"status_allocatable_hugepages_2Mi_bytes":     int64(0),
							"status_allocatable_memory_bytes":            int64(128732758016),
							"status_allocatable_pods":                    int64(110),
						},
						Tags: map[string]string{
							"node":                      "node1",
							"label_lab1":                "v1",
							"label_lab2":                "v2",
							"kernel_version":            "4.14.48-coreos-r2",
							"os_image":                  "Container Linux by CoreOS 1745.7.0 (Rhyolite)",
							"container_runtime_version": "docker://18.3.1",
							"kubelet_version":           "v1.10.3",
							"kubeproxy_version":         "v1.10.3",
							"status_phase":              "running",
							"provider_id":               "aws:///us-east-1c/i-0c00",
							"spec_unschedulable":        "false",
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
		ks := &KubenetesState{
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
				case nodeStatusConditionsMeasurement:
					keyTag = "condition"
				case nodeMeasurement:
					keyTag = "node"
				case nodeTaintMeasurement:
					keyTag = "key"
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
