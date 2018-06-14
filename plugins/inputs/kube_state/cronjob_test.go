package kube_state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCronJob(t *testing.T) {
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
			name: "no cronjobs",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/cronjobs/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "suspended cronjobs",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/cronjobs/": &batchv1beta1.CronJobList{
						Items: []batchv1beta1.CronJob{
							{Spec: batchv1beta1.CronJobSpec{Suspend: boolPtr(true)}},
						},
					},
				},
			},
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{},
			},
			hasError: false,
		},
		{
			name: "collect cronjob",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/cronjobs/": &batchv1beta1.CronJobList{
						Items: []batchv1beta1.CronJob{
							{
								Status: batchv1beta1.CronJobStatus{
									Active: []v1.ObjectReference{
										{}, {},
									},
									LastScheduleTime: &metav1.Time{Time: now},
								},
								Spec: batchv1beta1.CronJobSpec{
									Schedule:                "@every 1h30m",
									ConcurrencyPolicy:       batchv1beta1.AllowConcurrent,
									StartingDeadlineSeconds: int64Ptr(10),
								},
								ObjectMeta: metav1.ObjectMeta{
									Namespace: "ns1",
									Name:      "cron1",
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
							"status_active":                  2,
							"schedule":                       "@every 1h30m",
							"spec_starting_deadline_seconds": int64(10),
							"next_schedule_time":             36,
							"created":                        now.Unix(),
						},
						Tags: map[string]string{
							"label_lab1":         "v1",
							"label_lab2":         "v2",
							"namespace":          "ns1",
							"cronjob":            "cron1",
							"concurrency_policy": "allow",
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
		registerCronJobCollector(context.Background(), acc, ks)
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

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}
