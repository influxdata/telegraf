package kube_inventory

import (
	"reflect"
	"strings"
	"testing"
	"time"

	batchv1 "github.com/ericchiang/k8s/apis/batch/v1"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ericchiang/k8s/apis/resource"
	"github.com/influxdata/telegraf/testutil"
)

func TestJob(t *testing.T) {
	cli := &client{}
	selectInclude := []string{}
	selectExclude := []string{}
	now := time.Now()
	started := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-2, 1, 36, 0, now.Location())
	completed := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 1, 36, 0, now.Location())
	outputMetric := &testutil.Metric{
		Fields: map[string]interface{}{
			"active":    int32(0),
			"completed": completed.UnixNano(),
			"failed":    int32(0),
			"started":   started.UnixNano(),
			"succeeded": int32(1),
		},
		Tags: map[string]string{
			"namespace":        "ns1",
			"job_name":         "job1",
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
			name: "no jobs",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/jobs/": &batchv1.JobList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect jobs",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/jobs/": &batchv1.JobList{
						Items: []*batchv1.Job{
							{
								Status: &batchv1.JobStatus{
									StartTime:      &metav1.Time{Seconds: toInt64Ptr(started.Unix())},
									CompletionTime: &metav1.Time{Seconds: toInt64Ptr(completed.Unix())},
									Active:         toInt32Ptr(0),
									Succeeded:      toInt32Ptr(1),
									Failed:         toInt32Ptr(0),
								},
								Spec: &batchv1.JobSpec{
									Parallelism:           toInt32Ptr(1),
									Completions:           toInt32Ptr(1),
									ActiveDeadlineSeconds: toInt64Ptr(60),
									BackoffLimit:          toInt32Ptr(5),
									ManualSelector:        toBoolPtr(false),
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"select1": "s1",
											"select2": "s2",
										},
									},
									Template: &corev1.PodTemplateSpec{
										Metadata: &metav1.ObjectMeta{
											Generation: toInt64Ptr(11221),
											Namespace:  toStrPtr("ns1"),
											Name:       toStrPtr("job1"),
											Labels: map[string]string{
												"lab1": "v1",
												"lab2": "v2",
											},
											CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(now.Unix())},
										},
										Spec: &corev1.PodSpec{
											NodeName: toStrPtr("node1"),
											Containers: []*corev1.Container{
												{
													Name:  toStrPtr("forwarder"),
													Image: toStrPtr("image1"),
													Ports: []*corev1.ContainerPort{
														{
															ContainerPort: toInt32Ptr(8080),
															Protocol:      toStrPtr("TCP"),
														},
													},
													Resources: &corev1.ResourceRequirements{
														Limits: map[string]*resource.Quantity{
															"cpu": {String_: toStrPtr("100m")},
														},
														Requests: map[string]*resource.Quantity{
															"cpu": {String_: toStrPtr("100m")},
														},
													},
												},
											},
											Volumes: []*corev1.Volume{
												{
													Name: toStrPtr("vol1"),
													VolumeSource: &corev1.VolumeSource{
														PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
															ClaimName: toStrPtr("pc1"),
															ReadOnly:  toBoolPtr(true),
														},
													},
												},
												{
													Name: toStrPtr("vol2"),
												},
											},
											NodeSelector: map[string]string{
												"select1": "s1",
												"select2": "s2",
											},
										},
									},
								},
								Metadata: &metav1.ObjectMeta{
									Generation: toInt64Ptr(11221),
									Namespace:  toStrPtr("ns1"),
									Name:       toStrPtr("job1"),
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
		for _, job := range ((v.handler.responseMap["/jobs/"]).(*batchv1.JobList)).Items {
			err := ks.gatherJob(*job, acc)
			if err != nil {
				t.Errorf("Failed to gather job - %s", err.Error())
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

func TestJobSelectorFilter(t *testing.T) {
	cli := &client{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())
	started := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-2, 6, 28, 0, now.Location())
	completed := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 5, 37, 0, now.Location())

	responseMap := map[string]interface{}{
		"/jobs/": &batchv1.JobList{
			Items: []*batchv1.Job{
				{
					Status: &batchv1.JobStatus{
						StartTime:      &metav1.Time{Seconds: toInt64Ptr(started.Unix())},
						CompletionTime: &metav1.Time{Seconds: toInt64Ptr(completed.Unix())},
						Active:         toInt32Ptr(5),
						Succeeded:      toInt32Ptr(3),
						Failed:         toInt32Ptr(1),
					},
					Spec: &batchv1.JobSpec{
						Parallelism:           toInt32Ptr(5),
						Completions:           toInt32Ptr(0),
						ActiveDeadlineSeconds: toInt64Ptr(120),
						BackoffLimit:          toInt32Ptr(3),
						ManualSelector:        toBoolPtr(false),
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"select1": "s1",
								"select2": "s2",
							},
						},
						Template: &corev1.PodTemplateSpec{
							Metadata: &metav1.ObjectMeta{
								Generation: toInt64Ptr(1),
								Namespace:  toStrPtr("ns1"),
								Name:       toStrPtr("job1"),
								Labels: map[string]string{
									"lab1": "v1",
									"lab2": "v2",
								},
								CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(now.Unix())},
							},
							Spec: &corev1.PodSpec{
								NodeName: toStrPtr("node1"),
								Containers: []*corev1.Container{
									{
										Name:  toStrPtr("forwarder"),
										Image: toStrPtr("image1"),
										Ports: []*corev1.ContainerPort{
											{
												ContainerPort: toInt32Ptr(8080),
												Protocol:      toStrPtr("TCP"),
											},
										},
										Resources: &corev1.ResourceRequirements{
											Limits: map[string]*resource.Quantity{
												"cpu": {String_: toStrPtr("100m")},
											},
											Requests: map[string]*resource.Quantity{
												"cpu": {String_: toStrPtr("100m")},
											},
										},
									},
								},
								Volumes: []*corev1.Volume{
									{
										Name: toStrPtr("vol1"),
										VolumeSource: &corev1.VolumeSource{
											PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
												ClaimName: toStrPtr("pc1"),
												ReadOnly:  toBoolPtr(true),
											},
										},
									},
									{
										Name: toStrPtr("vol2"),
									},
								},
								NodeSelector: map[string]string{
									"select1": "s1",
									"select2": "s2",
								},
							},
						},
					},
					Metadata: &metav1.ObjectMeta{
						Generation: toInt64Ptr(11221),
						Namespace:  toStrPtr("ns1"),
						Name:       toStrPtr("job1"),
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
		for _, job := range ((v.handler.responseMap["/jobs/"]).(*batchv1.JobList)).Items {
			err := ks.gatherJob(*job, acc)
			if err != nil {
				t.Errorf("Failed to gather job - %s", err.Error())
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
