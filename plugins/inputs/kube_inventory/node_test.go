package kube_inventory

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestNode(t *testing.T) {
	cli := &client{}
	now := time.Now()
	created := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-2, 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   []telegraf.Metric
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
			output: []telegraf.Metric{
				testutil.MustMetric(
					nodeMeasurement,
					map[string]string{
						"node_name": "node1",
					},
					map[string]interface{}{
						"capacity_cpu_cores":         int64(16),
						"capacity_millicpu_cores":    int64(16000),
						"capacity_memory_bytes":      int64(1.28837533696e+11),
						"capacity_pods":              int64(110),
						"allocatable_cpu_cores":      int64(1),
						"allocatable_millicpu_cores": int64(1000),
						"allocatable_memory_bytes":   int64(1.28732676096e+11),
						"allocatable_pods":           int64(110),
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
		acc := new(testutil.Accumulator)
		for _, node := range ((v.handler.responseMap["/nodes/"]).(corev1.NodeList)).Items {
			ks.gatherNode(node, acc)
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
