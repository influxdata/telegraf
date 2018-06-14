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

func TestDaemonSet(t *testing.T) {
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
			name: "no daemon set",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/daemonsets/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "collect daemonsets",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/daemonsets/": &v1beta2.DaemonSetList{
						Items: []v1beta2.DaemonSet{
							{
								Status: v1beta2.DaemonSetStatus{
									CurrentNumberScheduled: 3,
									DesiredNumberScheduled: 5,
									NumberAvailable:        2,
									NumberMisscheduled:     2,
									NumberReady:            1,
									NumberUnavailable:      1,
									UpdatedNumberScheduled: 2,
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation: 11221,
									Namespace:  "ns1",
									Name:       "daemon1",
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
						Fields: map[string]interface{}{
							"metadata_generation":             int64(11221),
							"status_current_number_scheduled": int32(3),
							"status_desired_number_scheduled": int32(5),
							"status_number_available":         int32(2),
							"status_number_misscheduled":      int32(2),
							"status_number_ready":             int32(1),
							"status_number_unavailable":       int32(1),
							"status_updated_number_scheduled": int32(2),
							"created":                         now.Unix(),
						},
						Tags: map[string]string{
							"label_lab1": "v1",
							"label_lab2": "v2",
							"namespace":  "ns1",
							"daemonset":  "daemon1",
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
		registerDaemonSetCollector(context.Background(), acc, ks)
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
