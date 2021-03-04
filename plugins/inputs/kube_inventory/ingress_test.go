package kube_inventory

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
					"/ingress/": netv1.IngressList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect ingress",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/ingress/": netv1.IngressList{
						Items: []netv1.Ingress{
							{
								Status: netv1.IngressStatus{
									LoadBalancer: v1.LoadBalancerStatus{
										Ingress: []v1.LoadBalancerIngress{
											{
												Hostname: "chron-1",
												IP:       "1.0.0.127",
											},
										},
									},
								},
								Spec: netv1.IngressSpec{
									Rules: []netv1.IngressRule{
										{
											Host: "ui.internal",
											IngressRuleValue: netv1.IngressRuleValue{
												HTTP: &netv1.HTTPIngressRuleValue{
													Paths: []netv1.HTTPIngressPath{
														{
															Path: "/",
															Backend: netv1.IngressBackend{
																Service: &netv1.IngressServiceBackend{
																	Name: "chronografd",
																	Port: netv1.ServiceBackendPort{
																		Number: 8080,
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        12,
									Namespace:         "ns1",
									Name:              "ui-lb",
									CreationTimestamp: metav1.Time{time.Now()},
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
		for _, ingress := range ((v.handler.responseMap["/ingress/"]).(netv1.IngressList)).Items {
			err := ks.gatherIngress(ingress, acc)
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
