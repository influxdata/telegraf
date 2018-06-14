package kube_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLimitRange(t *testing.T) {
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
			name: "no limitranges",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/limitranges/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "collect limitranges",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/limitranges/": &v1.LimitRangeList{
						Items: []v1.LimitRange{
							{
								Spec: v1.LimitRangeSpec{
									Limits: []v1.LimitRangeItem{
										{
											Type: v1.LimitTypeContainer,
											Max: v1.ResourceList{
												v1.ResourceMemory: resource.MustParse("50Mi"),
											},
										},
										{
											Type: v1.LimitTypePod,
											Min: v1.ResourceList{
												v1.ResourceMemory: resource.MustParse("1250Mi"),
											},
											Default: v1.ResourceList{
												v1.ResourceCPU: resource.MustParse("10"),
											},
										},
										{
											Type: v1.LimitTypePersistentVolumeClaim,
											DefaultRequest: v1.ResourceList{
												v1.ResourceCPU: resource.MustParse("2"),
											},
											MaxLimitRequestRatio: v1.ResourceList{
												v1.ResourceEphemeralStorage: resource.MustParse("15Gi"),
											},
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        11221,
									Namespace:         "ns1",
									Name:              "lm1",
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
							"created":                                                         now.Unix(),
							"max_container_memory":                                            int64(52428800),
							"min_pod_memory":                                                  int64(1310720000),
							"default_request_persistentvolumeclaim_cpu":                       int64(2),
							"max_limit_request_ratio_persistentvolumeclaim_ephemeral_storage": int64(16106127360),
						},
						Tags: map[string]string{
							"namespace":  "ns1",
							"limitrange": "lm1",
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
		registerLimitRangeCollector(context.Background(), acc, ks)
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
