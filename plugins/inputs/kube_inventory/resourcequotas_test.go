package kube_inventory

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestResourceQuota(t *testing.T) {
	cli := &client{}
	now := time.Now()

	tests := []struct {
		name     string
		handler  *mockHandler
		output   []telegraf.Metric
		hasError bool
	}{
		{
			name: "no ressourcequota",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/resourcequotas/": corev1.ResourceQuotaList{},
				},
			},
			output:   []telegraf.Metric{},
			hasError: false,
		},
		{
			name: "collect resourceqota",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/resourcequotas/": corev1.ResourceQuotaList{
						Items: []corev1.ResourceQuota{
							{
								Status: corev1.ResourceQuotaStatus{
									Hard: corev1.ResourceList{
										"cpu":    resource.MustParse("16"),
										"memory": resource.MustParse("125817904Ki"),
										"pods":   resource.MustParse("110"),
									},
									Used: corev1.ResourceList{
										"cpu":    resource.MustParse("10"),
										"memory": resource.MustParse("125715504Ki"),
										"pods":   resource.MustParse("0"),
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation: 11232,
									Namespace:  "ns1",
									Name:       "rq1",
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: metav1.Time{Time: now},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					resourcequotaMeasurement,
					map[string]string{
						"resource":  "rq1",
						"namespace": "ns1",
					},
					map[string]interface{}{
						"hard_cpu":    int64(16),
						"hard_memory": int64(1.28837533696e+11),
						"hard_pods":   int64(110),
						"used_cpu":    int64(10),
						"used_memory": int64(1.28732676096e+11),
						"used_pods":   int64(0),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
	}

	for _, v := range tests {
		ks := &KubernetesInventory{
			client: cli,
		}
		acc := new(testutil.Accumulator)
		for _, quota := range ((v.handler.responseMap["/resourcequotas/"]).(corev1.ResourceQuotaList)).Items {
			ks.gatherResourceQuota(quota, acc)
		}

		err := acc.FirstError()
		if v.hasError {
			require.Errorf(t, err, "%s failed, should have error", v.name)
			continue
		}

		// No error case
		require.NoErrorf(t, err, "%s failed, err: %v", v.name, err)

		require.Len(t, acc.Metrics, len(v.output))
		testutil.RequireMetricsEqual(t, acc.GetTelegrafMetrics(), v.output, testutil.IgnoreTime())
	}
}
