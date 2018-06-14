package kube_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	autoscaling "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHorizontalPodAutoScalerMeasurement(t *testing.T) {
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
			name: "no hpa",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/horizontalpodautoscalers/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "collect hpa",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/horizontalpodautoscalers/": &autoscaling.HorizontalPodAutoscalerList{
						Items: []autoscaling.HorizontalPodAutoscaler{
							{
								Spec: autoscaling.HorizontalPodAutoscalerSpec{
									MaxReplicas: int32(200),
									MinReplicas: int32Prt(1),
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation: int64(11232),
									Namespace:  "ns1",
									Name:       "hpa1",
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: metav1.Time{Time: now},
								},
								Status: autoscaling.HorizontalPodAutoscalerStatus{
									CurrentReplicas: int32(23),
									DesiredReplicas: int32(40),
									Conditions: []autoscaling.HorizontalPodAutoscalerCondition{
										{Status: v1.ConditionTrue, Type: autoscaling.ScalingActive},
										{Status: v1.ConditionUnknown, Type: autoscaling.ScalingLimited},
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
							"current_replicas":  int32(23),
							"desired_replicas":  int32(40),
							"condition_true":    1,
							"condition_false":   0,
							"condition_unknown": 0,
						},
						Tags: map[string]string{
							"namespace": "ns1",
							"hpa":       "hpa1",
							"condition": "scalingactive",
						},
					},
					{
						Fields: map[string]interface{}{
							"current_replicas":  int32(23),
							"desired_replicas":  int32(40),
							"condition_true":    0,
							"condition_false":   0,
							"condition_unknown": 1,
						},
						Tags: map[string]string{
							"namespace": "ns1",
							"hpa":       "hpa1",
							"condition": "scalinglimited",
						},
					},
					{
						Fields: map[string]interface{}{
							"metadata_generation": int64(11232),
							"spec_max_replicas":   int32(200),
							"spec_min_replicas":   int32(1),
						},
						Tags: map[string]string{
							"label_lab1": "v1",
							"label_lab2": "v2",
							"namespace":  "ns1",
							"hpa":        "hpa1",
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
		registerHorizontalPodAutoScalerCollector(context.Background(), acc, ks)
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

func int32Ptr(n int32) *int32 {
	return &n
}
