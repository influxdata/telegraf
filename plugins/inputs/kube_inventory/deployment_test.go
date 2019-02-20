package kube_inventory

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1beta1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ericchiang/k8s/util/intstr"
	"github.com/influxdata/telegraf/testutil"
)

func TestDeployment(t *testing.T) {
	cli := &client{}

	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())
	outputMetric := &testutil.Metric{
		Fields: map[string]interface{}{
			"replicas_available":   int32(1),
			"replicas_unavailable": int32(4),
			"created":              now.UnixNano(),
		},
		Tags: map[string]string{
			"namespace":       "ns1",
			"deployment_name": "deploy1",
		},
	}

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
					"/deployments/": &v1beta1.DeploymentList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect deployments",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/deployments/": &v1beta1.DeploymentList{
						Items: []*v1beta1.Deployment{
							{
								Status: &v1beta1.DeploymentStatus{
									Replicas:            toInt32Ptr(3),
									AvailableReplicas:   toInt32Ptr(1),
									UnavailableReplicas: toInt32Ptr(4),
									UpdatedReplicas:     toInt32Ptr(2),
									ObservedGeneration:  toInt64Ptr(9121),
								},
								Spec: &v1beta1.DeploymentSpec{
									Strategy: &v1beta1.DeploymentStrategy{
										RollingUpdate: &v1beta1.RollingUpdateDeployment{
											MaxUnavailable: &intstr.IntOrString{
												IntVal: toInt32Ptr(30),
											},
											MaxSurge: &intstr.IntOrString{
												IntVal: toInt32Ptr(20),
											},
										},
									},
									Replicas: toInt32Ptr(4),
								},
								Metadata: &metav1.ObjectMeta{
									Generation: toInt64Ptr(11221),
									Namespace:  toStrPtr("ns1"),
									Name:       toStrPtr("deploy1"),
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(now.Unix())},
								},
							},
						},
					},
				},
			},
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					outputMetric,
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
		for _, deployment := range ((v.handler.responseMap["/deployments/"]).(*v1beta1.DeploymentList)).Items {
			err := ks.gatherDeployment(*deployment, acc)
			if err != nil {
				t.Errorf("Failed to gather deployment - %s", err.Error())
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
