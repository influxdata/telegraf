package kube_inventory

import (
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestDaemonSet(t *testing.T) {
	cli := &client{}
	selectInclude := []string{}
	selectExclude := []string{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())
	tests := []struct {
		name     string
		handler  *mockHandler
		output   []telegraf.Metric
		hasError bool
	}{
		{
			name: "no daemon set",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/daemonsets/": &v1.DaemonSetList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect daemonsets",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/daemonsets/": &v1.DaemonSetList{
						Items: []v1.DaemonSet{
							{
								Status: v1.DaemonSetStatus{
									CurrentNumberScheduled: 3,
									DesiredNumberScheduled: 5,
									NumberAvailable:        2,
									NumberMisscheduled:     2,
									NumberReady:            1,
									NumberUnavailable:      1,
									UpdatedNumberScheduled: 2,
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation: 11221,
									Namespace:  "ns1",
									Name:       "daemon1",
									Labels: map[string]string{
										"lab1": "v1",
										"lab2": "v2",
									},
									CreationTimestamp: metav1.Time{Time: now},
								},
								Spec: v1.DaemonSetSpec{
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"select1": "s1",
											"select2": "s2",
										},
									},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_daemonset",
					map[string]string{
						"daemonset_name":   "daemon1",
						"namespace":        "ns1",
						"selector_select1": "s1",
						"selector_select2": "s2",
					},
					map[string]interface{}{
						"generation":               int64(11221),
						"current_number_scheduled": int32(3),
						"desired_number_scheduled": int32(5),
						"number_available":         int32(2),
						"number_misscheduled":      int32(2),
						"number_ready":             int32(1),
						"number_unavailable":       int32(1),
						"updated_number_scheduled": int32(2),
						"created":                  now.UnixNano(),
					},
					time.Unix(0, 0),
				),
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
		require.NoError(t, ks.createSelectorFilters())
		acc := new(testutil.Accumulator)
		for _, dset := range ((v.handler.responseMap["/daemonsets/"]).(*v1.DaemonSetList)).Items {
			ks.gatherDaemonSet(dset, acc)
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

func TestDaemonSetSelectorFilter(t *testing.T) {
	cli := &client{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	responseMap := map[string]interface{}{
		"/daemonsets/": &v1.DaemonSetList{
			Items: []v1.DaemonSet{
				{
					Status: v1.DaemonSetStatus{
						CurrentNumberScheduled: 3,
						DesiredNumberScheduled: 5,
						NumberAvailable:        2,
						NumberMisscheduled:     2,
						NumberReady:            1,
						NumberUnavailable:      1,
						UpdatedNumberScheduled: 2,
					},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 11221,
						Namespace:  "ns1",
						Name:       "daemon1",
						Labels: map[string]string{
							"lab1": "v1",
							"lab2": "v2",
						},
						CreationTimestamp: metav1.Time{Time: time.Now()},
					},
					Spec: v1.DaemonSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"select1": "s1",
								"select2": "s2",
							},
						},
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
		require.NoError(t, ks.createSelectorFilters())
		acc := new(testutil.Accumulator)
		for _, dset := range ((v.handler.responseMap["/daemonsets/"]).(*v1.DaemonSetList)).Items {
			ks.gatherDaemonSet(dset, acc)
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

		require.Equalf(t, v.expected, actual,
			"actual selector tags (%v) do not match expected selector tags (%v)", actual, v.expected)
	}
}
