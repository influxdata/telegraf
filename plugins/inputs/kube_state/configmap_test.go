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

func TestConfigMap(t *testing.T) {
	cli := &client{
		httpClient: &http.Client{Transport: &http.Transport{}},
		semaphore:  make(chan struct{}, 1),
	}
	oldtime := metav1.Time{
		Time: time.Unix(1094505756, 0),
	}
	tests := []struct {
		name        string
		handler     *mockHandler
		output      *testutil.Accumulator
		firstGather bool
		hasError    bool
	}{
		{
			name: "no config map",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/configmaps/": &v1.ServiceStatus{},
				},
			},
			hasError: false,
		},
		{
			name: "no creation time",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/configmaps/": &v1.ConfigMapList{
						Items: []v1.ConfigMap{
							{},
						},
					},
				},
			},
			output:   &testutil.Accumulator{},
			hasError: false,
		},
		{
			name: "old config map",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/configmaps/": &v1.ConfigMapList{
						Items: []v1.ConfigMap{
							{
								ObjectMeta: metav1.ObjectMeta{CreationTimestamp: oldtime},
							},
						},
					},
				},
			},
			output:   &testutil.Accumulator{},
			hasError: false,
		},
		{
			name: "old config map first gather",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/configmaps/": &v1.ConfigMapList{
						Items: []v1.ConfigMap{
							{
								ObjectMeta: metav1.ObjectMeta{
									CreationTimestamp: oldtime,
									Name:              "name1",
									Namespace:         "ns1",
									ResourceVersion:   "rv1",
								},
							},
						},
					},
				},
			},
			firstGather: true,
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					{
						Time: oldtime.Time,
						Fields: map[string]interface{}{
							"gauge": 1,
						},
						Tags: map[string]string{
							"configmap":        "name1",
							"namespace":        "ns1",
							"resource_version": "rv1",
						},
					},
				},
			},
			hasError: false,
		},
		{
			name: "multiple config map",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/configmaps/": &v1.ConfigMapList{
						Items: []v1.ConfigMap{
							{
								ObjectMeta: metav1.ObjectMeta{
									CreationTimestamp: oldtime,
									Name:              "name1",
									Namespace:         "ns1",
									ResourceVersion:   "rv1",
								},
							},
							{
								ObjectMeta: metav1.ObjectMeta{
									CreationTimestamp: metav1.Time{Time: oldtime.Time.Add(time.Hour)},
									Name:              "name2",
									Namespace:         "ns2",
									ResourceVersion:   "rv2",
								},
							},
						},
					},
				},
			},
			firstGather: true,
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					{
						Measurement: configMapMeasurement,
						Time:        oldtime.Time,
						Fields: map[string]interface{}{
							"gauge": 1,
						},
						Tags: map[string]string{
							"configmap":        "name1",
							"namespace":        "ns1",
							"resource_version": "rv1",
						},
					},
					{
						Measurement: configMapMeasurement,
						Time:        oldtime.Time.Add(time.Hour),
						Fields: map[string]interface{}{
							"gauge": 1,
						},
						Tags: map[string]string{
							"configmap":        "name2",
							"namespace":        "ns2",
							"resource_version": "rv2",
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
			client:          cli,
			firstTimeGather: v.firstGather,
		}
		acc := new(testutil.Accumulator)
		registerConfigMapCollector(context.Background(), acc, ks)
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
