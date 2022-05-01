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

func TestStatefulSet(t *testing.T) {
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
			name: "no statefulsets",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/statefulsets/": &v1.StatefulSetList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect statefulsets",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/statefulsets/": &v1.StatefulSetList{
						Items: []v1.StatefulSet{
							{
								Status: v1.StatefulSetStatus{
									Replicas:           2,
									CurrentReplicas:    4,
									ReadyReplicas:      1,
									UpdatedReplicas:    3,
									ObservedGeneration: 119,
								},
								Spec: v1.StatefulSetSpec{
									Replicas: toInt32Ptr(3),
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"select1": "s1",
											"select2": "s2",
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        332,
									Namespace:         "ns1",
									Name:              "sts1",
									CreationTimestamp: metav1.Time{Time: now},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_statefulset",
					map[string]string{
						"namespace":        "ns1",
						"statefulset_name": "sts1",
						"selector_select1": "s1",
						"selector_select2": "s2",
					},
					map[string]interface{}{
						"generation":          int64(332),
						"observed_generation": int64(119),
						"created":             now.UnixNano(),
						"spec_replicas":       int32(3),
						"replicas":            int32(2),
						"replicas_current":    int32(4),
						"replicas_ready":      int32(1),
						"replicas_updated":    int32(3),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
		{
			name: "no label selector",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/statefulsets/": &v1.StatefulSetList{
						Items: []v1.StatefulSet{
							{
								Status: v1.StatefulSetStatus{
									Replicas:           2,
									CurrentReplicas:    4,
									ReadyReplicas:      1,
									UpdatedReplicas:    3,
									ObservedGeneration: 119,
								},
								Spec: v1.StatefulSetSpec{
									Replicas: toInt32Ptr(3),
									Selector: nil,
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        332,
									Namespace:         "ns1",
									Name:              "sts1",
									CreationTimestamp: metav1.Time{Time: now},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_statefulset",
					map[string]string{
						"namespace":        "ns1",
						"statefulset_name": "sts1",
					},
					map[string]interface{}{
						"generation":          int64(332),
						"observed_generation": int64(119),
						"created":             now.UnixNano(),
						"spec_replicas":       int32(3),
						"replicas":            int32(2),
						"replicas_current":    int32(4),
						"replicas_ready":      int32(1),
						"replicas_updated":    int32(3),
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
		{
			name: "no desired number of replicas",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/statefulsets/": &v1.StatefulSetList{
						Items: []v1.StatefulSet{
							{
								Status: v1.StatefulSetStatus{
									Replicas:           2,
									CurrentReplicas:    4,
									ReadyReplicas:      1,
									UpdatedReplicas:    3,
									ObservedGeneration: 119,
								},
								Spec: v1.StatefulSetSpec{
									Replicas: nil,
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"select1": "s1",
											"select2": "s2",
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Generation:        332,
									Namespace:         "ns1",
									Name:              "sts1",
									CreationTimestamp: metav1.Time{Time: now},
								},
							},
						},
					},
				},
			},
			output: []telegraf.Metric{
				testutil.MustMetric(
					"kubernetes_statefulset",
					map[string]string{
						"namespace":        "ns1",
						"statefulset_name": "sts1",
						"selector_select1": "s1",
						"selector_select2": "s2",
					},
					map[string]interface{}{
						"generation":          int64(332),
						"observed_generation": int64(119),
						"created":             now.UnixNano(),
						"replicas":            int32(2),
						"replicas_current":    int32(4),
						"replicas_ready":      int32(1),
						"replicas_updated":    int32(3),
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
		acc := &testutil.Accumulator{}
		for _, ss := range ((v.handler.responseMap["/statefulsets/"]).(*v1.StatefulSetList)).Items {
			ks.gatherStatefulSet(ss, acc)
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

func TestStatefulSetSelectorFilter(t *testing.T) {
	cli := &client{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	responseMap := map[string]interface{}{
		"/statefulsets/": &v1.StatefulSetList{
			Items: []v1.StatefulSet{
				{
					Status: v1.StatefulSetStatus{
						Replicas:           2,
						CurrentReplicas:    4,
						ReadyReplicas:      1,
						UpdatedReplicas:    3,
						ObservedGeneration: 119,
					},
					Spec: v1.StatefulSetSpec{
						Replicas: toInt32Ptr(3),
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"select1": "s1",
								"select2": "s2",
							},
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						Generation:        332,
						Namespace:         "ns1",
						Name:              "sts1",
						CreationTimestamp: metav1.Time{Time: now},
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
		for _, ss := range ((v.handler.responseMap["/statefulsets/"]).(*v1.StatefulSetList)).Items {
			ks.gatherStatefulSet(ss, acc)
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
