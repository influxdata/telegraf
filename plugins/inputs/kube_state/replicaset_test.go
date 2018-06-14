package kube_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReplicaSet(t *testing.T) {
	cli := &client{
		httpClient: &http.Client{Transport: &http.Transport{}},
		semaphore:  make(chan struct{}, 1),
	}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   *testutil.Accumulator
		hasError bool
	}{
		{
			name: "no replicasets",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/replicasets/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "collect replicasets",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/replicasets/": &v1beta2.ReplicaSetList{
						Items: []v1beta2.ReplicaSet{
							{
								Status: v1beta2.ReplicaSetStatus{
									Replicas:             int32(2),
									FullyLabeledReplicas: int32(1),
									ReadyReplicas:        int32(3),
									ObservedGeneration:   int64(5),
								},
								Spec: v1beta2.ReplicaSetSpec{
									Replicas: int32Prt(13),
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        int64(21321),
									Namespace:         "ns1",
									Name:              "rp1",
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
						Fields: map[string]interface{}{
							"metadata_generation":           int64(21321),
							"created":                       now.Unix(),
							"spec_replicas":                 int32(13),
							"status_replicas":               int32(2),
							"status_fully_labeled_replicas": int32(1),
							"status_ready_replicas":         int32(3),
							"status_observed_generation":    int64(5),
						},
						Tags: map[string]string{
							"namespace":  "ns1",
							"replicaset": "rp1",
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
		registerReplicaSetCollector(context.Background(), acc, ks)
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
