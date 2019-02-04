package kube_inventory

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1beta2"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"

	"github.com/influxdata/telegraf/testutil"
)

func TestDaemonSet(t *testing.T) {
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
			name: "no daemon set",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/daemonsets/": &v1beta2.DaemonSetList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect daemonsets",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/daemonsets/": &v1beta2.DaemonSetList{
						Items: []*v1beta2.DaemonSet{
							{
								Status: &v1beta2.DaemonSetStatus{
									CurrentNumberScheduled: toInt32Ptr(3),
									DesiredNumberScheduled: toInt32Ptr(5),
									NumberAvailable:        toInt32Ptr(2),
									NumberMisscheduled:     toInt32Ptr(2),
									NumberReady:            toInt32Ptr(1),
									NumberUnavailable:      toInt32Ptr(1),
									UpdatedNumberScheduled: toInt32Ptr(2),
								},
								Metadata: &metav1.ObjectMeta{
									Generation: toInt64Ptr(11221),
									Namespace:  toStrPtr("ns1"),
									Name:       toStrPtr("daemon1"),
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
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
							"generation":               int64(11221),
							"current_number_scheduled": int32(3),
							"desired_number_scheduled": int32(5),
							"number_available":         int32(2),
							"number_misscheduled":      int32(2),
							"number_ready":             int32(1),
							"number_unavailable":       int32(1),
							"updated_number_scheduled": int32(2),
							"created":                  now.UnixNano(),
						},
						Tags: map[string]string{
							"daemonset_name": "daemon1",
							"namespace":      "ns1",
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
		for _, dset := range ((v.handler.responseMap["/daemonsets/"]).(*v1beta2.DaemonSetList)).Items {
			err := ks.gatherDaemonSet(*dset, acc)
			if err != nil {
				t.Errorf("Failed to gather daemonset - %s", err.Error())
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
						t.Fatalf("%s: tag %s metrics unmatch Expected %s, got %s\n", v.name, k, m, acc.Metrics[i].Tags[k])
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
