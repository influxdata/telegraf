package kube_inventory

import (
	"reflect"
	"strings"
	"testing"
	"time"

	v1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ericchiang/k8s/apis/resource"
	"github.com/influxdata/telegraf/testutil"
)

func TestPod(t *testing.T) {
	cli := &client{}
	selectInclude := []string{}
	selectExclude := []string{}
	now := time.Now()
	started := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 1, 36, 0, now.Location())
	created := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-2, 1, 36, 0, now.Location())
	cond1 := time.Date(now.Year(), 7, 5, 7, 53, 29, 0, now.Location())
	cond2 := time.Date(now.Year(), 7, 5, 7, 53, 31, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   *testutil.Accumulator
		hasError bool
	}{
		{
			name: "no pods",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/pods/": &v1.PodList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect pods",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/pods/": &v1.PodList{
						Items: []*v1.Pod{
							{
								Spec: &v1.PodSpec{
									NodeName: toStrPtr("node1"),
									Containers: []*v1.Container{
										{
											Name:  toStrPtr("running"),
											Image: toStrPtr("image1"),
											Ports: []*v1.ContainerPort{
												{
													ContainerPort: toInt32Ptr(8080),
													Protocol:      toStrPtr("TCP"),
												},
											},
											Resources: &v1.ResourceRequirements{
												Limits: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
												Requests: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
											},
										},
										{
											Name:  toStrPtr("completed"),
											Image: toStrPtr("image1"),
											Ports: []*v1.ContainerPort{
												{
													ContainerPort: toInt32Ptr(8080),
													Protocol:      toStrPtr("TCP"),
												},
											},
											Resources: &v1.ResourceRequirements{
												Limits: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
												Requests: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
											},
										},
										{
											Name:  toStrPtr("waiting"),
											Image: toStrPtr("image1"),
											Ports: []*v1.ContainerPort{
												{
													ContainerPort: toInt32Ptr(8080),
													Protocol:      toStrPtr("TCP"),
												},
											},
											Resources: &v1.ResourceRequirements{
												Limits: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
												Requests: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
											},
										},
									},
									Volumes: []*v1.Volume{
										{
											Name: toStrPtr("vol1"),
											VolumeSource: &v1.VolumeSource{
												PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
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
								Status: &v1.PodStatus{
									Phase:     toStrPtr("Running"),
									HostIP:    toStrPtr("180.12.10.18"),
									PodIP:     toStrPtr("10.244.2.15"),
									StartTime: &metav1.Time{Seconds: toInt64Ptr(started.Unix())},
									Conditions: []*v1.PodCondition{
										{
											Type:               toStrPtr("Initialized"),
											Status:             toStrPtr("True"),
											LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(cond1.Unix())},
										},
										{
											Type:               toStrPtr("Ready"),
											Status:             toStrPtr("True"),
											LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(cond2.Unix())},
										},
										{
											Type:               toStrPtr("Scheduled"),
											Status:             toStrPtr("True"),
											LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(cond1.Unix())},
										},
									},
									ContainerStatuses: []*v1.ContainerStatus{
										{
											Name: toStrPtr("running"),
											State: &v1.ContainerState{
												Running: &v1.ContainerStateRunning{
													StartedAt: &metav1.Time{Seconds: toInt64Ptr(cond2.Unix())},
												},
											},
											Ready:        toBoolPtr(true),
											RestartCount: toInt32Ptr(3),
											Image:        toStrPtr("image1"),
											ImageID:      toStrPtr("image_id1"),
											ContainerID:  toStrPtr("docker://54abe32d0094479d3d"),
										},
										{
											Name: toStrPtr("completed"),
											State: &v1.ContainerState{
												Terminated: &v1.ContainerStateTerminated{
													StartedAt: &metav1.Time{Seconds: toInt64Ptr(cond2.Unix())},
													ExitCode:  toInt32Ptr(0),
													Reason:    toStrPtr("Completed"),
												},
											},
											Ready:        toBoolPtr(false),
											RestartCount: toInt32Ptr(3),
											Image:        toStrPtr("image1"),
											ImageID:      toStrPtr("image_id1"),
											ContainerID:  toStrPtr("docker://54abe32d0094479d3d"),
										},
										{
											Name: toStrPtr("waiting"),
											State: &v1.ContainerState{
												Waiting: &v1.ContainerStateWaiting{
													Reason: toStrPtr("PodUninitialized"),
												},
											},
											Ready:        toBoolPtr(false),
											RestartCount: toInt32Ptr(3),
											Image:        toStrPtr("image1"),
											ImageID:      toStrPtr("image_id1"),
											ContainerID:  toStrPtr("docker://54abe32d0094479d3d"),
										},
									},
								},
								Metadata: &metav1.ObjectMeta{
									OwnerReferences: []*metav1.OwnerReference{
										{
											ApiVersion: toStrPtr("apps/v1"),
											Kind:       toStrPtr("DaemonSet"),
											Name:       toStrPtr("forwarder"),
											Controller: toBoolPtr(true),
										},
									},
									Generation: toInt64Ptr(11232),
									Namespace:  toStrPtr("ns1"),
									Name:       toStrPtr("pod1"),
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(created.Unix())},
								},
							},
						},
					},
				},
			},
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					{
						Measurement: podContainerMeasurement,
						Fields: map[string]interface{}{
							"restarts_total":                   int32(3),
							"state_code":                       0,
							"resource_requests_millicpu_units": int64(100),
							"resource_limits_millicpu_units":   int64(100),
						},
						Tags: map[string]string{
							"namespace":             "ns1",
							"container_name":        "running",
							"node_name":             "node1",
							"pod_name":              "pod1",
							"phase":                 "Running",
							"state":                 "running",
							"readiness":             "ready",
							"node_selector_select1": "s1",
							"node_selector_select2": "s2",
						},
					},
					{
						Measurement: podContainerMeasurement,
						Fields: map[string]interface{}{
							"restarts_total":                   int32(3),
							"state_code":                       1,
							"state_reason":                     "Completed",
							"resource_requests_millicpu_units": int64(100),
							"resource_limits_millicpu_units":   int64(100),
						},
						Tags: map[string]string{
							"namespace":      "ns1",
							"container_name": "completed",
							"node_name":      "node1",
							"pod_name":       "pod1",
							"phase":          "Running",
							"state":          "terminated",
							"readiness":      "unready",
						},
					},
					{
						Measurement: podContainerMeasurement,
						Fields: map[string]interface{}{
							"restarts_total":                   int32(3),
							"state_code":                       2,
							"state_reason":                     "PodUninitialized",
							"resource_requests_millicpu_units": int64(100),
							"resource_limits_millicpu_units":   int64(100),
						},
						Tags: map[string]string{
							"namespace":      "ns1",
							"container_name": "waiting",
							"node_name":      "node1",
							"pod_name":       "pod1",
							"phase":          "Running",
							"state":          "waiting",
							"readiness":      "unready",
						},
					},
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
		for _, pod := range ((v.handler.responseMap["/pods/"]).(*v1.PodList)).Items {
			err := ks.gatherPod(*pod, acc)
			if err != nil {
				t.Errorf("Failed to gather pod - %s", err.Error())
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
						t.Fatalf("%s: tag %s metrics unmatch Expected %s, got %s, i %d\n", v.name, k, m, acc.Metrics[i].Tags[k], i)
					}
				}
				for k, m := range v.output.Metrics[i].Fields {
					if acc.Metrics[i].Fields[k] != m {
						t.Fatalf("%s: field %s metrics unmatch Expected %v(%T), got %v(%T), i %d\n", v.name, k, m, m, acc.Metrics[i].Fields[k], acc.Metrics[i].Fields[k], i)
					}
				}
			}
		}
	}
}

func TestPodSelectorFilter(t *testing.T) {
	cli := &client{}
	now := time.Now()
	started := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 1, 36, 0, now.Location())
	created := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-2, 1, 36, 0, now.Location())
	cond1 := time.Date(now.Year(), 7, 5, 7, 53, 29, 0, now.Location())
	cond2 := time.Date(now.Year(), 7, 5, 7, 53, 31, 0, now.Location())

	responseMap := map[string]interface{}{
		"/pods/": &v1.PodList{
			Items: []*v1.Pod{
				{
					Spec: &v1.PodSpec{
						NodeName: toStrPtr("node1"),
						Containers: []*v1.Container{
							{
								Name:  toStrPtr("forwarder"),
								Image: toStrPtr("image1"),
								Ports: []*v1.ContainerPort{
									{
										ContainerPort: toInt32Ptr(8080),
										Protocol:      toStrPtr("TCP"),
									},
								},
								Resources: &v1.ResourceRequirements{
									Limits: map[string]*resource.Quantity{
										"cpu": {String_: toStrPtr("100m")},
									},
									Requests: map[string]*resource.Quantity{
										"cpu": {String_: toStrPtr("100m")},
									},
								},
							},
						},
						Volumes: []*v1.Volume{
							{
								Name: toStrPtr("vol1"),
								VolumeSource: &v1.VolumeSource{
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
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
					Status: &v1.PodStatus{
						Phase:     toStrPtr("Running"),
						HostIP:    toStrPtr("180.12.10.18"),
						PodIP:     toStrPtr("10.244.2.15"),
						StartTime: &metav1.Time{Seconds: toInt64Ptr(started.Unix())},
						Conditions: []*v1.PodCondition{
							{
								Type:               toStrPtr("Initialized"),
								Status:             toStrPtr("True"),
								LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(cond1.Unix())},
							},
							{
								Type:               toStrPtr("Ready"),
								Status:             toStrPtr("True"),
								LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(cond2.Unix())},
							},
							{
								Type:               toStrPtr("Scheduled"),
								Status:             toStrPtr("True"),
								LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(cond1.Unix())},
							},
						},
						ContainerStatuses: []*v1.ContainerStatus{
							{
								Name: toStrPtr("forwarder"),
								State: &v1.ContainerState{
									Running: &v1.ContainerStateRunning{
										StartedAt: &metav1.Time{Seconds: toInt64Ptr(cond2.Unix())},
									},
								},
								Ready:        toBoolPtr(true),
								RestartCount: toInt32Ptr(3),
								Image:        toStrPtr("image1"),
								ImageID:      toStrPtr("image_id1"),
								ContainerID:  toStrPtr("docker://54abe32d0094479d3d"),
							},
						},
					},
					Metadata: &metav1.ObjectMeta{
						OwnerReferences: []*metav1.OwnerReference{
							{
								ApiVersion: toStrPtr("apps/v1"),
								Kind:       toStrPtr("DaemonSet"),
								Name:       toStrPtr("forwarder"),
								Controller: toBoolPtr(true),
							},
						},
						Generation: toInt64Ptr(11232),
						Namespace:  toStrPtr("ns1"),
						Name:       toStrPtr("pod1"),
						Labels: map[string]string{
							"lab1": "v1",
							"lab2": "v2",
						},
						CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(created.Unix())},
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
				"node_selector_select1": "s1",
				"node_selector_select2": "s2",
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
				"node_selector_select1": "s1",
				"node_selector_select2": "s2",
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
				"node_selector_select1": "s1",
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
				"node_selector_select1": "s1",
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
				"node_selector_select1": "s1",
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
				"node_selector_select1": "s1",
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
				"node_selector_select1": "s1",
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
		for _, pod := range ((v.handler.responseMap["/pods/"]).(*v1.PodList)).Items {
			err := ks.gatherPod(*pod, acc)
			if err != nil {
				t.Errorf("Failed to gather pod - %s", err.Error())
			}
		}

		// Grab selector tags
		actual := map[string]string{}
		for _, metric := range acc.Metrics {
			for key, val := range metric.Tags {
				if strings.Contains(key, "node_selector_") {
					actual[key] = val
				}
			}
		}

		if !reflect.DeepEqual(v.expected, actual) {
			t.Fatalf("actual selector tags (%v) do not match expected selector tags (%v)", actual, v.expected)
		}
	}
}

func TestPodPendingContainers(t *testing.T) {
	cli := &client{}
	selectInclude := []string{}
	selectExclude := []string{}
	now := time.Now()
	started := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 1, 36, 0, now.Location())
	created := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-2, 1, 36, 0, now.Location())
	cond1 := time.Date(now.Year(), 7, 5, 7, 53, 29, 0, now.Location())
	cond2 := time.Date(now.Year(), 7, 5, 7, 53, 31, 0, now.Location())

	tests := []struct {
		name     string
		handler  *mockHandler
		output   *testutil.Accumulator
		hasError bool
	}{
		{
			name: "collect pods",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/pods/": &v1.PodList{
						Items: []*v1.Pod{
							{
								Spec: &v1.PodSpec{
									NodeName: toStrPtr("node1"),
									Containers: []*v1.Container{
										{
											Name:  toStrPtr("waiting"),
											Image: toStrPtr("image1"),
											Ports: []*v1.ContainerPort{
												{
													ContainerPort: toInt32Ptr(8080),
													Protocol:      toStrPtr("TCP"),
												},
											},
											Resources: &v1.ResourceRequirements{
												Limits: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
												Requests: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
											},
										},
										{
											Name:  toStrPtr("terminated"),
											Image: toStrPtr("image1"),
											Ports: []*v1.ContainerPort{
												{
													ContainerPort: toInt32Ptr(8080),
													Protocol:      toStrPtr("TCP"),
												},
											},
											Resources: &v1.ResourceRequirements{
												Limits: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
												Requests: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("100m")},
												},
											},
										},
									},
									Volumes: []*v1.Volume{
										{
											Name: toStrPtr("vol1"),
											VolumeSource: &v1.VolumeSource{
												PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
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
								Status: &v1.PodStatus{
									Phase:     toStrPtr("Pending"),
									Reason:    toStrPtr("NetworkNotReady"),
									HostIP:    toStrPtr("180.12.10.18"),
									PodIP:     toStrPtr("10.244.2.15"),
									StartTime: &metav1.Time{Seconds: toInt64Ptr(started.Unix())},
									Conditions: []*v1.PodCondition{
										{
											Type:               toStrPtr("Initialized"),
											Status:             toStrPtr("True"),
											LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(cond1.Unix())},
										},
										{
											Type:               toStrPtr("Ready"),
											Status:             toStrPtr("True"),
											LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(cond2.Unix())},
										},
										{
											Type:               toStrPtr("Scheduled"),
											Status:             toStrPtr("True"),
											LastTransitionTime: &metav1.Time{Seconds: toInt64Ptr(cond1.Unix())},
										},
									},
									ContainerStatuses: []*v1.ContainerStatus{},
								},
								Metadata: &metav1.ObjectMeta{
									OwnerReferences: []*metav1.OwnerReference{
										{
											ApiVersion: toStrPtr("apps/v1"),
											Kind:       toStrPtr("DaemonSet"),
											Name:       toStrPtr("forwarder"),
											Controller: toBoolPtr(true),
										},
									},
									Generation: toInt64Ptr(11232),
									Namespace:  toStrPtr("ns1"),
									Name:       toStrPtr("pod1"),
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: &metav1.Time{Seconds: toInt64Ptr(created.Unix())},
								},
							},
						},
					},
				},
			},
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					{
						Measurement: podContainerMeasurement,
						Fields: map[string]interface{}{
							"phase_reason":                     "NetworkNotReady",
							"restarts_total":                   int32(0),
							"state_code":                       3,
							"resource_requests_millicpu_units": int64(100),
							"resource_limits_millicpu_units":   int64(100),
						},
						Tags: map[string]string{
							"namespace":             "ns1",
							"container_name":        "waiting",
							"node_name":             "node1",
							"pod_name":              "pod1",
							"phase":                 "Pending",
							"state":                 "unknown",
							"readiness":             "unready",
							"node_selector_select1": "s1",
							"node_selector_select2": "s2",
						},
					},
					{
						Measurement: podContainerMeasurement,
						Fields: map[string]interface{}{
							"phase_reason":                     "NetworkNotReady",
							"restarts_total":                   int32(0),
							"state_code":                       3,
							"resource_requests_millicpu_units": int64(100),
							"resource_limits_millicpu_units":   int64(100),
						},
						Tags: map[string]string{
							"namespace":      "ns1",
							"container_name": "terminated",
							"node_name":      "node1",
							"pod_name":       "pod1",
							"phase":          "Pending",
							"state":          "unknown",
							"readiness":      "unready",
						},
					},
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
		for _, pod := range ((v.handler.responseMap["/pods/"]).(*v1.PodList)).Items {
			err := ks.gatherPod(*pod, acc)
			if err != nil {
				t.Errorf("Failed to gather pod - %s", err.Error())
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
						t.Fatalf("%s: tag %s metrics unmatch Expected %s, got %s, i %d\n", v.name, k, m, acc.Metrics[i].Tags[k], i)
					}
				}
				for k, m := range v.output.Metrics[i].Fields {
					if acc.Metrics[i].Fields[k] != m {
						t.Fatalf("%s: field %s metrics unmatch Expected %v(%T), got %v(%T), i %d\n", v.name, k, m, m, acc.Metrics[i].Fields[k], acc.Metrics[i].Fields[k], i)
					}
				}
			}
		}
	}
}
