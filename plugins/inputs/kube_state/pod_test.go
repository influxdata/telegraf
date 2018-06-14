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
						Items: []v1.Pod{
							{
								Spec: v1.PodSpec{
									NodeName: "node1",
									Containers: []v1.Container{
										{
											Name:  "forwarder",
											Image: "image1",
											Ports: []v1.ContainerPort{
												{
													ContainerPort: 8080,
													Protocol:      "TCP",
												},
											},
										},
									},
									Volumes: []v1.Volume{
										{
											Name: "vol1",
											VolumeSource: v1.VolumeSource{
												PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
													ClaimName: "pc1",
													ReadOnly:  true,
												},
											},
										},
										{
											Name: "vol2",
										},
									},
								},
								Status: v1.PodStatus{
									Phase:     v1.PodRunning,
									HostIP:    "180.12.10.18",
									PodIP:     "10.244.2.15",
									StartTime: &metav1.Time{Time: started},
									Conditions: []v1.PodCondition{
										{
											Type:               v1.PodInitialized,
											Status:             v1.ConditionTrue,
											LastTransitionTime: metav1.Time{Time: cond1},
										},
										{
											Type:               v1.PodReady,
											Status:             v1.ConditionTrue,
											LastTransitionTime: metav1.Time{Time: cond2},
										},
										{
											Type:               v1.PodScheduled,
											Status:             v1.ConditionTrue,
											LastTransitionTime: metav1.Time{Time: cond1},
										},
									},
									ContainerStatuses: []v1.ContainerStatus{
										{

											Name: "forwarder",
											State: v1.ContainerState{
												Running: &v1.ContainerStateRunning{
													StartedAt: metav1.Time{Time: cond2},
												},
											},
											Ready:        true,
											RestartCount: int32(3),
											Image:        "image1",
											ImageID:      "image_id1",
											ContainerID:  "docker://54abe32d0094479d3d",
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									OwnerReferences: []metav1.OwnerReference{
										{
											APIVersion: "apps/v1",
											Kind:       "DaemonSet",
											Name:       "forwarder",
											Controller: boolPtr(true),
										},
									},
									Generation: int64(11232),
									Namespace:  "ns1",
									Name:       "pod1",
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: metav1.Time{Time: created},
								},
							},
						},
					},
				},
			},
			output: &testutil.Accumulator{
				Metrics: []*testutil.Metric{
					{
						Measurement: podVolumeMeasurement,
						Fields: map[string]interface{}{
							"read_only": 1,
						},
						Tags: map[string]string{
							"namespace":             "ns1",
							"pod":                   "pod1",
							"volume":                "vol1",
							"persistentvolumeclaim": "pc1",
						},
					},
					{
						Measurement: podMeasurement,
						Fields: map[string]interface{}{
							"gauge": 1,
						},
						Tags: map[string]string{
							"namespace":           "ns1",
							"pod":                 "pod1",
							"node":                "node1",
							"label_lab1":          "v1",
							"label_lab2":          "v2",
							"owner_kind":          "DaemonSet",
							"owner_name":          "forwarder",
							"owner_is_controller": "true",
							"created_by_kind":     "DaemonSet",
							"created_by_name":     "forwarder",
						},
					},
					{
						Measurement: podContainerMeasurement,
						Fields: map[string]interface{}{
							"status_restarts_total": int32(3),
							"status_waiting":        0,
							"status_running":        1,
							"status_terminated":     0,
							"status_ready":          1,
						},
						Tags: map[string]string{
							"namespace":                "ns1",
							"pod_name":                 "pod1",
							"node_name":                "node1",
							"container":                "forwarder",
							"image":                    "image1",
							"image_id":                 "image_id1",
							"container_id":             "docker://54abe32d0094479d3d",
							"status_waiting_reason":    "",
							"status_terminated_reason": "",
						},
					},
					{
						Measurement: podStatusMeasurement,
						Fields: map[string]interface{}{
							"start_time":             started.Unix(),
							"status_phase_pending":   0,
							"status_phase_succeeded": 0,
							"status_phase_failed":    0,
							"status_phase_running":   1,
							"status_phase_unknown":   0,
						},
						Tags: map[string]string{
							"namespace":    "ns1",
							"pod":          "pod1",
							"node":         "node1",
							"host_ip":      "180.12.10.18",
							"pod_ip":       "10.244.2.15",
							"status_phase": "running",
							"ready":        "false",
							"scheduled":    "false",
						},
					},
					{
						Measurement: podStatusMeasurement,
						Fields: map[string]interface{}{
							"start_time":             started.Unix(),
							"status_phase_pending":   0,
							"status_phase_succeeded": 0,
							"status_phase_failed":    0,
							"status_phase_running":   1,
							"status_phase_unknown":   0,
						},
						Tags: map[string]string{
							"namespace":    "ns1",
							"pod":          "pod1",
							"node":         "node1",
							"host_ip":      "180.12.10.18",
							"pod_ip":       "10.244.2.15",
							"status_phase": "running",
							"ready":        "true",
							"scheduled":    "false",
						},
					},
					{
						Measurement: podStatusMeasurement,
						Fields: map[string]interface{}{
							"start_time":             started.Unix(),
							"scheduled_time":         cond1.Unix(),
							"status_phase_pending":   0,
							"status_phase_succeeded": 0,
							"status_phase_failed":    0,
							"status_phase_running":   1,
							"status_phase_unknown":   0,
						},
						Tags: map[string]string{
							"namespace":    "ns1",
							"pod":          "pod1",
							"node":         "node1",
							"host_ip":      "180.12.10.18",
							"pod_ip":       "10.244.2.15",
							"status_phase": "running",
							"ready":        "false",
							"scheduled":    "true",
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
