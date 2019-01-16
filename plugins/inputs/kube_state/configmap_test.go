package kube_state

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"

	"github.com/influxdata/telegraf/testutil"
)

func TestConfigMap(t *testing.T) {
	cli := &client{}
	oldtime := metav1.Time{
		Seconds: toInt64Ptr(1094505756),
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
					"/configmaps/": &v1.ConfigMapList{},
				},
			},
			hasError: false,
		},
		{
			name: "no creation time",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/configmaps/": &v1.ConfigMapList{
						Items: nil,
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
						Items: []*v1.ConfigMap{
							{
								Metadata: &metav1.ObjectMeta{CreationTimestamp: &oldtime},
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
						Items: []*v1.ConfigMap{
							{
								Metadata: &metav1.ObjectMeta{
									CreationTimestamp: &oldtime,
									Name:              toStrPtr("name1"),
									Namespace:         toStrPtr("ns1"),
									ResourceVersion:   toStrPtr("rv1"),
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
						Fields: map[string]interface{}{
							"created": time.Unix(oldtime.GetSeconds(), int64(oldtime.GetNanos())).UnixNano(),
						},
						Tags: map[string]string{
							"configmap_name":   "name1",
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
						Items: []*v1.ConfigMap{
							{
								Metadata: &metav1.ObjectMeta{
									CreationTimestamp: &oldtime,
									Name:              toStrPtr("name1"),
									Namespace:         toStrPtr("ns1"),
									ResourceVersion:   toStrPtr("rv1"),
								},
							},
							{
								Metadata: &metav1.ObjectMeta{
									CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(oldtime.GetSeconds() + 3600)},
									Name:              toStrPtr("name2"),
									Namespace:         toStrPtr("ns2"),
									ResourceVersion:   toStrPtr("rv2"),
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
						Fields: map[string]interface{}{
							"created": time.Unix(oldtime.GetSeconds(), int64(oldtime.GetNanos())).UnixNano(),
						},
						Tags: map[string]string{
							"configmap_name":   "name1",
							"namespace":        "ns1",
							"resource_version": "rv1",
						},
					},
					{
						Measurement: configMapMeasurement,
						Fields: map[string]interface{}{
							"created": time.Unix(oldtime.GetSeconds()+3600, int64(oldtime.GetNanos())).UnixNano(),
						},
						Tags: map[string]string{
							"configmap_name":   "name2",
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
		ks := &KubernetesState{
			client:          cli,
			firstTimeGather: v.firstGather,
		}
		acc := new(testutil.Accumulator)

		for _, cmap := range ((v.handler.responseMap["/configmaps/"]).(*v1.ConfigMapList)).Items {
			err := ks.gatherConfigMap(*cmap, acc)
			if err != nil {
				t.Errorf("Failed to gather config map - %s", err.Error())
			}
		}

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
