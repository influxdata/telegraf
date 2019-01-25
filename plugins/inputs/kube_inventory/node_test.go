package kube_inventory

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ericchiang/k8s/apis/resource"

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
					"/nodes/": &v1.NodeList{},
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
										"cpu":                     {String_: toStrPtr("16")},
										"ephemeral_storage_bytes": {String_: toStrPtr("49536401408")},
										"hugepages_1Gi_bytes":     {String_: toStrPtr("0")},
										"hugepages_2Mi_bytes":     {String_: toStrPtr("0")},
										"memory":                  {String_: toStrPtr("125817904Ki")},
										"pods":                    {String_: toStrPtr("110")},
									},
									Allocatable: map[string]*resource.Quantity{
										"cpu":                     {String_: toStrPtr("16")},
										"ephemeral_storage_bytes": {String_: toStrPtr("44582761194")},
										"hugepages_1Gi_bytes":     {String_: toStrPtr("0")},
										"hugepages_2Mi_bytes":     {String_: toStrPtr("0")},
										"memory":                  {String_: toStrPtr("125715504Ki")},
										"pods":                    {String_: toStrPtr("110")},
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
							"capacity_cpu_cores":       int64(16),
							"capacity_memory_bytes":    int64(1.28837533696e+11),
							"capacity_pods":            int64(110),
							"allocatable_cpu_cores":    int64(16),
							"allocatable_memory_bytes": int64(1.28732676096e+11),
							"allocatable_pods":         int64(110),
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
		for _, node := range ((v.handler.responseMap["/nodes/"]).(*v1.NodeList)).Items {
			err := ks.gatherNode(*node, acc)
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
