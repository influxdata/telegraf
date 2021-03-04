package kube_inventory

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/influxdata/telegraf/testutil"
)

func TestNode(t *testing.T) {
	cli := &client{}
	now := time.Now()
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
					"/nodes/": corev1.NodeList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect nodes",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/nodes/": corev1.NodeList{
						Items: []corev1.Node{
							{
								Status: corev1.NodeStatus{
									NodeInfo: corev1.NodeSystemInfo{
										KernelVersion:           "4.14.48-coreos-r2",
										OSImage:                 "Container Linux by CoreOS 1745.7.0 (Rhyolite)",
										ContainerRuntimeVersion: "docker://18.3.1",
										KubeletVersion:          "v1.10.3",
										KubeProxyVersion:        "v1.10.3",
									},
									Phase: "Running",
									Capacity: corev1.ResourceList{
										"cpu":                     resource.Quantity{Format: "16"},
										"ephemeral_storage_bytes": resource.Quantity{Format: "49536401408"},
										"hugepages_1Gi_bytes":     resource.Quantity{Format: "0"},
										"hugepages_2Mi_bytes":     resource.Quantity{Format: "0"},
										"memory":                  resource.Quantity{Format: "125817904Ki"},
										"pods":                    resource.Quantity{Format: "110"},
									},
									Allocatable: corev1.ResourceList{
										"cpu":                     resource.Quantity{Format: "1000m"},
										"ephemeral_storage_bytes": resource.Quantity{Format: "44582761194"},
										"hugepages_1Gi_bytes":     resource.Quantity{Format: "0"},
										"hugepages_2Mi_bytes":     resource.Quantity{Format: "0"},
										"memory":                  resource.Quantity{Format: "125715504Ki"},
										"pods":                    resource.Quantity{Format: "110"},
									},
									Conditions: []corev1.NodeCondition{
										{Type: "Ready", Status: "true", LastTransitionTime: metav1.Time{Time: now}},
										{Type: "OutOfDisk", Status: "false", LastTransitionTime: metav1.Time{Time: created}},
									},
								},
								Spec: corev1.NodeSpec{
									ProviderID: "aws:///us-east-1c/i-0c00",
									Taints: []corev1.Taint{
										{
											Key:    "k1",
											Value:  "v1",
											Effect: "NoExecute",
										},
										{
											Key:    "k2",
											Value:  "v2",
											Effect: "NoSchedule",
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation: 11232,
									Namespace:  "ns1",
									Name:       "node1",
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
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
						Measurement: nodeMeasurement,
						Fields: map[string]interface{}{
							"capacity_cpu_cores":         int64(16),
							"capacity_millicpu_cores":    int64(16000),
							"capacity_memory_bytes":      int64(1.28837533696e+11),
							"capacity_pods":              int64(110),
							"allocatable_cpu_cores":      int64(1),
							"allocatable_millicpu_cores": int64(1000),
							"allocatable_memory_bytes":   int64(1.28732676096e+11),
							"allocatable_pods":           int64(110),
						},
						Tags: map[string]string{
							"node_name": "node1",
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
		for _, node := range ((v.handler.responseMap["/nodes/"]).(corev1.NodeList)).Items {
			err := ks.gatherNode(node, acc)
			if err != nil {
				t.Errorf("Failed to gather node - %s", err.Error())
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
