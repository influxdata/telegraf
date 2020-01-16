package kube_inventory

import (
	"testing"
	"time"

	v1 "github.com/ericchiang/k8s/apis/core/v1"
	v1beta1EXT "github.com/ericchiang/k8s/apis/extensions/v1beta1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/influxdata/telegraf/testutil"
)

func TestIngress(t *testing.T) {
	cli := &client{}

	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   *testutil.Accumulator
		hasError bool
	}{
		{
			name: "no ingress",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/ingress/": &v1beta1EXT.IngressList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect ingress",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/ingress/": &v1beta1EXT.IngressList{
						Items: []*v1beta1EXT.Ingress{
							{
								Status: &v1beta1EXT.IngressStatus{
									LoadBalancer: &v1.LoadBalancerStatus{
										Ingress: []*v1.LoadBalancerIngress{
											{
												Hostname: toStrPtr("chron-1"),
												Ip:       toStrPtr("1.0.0.127"),
											},
										},
									},
								},
								Spec: &v1beta1EXT.IngressSpec{
									Rules: []*v1beta1EXT.IngressRule{
										{
											Host: toStrPtr("ui.internal"),
											IngressRuleValue: &v1beta1EXT.IngressRuleValue{
												Http: &v1beta1EXT.HTTPIngressRuleValue{
													Paths: []*v1beta1EXT.HTTPIngressPath{
														{
															Path: toStrPtr("/"),
															Backend: &v1beta1EXT.IngressBackend{
																ServiceName: toStrPtr("chronografd"),
																ServicePort: toIntStrPtrI(8080),
															},
														},
													},
												},
											},
										},
									},
								},
								Metadata: &metav1.ObjectMeta{
									Generation:        toInt64Ptr(12),
									Namespace:         toStrPtr("ns1"),
									Name:              toStrPtr("ui-lb"),
									CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(now.Unix())},
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
							"tls":                  false,
							"backend_service_port": int32(8080),
							"generation":           int64(12),
							"created":              now.UnixNano(),
						},
						Tags: map[string]string{
							"ingress_name":         "ui-lb",
							"namespace":            "ns1",
							"ip":                   "1.0.0.127",
							"hostname":             "chron-1",
							"backend_service_name": "chronografd",
							"host":                 "ui.internal",
							"path":                 "/",
						},
					},
				},
			},
			hasError: false,
		},
	}

	for _, v := range tests {
		ks := &KubernetesInventory{
			client: cli,
		}
		acc := new(testutil.Accumulator)
		for _, ingress := range ((v.handler.responseMap["/ingress/"]).(*v1beta1EXT.IngressList)).Items {
			err := ks.gatherIngress(*ingress, acc)
			if err != nil {
				t.Errorf("Failed to gather ingress - %s", err.Error())
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
						t.Fatalf("%s: tag %s metrics unmatch Expected %s, got '%v'\n", v.name, k, m, acc.Metrics[i].Tags[k])
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
