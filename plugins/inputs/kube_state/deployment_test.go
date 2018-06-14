package kube_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestDeployment(t *testing.T) {
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
			name: "no deployments",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/deployments/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "collect deployments",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/deployments/": &v1beta1.DeploymentList{
						Items: []v1beta1.Deployment{
							{
								Status: v1beta1.DeploymentStatus{
									Replicas:            3,
									AvailableReplicas:   1,
									UnavailableReplicas: 4,
									UpdatedReplicas:     2,
									ObservedGeneration:  9121,
								},
								Spec: v1beta1.DeploymentSpec{
									Paused: false,
									Strategy: v1beta1.DeploymentStrategy{
										RollingUpdate: &v1beta1.RollingUpdateDeployment{
											MaxUnavailable: &intstr.IntOrString{
												Type:   intstr.Int,
												IntVal: int32(30),
											},
											MaxSurge: &intstr.IntOrString{
												Type:   intstr.Int,
												IntVal: int32(20),
											},
										},
									},
									Replicas: int32Prt(4),
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation: 11221,
									Namespace:  "ns1",
									Name:       "deploy1",
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
							"spec_replicas":               int32(4),
							"metadata_generation":         int64(11221),
							"status_replicas":             int32(3),
							"status_replicas_available":   int32(1),
							"status_replicas_unavailable": int32(4),
							"status_replicas_updated":     int32(2),
							"status_observed_generation":  int64(9121),
							"created":                     now.Unix(),
							"spec_strategy_rollingupdate_max_unavailable": 30,
							"spec_strategy_rollingupdate_max_surge":       20,
						},
						Tags: map[string]string{
							"label_lab1":  "v1",
							"label_lab2":  "v2",
							"namespace":   "ns1",
							"deployment":  "deploy1",
							"spec_paused": "false",
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
		registerDeploymentCollector(context.Background(), acc, ks)
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

func int32Prt(n int32) *int32 {
	return &n
}
