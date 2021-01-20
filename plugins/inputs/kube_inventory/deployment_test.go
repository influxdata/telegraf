package kube_inventory

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/apps/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ericchiang/k8s/util/intstr"
	"github.com/influxdata/telegraf/testutil"
)

func TestDeployment(t *testing.T) {
	cli := &client{}
	selectInclude := []string{}
	selectExclude := []string{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())
	outputMetric := &testutil.Metric{
		Fields: map[string]interface{}{
			"replicas_available":   int32(1),
			"replicas_unavailable": int32(4),
			"created":              now.UnixNano(),
		},
		Tags: map[string]string{
			"namespace":        "ns1",
			"deployment_name":  "deploy1",
			"selector_select1": "s1",
			"selector_select2": "s2",
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
					"/deployments/": &v1.DeploymentList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect deployments",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/deployments/": &v1.DeploymentList{
						Items: []*v1.Deployment{
							{
								Status: &v1.DeploymentStatus{
									Replicas:            toInt32Ptr(3),
									AvailableReplicas:   toInt32Ptr(1),
									UnavailableReplicas: toInt32Ptr(4),
									UpdatedReplicas:     toInt32Ptr(2),
									ObservedGeneration:  toInt64Ptr(9121),
								},
								Spec: &v1.DeploymentSpec{
									Strategy: &v1.DeploymentStrategy{
										RollingUpdate: &v1.RollingUpdateDeployment{
											MaxUnavailable: &intstr.IntOrString{
												IntVal: toInt32Ptr(30),
											},
											MaxSurge: &intstr.IntOrString{
												IntVal: toInt32Ptr(20),
											},
										},
									},
									Replicas: toInt32Ptr(4),
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"select1": "s1",
											"select2": "s2",
										},
									},
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
			client:          cli,
			SelectorInclude: selectInclude,
			SelectorExclude: selectExclude,
		}
		ks.createSelectorFilters()
		acc := new(testutil.Accumulator)
		for _, deployment := range ((v.handler.responseMap["/deployments/"]).(*v1.DeploymentList)).Items {
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

func TestDeploymentSelectorFilter(t *testing.T) {
	cli := &client{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	responseMap := map[string]interface{}{
		"/deployments/": &v1.DeploymentList{
			Items: []*v1.Deployment{
				{
					Status: &v1.DeploymentStatus{
						Replicas:            toInt32Ptr(3),
						AvailableReplicas:   toInt32Ptr(1),
						UnavailableReplicas: toInt32Ptr(4),
						UpdatedReplicas:     toInt32Ptr(2),
						ObservedGeneration:  toInt64Ptr(9121),
					},
					Spec: &v1.DeploymentSpec{
						Strategy: &v1.DeploymentStrategy{
							RollingUpdate: &v1.RollingUpdateDeployment{
								MaxUnavailable: &intstr.IntOrString{
									IntVal: toInt32Ptr(30),
								},
								MaxSurge: &intstr.IntOrString{
									IntVal: toInt32Ptr(20),
								},
							},
						},
						Replicas: toInt32Ptr(4),
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"select1": "s1",
								"select2": "s2",
							},
						},
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
	}

	tests := []struct {
		name     string
		handler  *mockHandler
		hasError bool
		include  []string
		exclude  []string
		expected map[string]string
	}{
		{
			name: "nil filters equals all selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  nil,
			exclude:  nil,
			expected: map[string]string{
				"selector_select1": "s1",
				"selector_select2": "s2",
			},
		},
		{
			name: "empty filters equals all selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{},
			exclude:  []string{},
			expected: map[string]string{
				"selector_select1": "s1",
				"selector_select2": "s2",
			},
		},
		{
			name: "include filter equals only include-matched selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{"select1"},
			exclude:  []string{},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
		{
			name: "exclude filter equals only non-excluded selectors (overrides include filter)",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{},
			exclude:  []string{"select2"},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
		{
			name: "include glob filter equals only include-matched selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{"*1"},
			exclude:  []string{},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
		{
			name: "exclude glob filter equals only non-excluded selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{},
			exclude:  []string{"*2"},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
		{
			name: "exclude glob filter equals only non-excluded selectors",
			handler: &mockHandler{
				responseMap: responseMap,
			},
			hasError: false,
			include:  []string{},
			exclude:  []string{"*2"},
			expected: map[string]string{
				"selector_select1": "s1",
			},
		},
	}
	for _, v := range tests {
		ks := &KubernetesInventory{
			client: cli,
		}
		ks.SelectorInclude = v.include
		ks.SelectorExclude = v.exclude
		ks.createSelectorFilters()
		acc := new(testutil.Accumulator)
		for _, deployment := range ((v.handler.responseMap["/deployments/"]).(*v1.DeploymentList)).Items {
			err := ks.gatherDeployment(*deployment, acc)
			if err != nil {
				t.Errorf("Failed to gather deployment - %s", err.Error())
			}
		}

		// Grab selector tags
		actual := map[string]string{}
		for _, metric := range acc.Metrics {
			for key, val := range metric.Tags {
				if strings.Contains(key, "selector_") {
					actual[key] = val
				}
			}
		}

		if !reflect.DeepEqual(v.expected, actual) {
			t.Fatalf("actual selector tags (%v) do not match expected selector tags (%v)", actual, v.expected)
		}
	}
}
