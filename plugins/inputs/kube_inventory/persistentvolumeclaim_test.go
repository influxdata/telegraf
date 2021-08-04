package kube_inventory

import (
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestPersistentVolumeClaim(t *testing.T) {
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
			name: "no pv claims",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/persistentvolumeclaims/": &corev1.PersistentVolumeClaimList{},
				},
			},
			hasError: false,
		},
		{
			name: "collect pv claims",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/persistentvolumeclaims/": &corev1.PersistentVolumeClaimList{
						Items: []corev1.PersistentVolumeClaim{
							{
								Status: corev1.PersistentVolumeClaimStatus{
									Phase: "bound",
								},
								Spec: corev1.PersistentVolumeClaimSpec{
									VolumeName:       "pvc-dc870fd6-1e08-11e8-b226-02aa4bc06eb8",
									StorageClassName: toStrPtr("ebs-1"),
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"select1": "s1",
											"select2": "s2",
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Namespace: "ns1",
									Name:      "pc1",
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
					"kubernetes_persistentvolumeclaim",
					map[string]string{
						"pvc_name":         "pc1",
						"namespace":        "ns1",
						"storageclass":     "ebs-1",
						"phase":            "bound",
						"selector_select1": "s1",
						"selector_select2": "s2",
					},
					map[string]interface{}{
						"phase_type": 0,
					},
					time.Unix(0, 0),
				),
			},
			hasError: false,
		},
		{
			name:     "no label selectors",
			hasError: false,
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/persistentvolumeclaims/": &corev1.PersistentVolumeClaimList{
						Items: []corev1.PersistentVolumeClaim{
							{
								Status: corev1.PersistentVolumeClaimStatus{
									Phase: "bound",
								},
								Spec: corev1.PersistentVolumeClaimSpec{
									VolumeName:       "pvc-dc870fd6-1e08-11e8-b226-02aa4bc06eb8",
									StorageClassName: toStrPtr("ebs-1"),
									Selector:         nil,
								},
								ObjectMeta: metav1.ObjectMeta{
									Namespace: "ns1",
									Name:      "pc1",
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
					"kubernetes_persistentvolumeclaim",
					map[string]string{
						"pvc_name":     "pc1",
						"namespace":    "ns1",
						"storageclass": "ebs-1",
						"phase":        "bound",
					},
					map[string]interface{}{
						"phase_type": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "no storage class name",
			handler: &mockHandler{
				responseMap: map[string]interface{}{
					"/persistentvolumeclaims/": &corev1.PersistentVolumeClaimList{
						Items: []corev1.PersistentVolumeClaim{
							{
								Status: corev1.PersistentVolumeClaimStatus{
									Phase: "bound",
								},
								Spec: corev1.PersistentVolumeClaimSpec{
									VolumeName:       "pvc-dc870fd6-1e08-11e8-b226-02aa4bc06eb8",
									StorageClassName: nil,
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"select1": "s1",
											"select2": "s2",
										},
									},
								},
								ObjectMeta: metav1.ObjectMeta{
									Namespace: "ns1",
									Name:      "pc1",
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
					"kubernetes_persistentvolumeclaim",
					map[string]string{
						"pvc_name":         "pc1",
						"namespace":        "ns1",
						"phase":            "bound",
						"selector_select1": "s1",
						"selector_select2": "s2",
					},
					map[string]interface{}{
						"phase_type": 0,
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
		for _, pvc := range ((v.handler.responseMap["/persistentvolumeclaims/"]).(*corev1.PersistentVolumeClaimList)).Items {
			ks.gatherPersistentVolumeClaim(pvc, acc)
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

func TestPersistentVolumeClaimSelectorFilter(t *testing.T) {
	cli := &client{}
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 1, 36, 0, now.Location())

	responseMap := map[string]interface{}{
		"/persistentvolumeclaims/": &corev1.PersistentVolumeClaimList{
			Items: []corev1.PersistentVolumeClaim{
				{
					Status: corev1.PersistentVolumeClaimStatus{
						Phase: "bound",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName:       "pvc-dc870fd6-1e08-11e8-b226-02aa4bc06eb8",
						StorageClassName: toStrPtr("ebs-1"),
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"select1": "s1",
								"select2": "s2",
							},
						},
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "ns1",
						Name:      "pc1",
						Labels: map[string]string{
							"lab1": "v1",
							"lab2": "v2",
						},
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
		for _, pvc := range ((v.handler.responseMap["/persistentvolumeclaims/"]).(*corev1.PersistentVolumeClaimList)).Items {
			ks.gatherPersistentVolumeClaim(pvc, acc)
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
