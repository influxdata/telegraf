package kube_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	v1batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJob(t *testing.T) {
	cli := &client{
		httpClient: &http.Client{Transport: &http.Transport{}},
		semaphore:  make(chan struct{}, 1),
	}
	now := time.Now()
	started := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 1, 36, 0, now.Location())
	created := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-2, 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   *testutil.Accumulator
		hasError bool
	}{
		{
			name: "no job",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/jobs/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "collect jobs",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/jobs/": &v1batch.JobList{
						Items: []v1batch.Job{
							{
								Spec: v1batch.JobSpec{
									Parallelism:           int32Prt(20),
									Completions:           int32Prt(10),
									ActiveDeadlineSeconds: int64Ptr(13),
								},
								Status: v1batch.JobStatus{
									CompletionTime: &metav1.Time{Time: now},
									StartTime:      &metav1.Time{Time: started},
									Succeeded:      int32(5),
									Failed:         int32(3),
									Active:         int32(4),
									Conditions: []v1batch.JobCondition{
										{
											Status:             v1.ConditionTrue,
											Type:               v1batch.JobComplete,
											LastTransitionTime: metav1.Time{Time: created},
										},
										{
											Status:             v1.ConditionFalse,
											Type:               v1batch.JobFailed,
											LastTransitionTime: metav1.Time{Time: started},
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation: int64(11232),
									Namespace:  "ns1",
									Name:       "job1",
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: metav1.Time{Time: created},
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
							"created":                      created.Unix(),
							"status_succeeded":             int32(5),
							"status_failed":                int32(3),
							"status_active":                int32(4),
							"spec_parallelism":             int32(20),
							"spec_completions":             int32(10),
							"spec_active_deadline_seconds": int64(13),
							"status_start_time":            started.Unix(),
						},
						Tags: map[string]string{
							"namespace":  "ns1",
							"job_name":   "job1",
							"label_lab1": "v1",
							"label_lab2": "v2",
						},
					},
					{
						Fields: map[string]interface{}{
							"completed": 1,
							"failed":    0,
						},
						Tags: map[string]string{
							"namespace": "ns1",
							"job_name":  "job1",
							"condition": "true",
						},
					},
					{
						Fields: map[string]interface{}{
							"completed": 0,
							"failed":    1,
						},
						Tags: map[string]string{
							"namespace": "ns1",
							"job_name":  "job1",
							"condition": "false",
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
		registerJobCollector(context.Background(), acc, ks)
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
