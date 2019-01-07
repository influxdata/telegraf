package kube_lite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ericchiang/k8s/apis/resource"
	"github.com/influxdata/telegraf/testutil"
)

func TestPod(t *testing.T) {
	cli := &client{
		httpClient: &http.Client{Transport: &http.Transport{}},
		semaphore:  make(chan struct{}, 1),
	}
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
					"/pods/": &v1.ServiceStatus{},
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
													"cpu": {String_: toStrPtr("8")},
												},
												Requests: map[string]*resource.Quantity{
													"cpu": {String_: toStrPtr("8")},
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
												// Terminated: &v1.ContainerStateTerminated{
												// 	Reason: toStrPtr("completed"),
												// },
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
				},
			},
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					// {
					// 	Measurement: podVolumeMeasurement,
					// 	Fields: map[string]interface{}{
					// 		"read_only": 1,
					// 	},
					// 	Tags: map[string]string{
					// 		"namespace":             "ns1",
					// 		"pod":                   "pod1",
					// 		"volume":                "vol1",
					// 		"persistentvolumeclaim": "pc1",
					// 	},
					// },
					// {
					// 	Measurement: podMeasurement,
					// 	Fields: map[string]interface{}{
					// 		"gauge": 1,
					// 	},
					// 	Tags: map[string]string{
					// 		"namespace":           "ns1",
					// 		"pod":                 "pod1",
					// 		"node":                "node1",
					// 		"label_lab1":          "v1",
					// 		"label_lab2":          "v2",
					// 		"owner_kind":          "DaemonSet",
					// 		"owner_name":          "forwarder",
					// 		"owner_is_controller": "true",
					// 		"created_by_kind":     "DaemonSet",
					// 		"created_by_name":     "forwarder",
					// 	},
					// },
					{
						Measurement: podContainerMeasurement,
						Fields: map[string]interface{}{
							"status_restarts_total": int32(3),
							"status_running":        1,
							"status_terminated":     0,
							// "status_terminated_reason":    "completed",
							"resource_requests_cpu_cores": "8",
							"resource_limits_cpu_cores":   "8",
							// "status_waiting":        0,
							// "status_ready":          1,
						},
						Tags: map[string]string{
							"namespace": "ns1",
							"name":      "forwarder",
							"node":      "node1",
							"pod":       "pod1",
							// "image":                    "image1",
							// "image_id":                 "image_id1",
							// "container_id":             "docker://54abe32d0094479d3d",
							// "status_waiting_reason":    "",
							// "status_terminated_reason": "",
						},
					},
					{
						Measurement: podStatusMeasurement,
						Fields: map[string]interface{}{
							"ready": "false",
						},
						Tags: map[string]string{
							"namespace": "ns1",
							"name":      "pod1",
							"node":      "node1",
						},
					},
					// {
					// 	Measurement: podStatusMeasurement,
					// 	Fields: map[string]interface{}{
					// 		"start_time":             started.Unix(),
					// 		"status_phase_pending":   0,
					// 		"status_phase_succeeded": 0,
					// 		"status_phase_failed":    0,
					// 		"status_phase_running":   1,
					// 		"status_phase_unknown":   0,
					// 		"resource_requests_cpu_cores": "8",
					// 		"resource_limits_cpu_cores":   "8",
					// 	},
					// 	Tags: map[string]string{
					// 		"namespace": "ns1",
					// 		"pod":          "pod1",
					// 		"node":         "node1",
					// 		"host_ip":      "180.12.10.18",
					// 		"pod_ip":       "10.244.2.15",
					// 		"status_phase": "running",
					// 		"ready":        "true",
					// 		"scheduled":    "false",
					// 	},
					// },
					// {
					// 	Measurement: podStatusMeasurement,
					// 	Fields: map[string]interface{}{
					// 		"start_time":             started.Unix(),
					// 		"scheduled_time":         cond1.Unix(),
					// 		"status_phase_pending":   0,
					// 		"status_phase_succeeded": 0,
					// 		"status_phase_failed":    0,
					// 		"status_phase_running":   1,
					// 		"resource_requests_cpu_cores": "8",
					// 		"resource_limits_cpu_cores":   "8",
					// 		"status_phase_unknown":        0,
					// 	},
					// 	Tags: map[string]string{
					// 		"namespace":    "ns1",
					// 		"pod":          "pod1",
					// 		"node":         "node1",
					// 		"host_ip":      "180.12.10.18",
					// 		"pod_ip":       "10.244.2.15",
					// 		"status_phase": "running",
					// 		"ready":        "false",
					// 		"scheduled": "true",
					// 	},
					// },
				},
			},
			hasError: false,
		},
	}
	for _, v := range tests {
		ts := httptest.NewServer(v.handler)
		defer ts.Close()

		cli.baseURL = ts.URL
		ks := &KubernetesState{
			client: cli,
		}
		acc := new(testutil.Accumulator)
		registerPodCollector(context.Background(), acc, ks)
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
