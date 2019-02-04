package kube_inventory

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1beta1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"

	"github.com/influxdata/telegraf/testutil"
)

func TestStatefulSet(t *testing.T) {
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
			name: "no statefulsets",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/statefulsets/": &v1beta1.StatefulSetList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect statefulsets",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/statefulsets/": &v1beta1.StatefulSetList{
						Items: []*v1beta1.StatefulSet{
							{
								Status: &v1beta1.StatefulSetStatus{
									Replicas:           toInt32Ptr(2),
									CurrentReplicas:    toInt32Ptr(4),
									ReadyReplicas:      toInt32Ptr(1),
									UpdatedReplicas:    toInt32Ptr(3),
									ObservedGeneration: toInt64Ptr(119),
								},
								Spec: &v1beta1.StatefulSetSpec{
									Replicas: toInt32Ptr(3),
								},
								Metadata: &metav1.ObjectMeta{
									Generation: toInt64Ptr(332),
									Namespace:  toStrPtr("ns1"),
									Name:       toStrPtr("sts1"),
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
							"generation":          int64(332),
							"observed_generation": int64(119),
							"created":             now.UnixNano(),
							"spec_replicas":       int32(3),
							"replicas":            int32(2),
							"replicas_current":    int32(4),
							"replicas_ready":      int32(1),
							"replicas_updated":    int32(3),
						},
						Tags: map[string]string{
							"namespace":        "ns1",
							"statefulset_name": "sts1",
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
		for _, ss := range ((v.handler.responseMap["/statefulsets/"]).(*v1beta1.StatefulSetList)).Items {
			err := ks.gatherStatefulSet(*ss, acc)
			if err != nil {
				t.Errorf("Failed to gather ss - %s", err.Error())
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
