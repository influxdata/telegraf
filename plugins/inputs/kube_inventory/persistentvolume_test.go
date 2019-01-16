package kube_inventory

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"

	"github.com/influxdata/telegraf/testutil"
)

func TestPersistentVolume(t *testing.T) {
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
			name: "no pv",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/persistentvolumes/": &v1.PersistentVolumeList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect pvs",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/persistentvolumes/": &v1.PersistentVolumeList{
						Items: []*v1.PersistentVolume{
							{
								Status: &v1.PersistentVolumeStatus{
									Phase: toStrPtr("pending"),
								},
								Spec: &v1.PersistentVolumeSpec{
									StorageClassName: toStrPtr("ebs-1"),
								},
								Metadata: &metav1.ObjectMeta{
									Name: toStrPtr("pv1"),
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
							"phase_type": 2,
						},
						Tags: map[string]string{
							"pv_name":      "pv1",
							"storageclass": "ebs-1",
							"phase":        "pending",
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
		for _, pv := range ((v.handler.responseMap["/persistentvolumes/"]).(*v1.PersistentVolumeList)).Items {
			err := ks.gatherPersistentVolume(*pv, acc)
			if err != nil {
				t.Errorf("Failed to gather pv - %s", err.Error())
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
