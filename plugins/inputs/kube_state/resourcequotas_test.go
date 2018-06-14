package kube_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/influxdata/telegraf/testutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResourceQuotas(t *testing.T) {
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
			name: "no resource quotas",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/resourcequotas/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "collect resource quotas",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/resourcequotas/": &v1.ResourceQuotaList{
						Items: []v1.ResourceQuota{
							{
								Status: v1.ResourceQuotaStatus{
									Hard: v1.ResourceList{
										v1.ResourceCPU:              resource.MustParse("13"),
										v1.ResourceEphemeralStorage: resource.MustParse("48375392Ki"),
									},
									Used: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("50Mi"),
										v1.ResourcePods:   resource.MustParse("21"),
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        int64(21321),
									Namespace:         "ns1",
									Name:              "rq1",
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
							"gauge": int64(13),
						},
						Tags: map[string]string{
							"namespace":     "ns1",
							"resourcequota": "rq1",
							"resource":      "cpu",
							"type":          "hard",
						},
					},
					{
						Fields: map[string]interface{}{
							"gauge": int64(49536401408),
						},
						Tags: map[string]string{
							"namespace":     "ns1",
							"resourcequota": "rq1",
							"resource":      "ephemeral-storage",
							"type":          "hard",
						},
					},
					{
						Fields: map[string]interface{}{
							"gauge": int64(52428800),
						},
						Tags: map[string]string{
							"namespace":     "ns1",
							"resourcequota": "rq1",
							"resource":      "memory",
							"type":          "used",
						},
					},
					{
						Fields: map[string]interface{}{
							"gauge": int64(21),
						},
						Tags: map[string]string{
							"namespace":     "ns1",
							"resourcequota": "rq1",
							"resource":      "pods",
							"type":          "used",
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
		registerResourceQuotaCollector(context.Background(), acc, ks)
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
				resource := v.output.Metrics[i].Tags["resource"]
				var j int
				for j = range acc.Metrics {
					if acc.Metrics[j].Tags["resource"] == resource {
						break
					}
				}
				for k, m := range v.output.Metrics[i].Tags {
					if acc.Metrics[j].Tags[k] != m {
						t.Fatalf("%s: tag %s metrics unmatch Expected %s, got %s\n", v.name, k, m, acc.Metrics[j].Tags[k])
					}
				}
				for k, m := range v.output.Metrics[i].Fields {
					if acc.Metrics[j].Fields[k] != m {
						t.Fatalf("%s: field %s metrics unmatch Expected %v(%T), got %v(%T)\n", v.name, k, m, m, acc.Metrics[j].Fields[k], acc.Metrics[i].Fields[k])
					}
				}
			}
		}

	}
}
