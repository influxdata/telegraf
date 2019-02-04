package kube_inventory

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"

	"github.com/influxdata/telegraf/testutil"
)

func TestPersistentVolumeClaim(t *testing.T) {
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
			name: "no pv claims",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/persistentvolumeclaims/": &v1.PersistentVolumeClaimList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect pv claims",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/persistentvolumeclaims/": &v1.PersistentVolumeClaimList{
						Items: []*v1.PersistentVolumeClaim{
							{
								Status: &v1.PersistentVolumeClaimStatus{
									Phase: toStrPtr("bound"),
								},
								Spec: &v1.PersistentVolumeClaimSpec{
									VolumeName:       toStrPtr("pvc-dc870fd6-1e08-11e8-b226-02aa4bc06eb8"),
									StorageClassName: toStrPtr("ebs-1"),
								},
								Metadata: &metav1.ObjectMeta{
									Namespace: toStrPtr("ns1"),
									Name:      toStrPtr("pc1"),
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
							"phase_type": 0,
						},
						Tags: map[string]string{
							"pvc_name":     "pc1",
							"namespace":    "ns1",
							"storageclass": "ebs-1",
							"phase":        "bound",
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
		for _, pvc := range ((v.handler.responseMap["/persistentvolumeclaims/"]).(*v1.PersistentVolumeClaimList)).Items {
			err := ks.gatherPersistentVolumeClaim(*pvc, acc)
			if err != nil {
				t.Errorf("Failed to gather pvc - %s", err.Error())
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
