package kube_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespace(t *testing.T) {
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
			name: "no namespaces",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/namespaces/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "collect namespaces",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/namespaces/": &v1.NamespaceList{
						Items: []v1.Namespace{
							{
								Status: v1.NamespaceStatus{
									Phase: v1.NamespaceActive,
								},
								ObjectMeta: metav1.ObjectMeta{
									Name:              "ns1",
									CreationTimestamp: metav1.Time{Time: now},
									Labels: map[string]string{
										"lab1": "l1",
										"lab2": "l2",
									},
									Annotations: map[string]string{
										"anot1": "a1",
										"anot2": "a2",
									},
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
							"created":           now.Unix(),
							"status_phase_code": 1,
						},
						Tags: map[string]string{
							"status_phase":     "active",
							"namespace":        "ns1",
							"label_lab1":       "l1",
							"label_lab2":       "l2",
							"annotation_anot1": "a1",
							"annotation_anot2": "a2",
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
		registerNamespaceCollector(context.Background(), acc, ks)
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
