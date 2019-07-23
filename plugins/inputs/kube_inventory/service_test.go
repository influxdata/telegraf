package kube_inventory

import (
	"testing"
	"time"

	"github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/influxdata/telegraf/testutil"
)

func TestService(t *testing.T) {
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
			name: "no service",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/service/": &v1.ServiceList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect service",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/service/": &v1.ServiceList{
						Items: []*v1.Service{
							{
								Spec: &v1.ServiceSpec{
									Ports: []*v1.ServicePort{
										{
											Port:       toInt32Ptr(8080),
											TargetPort: toIntStrPtrI(1234),
											Name:       toStrPtr("diagnostic"),
											Protocol:   toStrPtr("TCP"),
										},
									},
									ExternalIPs: []string{"1.0.0.127"},
									ClusterIP:   toStrPtr("127.0.0.1"),
								},
								Metadata: &metav1.ObjectMeta{
									Generation:        toInt64Ptr(12),
									Namespace:         toStrPtr("ns1"),
									Name:              toStrPtr("checker"),
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
							"port":        int32(8080),
							"target_port": int32(1234),
							"generation":  int64(12),
							"created":     now.UnixNano(),
						},
						Tags: map[string]string{
							"service_name":  "checker",
							"namespace":     "ns1",
							"port_name":     "diagnostic",
							"port_protocol": "TCP",
							"cluster_ip":    "127.0.0.1",
							"ip":            "1.0.0.127",
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
		for _, service := range ((v.handler.responseMap["/service/"]).(*v1.ServiceList)).Items {
			err := ks.gatherService(*service, acc)
			if err != nil {
				t.Errorf("Failed to gather service - %s", err.Error())
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
