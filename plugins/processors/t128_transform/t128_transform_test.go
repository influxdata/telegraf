package t128_transform

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/stretchr/testify/assert"
)

var transformTypes []string = []string{
	"rate",
	"diff",
	"state-change",
}

func newMetric(name string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}

	return metric.New(name, tags, fields, timestamp)
}

func TestRemovesFirstSample(t *testing.T) {
	r := newTransform()
	r.Fields = map[string]string{"/my/rate": "/my/rate"}
	assert.Nil(t, r.Init())

	m := newMetric("foo", nil, map[string]interface{}{"/my/rate": 50}, time.Now())

	rate := r.Apply(m)
	assert.Len(t, rate, 1)
	assert.Empty(t, rate[0].FieldList())
}

func TestRemovesExpiredSample(t *testing.T) {
	r := newTransform()
	r.Fields = map[string]string{"/my/rate": "/my/rate"}
	r.Expiration = config.Duration(10 * time.Second)
	assert.Nil(t, r.Init())

	t1 := time.Now()
	t2 := t1.Add(10 * time.Second)
	t3 := t2.Add(10*time.Second + 1*time.Nanosecond)

	m1 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 50}, t1)
	m2 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 60}, t2)
	m3 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 70}, t3)

	r.Apply(m1)
	assert.Len(t, r.Apply(m2)[0].FieldList(), 1)
	assert.Len(t, r.Apply(m3)[0].FieldList(), 0)
}

func TestCalculatesDiffs(t *testing.T) {
	cases := []struct {
		transform string
		value     float64
	}{
		{
			transform: "diff",
			value:     10,
		},
		{
			transform: "rate",
			value:     5,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.transform, func(t *testing.T) {
			r := newTransform()
			r.Fields = map[string]string{
				"/my/rate": "/my/rate",
			}
			r.Transform = testCase.transform
			r.Expiration = config.Duration(5 * time.Second)

			assert.Nil(t, r.Init())

			t1 := time.Now()
			t2 := t1.Add(time.Second * 2)
			t3 := t2.Add(time.Second * 2)

			// expire at t4
			t4 := t3.Add(time.Second * 6)
			t5 := t4.Add(time.Second * 2)

			m1 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 50}, t1)
			m2 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 60}, t2)
			m3 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 70}, t3)
			m4 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 80}, t4)
			m5 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 90}, t5)

			// nothing on first item
			results := r.Apply(m1)
			assert.Len(t, results, 1)
			assert.Len(t, results[0].FieldList(), 0)

			// starts reporting on the second observation
			results = r.Apply(m2)
			expected := newMetric("foo", nil, map[string]interface{}{"/my/rate": testCase.value}, t2)

			assert.Len(t, results, 1)
			assert.Equal(t, expected, results[0])

			// continues reporting rates based on prior value
			results = r.Apply(m3)
			expected = newMetric("foo", nil, map[string]interface{}{"/my/rate": testCase.value}, t3)

			assert.Len(t, results, 1)
			assert.Equal(t, expected, results[0])

			// doesn't report expired value
			results = r.Apply(m4)
			assert.Len(t, results, 1)
			assert.Len(t, results[0].FieldList(), 0)

			// resumes reporting after an expiration
			results = r.Apply(m5)
			expected = newMetric("foo", nil, map[string]interface{}{"/my/rate": testCase.value}, t5)

			assert.Len(t, results, 1)
			assert.Equal(t, expected, results[0])
		})
	}
}

func TestLeavesUnmarkedFieldsInTact(t *testing.T) {
	r := newTransform()
	r.Fields = map[string]string{"/my/rate": "/my/rate"}
	assert.Nil(t, r.Init())

	t1 := time.Now()
	t2 := t1.Add(time.Second * 1)

	m1 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 50, "/unmarked": float64(50)}, t1)
	m2 := newMetric("foo", nil, map[string]interface{}{"/my/rate": 60, "/unmarked": float64(60)}, t2)

	outputs := r.Apply(m1)
	assert.Len(t, outputs, 1)
	assert.Len(t, outputs[0].FieldList(), 1)

	v, ok := outputs[0].GetField("/unmarked")
	assert.True(t, ok)
	assert.Equal(t, float64(50), v)

	outputs = r.Apply(m2)
	assert.Len(t, outputs, 1)

	v, ok = outputs[0].GetField("/my/rate")
	assert.True(t, ok)
	assert.Equal(t, float64(10), v)

	v, ok = outputs[0].GetField("/unmarked")
	assert.True(t, ok)
	assert.Equal(t, float64(60), v)
}

func TestRemoveOriginalAndRename(t *testing.T) {
	type sample = map[string]interface{}

	testCases := []struct {
		Name           string
		Fields         map[string]string
		Samples        []sample
		PreviousFields map[string]string
		Timestamps     []int
		RemoveOriginal bool
		Result         sample
	}{
		{
			Name:   "remove-matching-new",
			Fields: map[string]string{"/rate": "/rate"},
			Samples: []sample{
				{},
				{"/rate": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{},
		},
		{
			Name:   "remove-matching-existing",
			Fields: map[string]string{"/rate": "/rate"},
			Samples: []sample{
				{"/rate": 45},
				{"/rate": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{"/rate": float64(1)},
		},
		{
			Name:   "remove-matching-expired",
			Fields: map[string]string{"/rate": "/rate"},
			Samples: []sample{
				{"/rate": 45},
				{"/rate": 50},
			},
			Timestamps:     []int{0, 6},
			RemoveOriginal: true,
			Result:         sample{},
		},
		{
			Name:   "remove-mismatching-new",
			Fields: map[string]string{"/rate": "/non-rate"},
			Samples: []sample{
				{},
				{"/non-rate": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{},
		},
		{
			Name:   "remove-mismatching-existing",
			Fields: map[string]string{"/rate": "/non-rate"},
			Samples: []sample{
				{"/non-rate": 45},
				{"/non-rate": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{"/rate": float64(1)},
		},
		{
			Name:   "remove-mismatching-expired",
			Fields: map[string]string{"/rate": "/non-rate"},
			Samples: []sample{
				{"/non-rate": 45},
				{"/non-rate": 50},
			},
			Timestamps:     []int{0, 6},
			RemoveOriginal: true,
			Result:         sample{},
		},

		{
			Name:   "leave-matching-new",
			Fields: map[string]string{"/rate": "/rate"},
			Samples: []sample{
				{},
				{"/rate": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{},
		},
		{
			Name:   "leave-matching-existing",
			Fields: map[string]string{"/rate": "/rate"},
			Samples: []sample{
				{"/rate": 45},
				{"/rate": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{"/rate": float64(1)},
		},
		{
			Name:   "leave-matching-expired",
			Fields: map[string]string{"/rate": "/rate"},
			Samples: []sample{
				{"/rate": 45},
				{"/rate": 50},
			},
			Timestamps:     []int{0, 6},
			RemoveOriginal: false,
			Result:         sample{},
		},
		{
			Name:   "leave-mismatching-new",
			Fields: map[string]string{"/rate": "/non-rate"},
			Samples: []sample{
				{},
				{"/non-rate": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{"/non-rate": int64(50)},
		},
		{
			Name:   "leave-mismatching-existing",
			Fields: map[string]string{"/rate": "/non-rate"},
			Samples: []sample{
				{"/non-rate": 45},
				{"/non-rate": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{"/non-rate": int64(50), "/rate": float64(1)},
		},
		{
			Name:   "leave-mismatching-expired",
			Fields: map[string]string{"/rate": "/non-rate"},
			Samples: []sample{
				{"/non-rate": 45},
				{"/non-rate": 50},
			},
			Timestamps:     []int{0, 6},
			RemoveOriginal: false,
			Result:         sample{"/non-rate": int64(50)},
		},

		{
			Name:           "previous-new",
			Fields:         map[string]string{"/rate": "/total"},
			PreviousFields: map[string]string{"/previous": "/rate"},
			Samples: []sample{
				{"/total": 50},
			},
			Timestamps:     []int{0},
			RemoveOriginal: true,
			Result:         sample{},
		},
		{
			Name:           "previous-new-no-previous-value",
			Fields:         map[string]string{"/rate": "/total"},
			PreviousFields: map[string]string{"/previous": "/rate"},
			Samples: []sample{
				{"/total": 50},
				{"/total": 55},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{"/rate": float64(1)},
		},
		{
			Name:           "previous-new-previous-value",
			Fields:         map[string]string{"/rate": "/total"},
			PreviousFields: map[string]string{"/previous": "/rate"},
			Samples: []sample{
				{"/total": 45},
				{"/total": 55},
				{"/total": 60},
			},
			Timestamps:     []int{0, 5, 10},
			RemoveOriginal: true,
			Result:         sample{"/rate": float64(1), "/previous": float64(2)},
		},
		{
			Name:           "previous-expired",
			Fields:         map[string]string{"/rate": "/total"},
			PreviousFields: map[string]string{"/previous": "/rate"},
			Samples: []sample{
				{"/total": 45},
				{"/total": 50},
				{"/total": 55}, // previous assigned here
				{"/total": 60},
			},
			Timestamps:     []int{0, 5, 10, 16},
			RemoveOriginal: true,
			Result:         sample{},
		},
		{
			Name:           "previous-null-after-expired",
			Fields:         map[string]string{"/rate": "/total"},
			PreviousFields: map[string]string{"/previous": "/rate"},
			Samples: []sample{
				{"/total": 45},
				{"/total": 50},
				{"/total": 55}, // previous assigned here
				{"/total": 60}, // expired here
				{"/total": 65}, // first point - shouldn't have previous
			},
			Timestamps:     []int{0, 5, 10, 16, 21},
			RemoveOriginal: true,
			Result:         sample{"/rate": float64(1)},
		},
		{
			Name:           "previous-valid-multiple-after-expired",
			Fields:         map[string]string{"/rate": "/total"},
			PreviousFields: map[string]string{"/previous": "/rate"},
			Samples: []sample{
				{"/total": 45},
				{"/total": 50},
				{"/total": 55}, // previous assigned
				{"/total": 65},
				{"/total": 70}, // first point
				{"/total": 80}, // first with previous
			},
			Timestamps:     []int{0, 5, 10, 16, 21, 26},
			RemoveOriginal: true,
			Result:         sample{"/rate": float64(2), "/previous": float64(1)},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			assert.True(t, len(testCase.Samples) > 0)

			assert.Truef(t,
				len(testCase.Samples) == len(testCase.Timestamps),
				"The number of timestamps (%v) needs to equal the number of samples (%v)",
				len(testCase.Timestamps),
				len(testCase.Samples),
			)

			r := newTransform()
			r.Fields = testCase.Fields
			r.PreviousFields = testCase.PreviousFields
			r.RemoveOriginal = testCase.RemoveOriginal
			r.Expiration = config.Duration(5 * time.Second)
			r.Log = testutil.Logger{}
			assert.Nil(t, r.Init())

			t1 := time.Now()
			sampleIndex := 0
			for sampleIndex = 0; sampleIndex < len(testCase.Samples)-1; sampleIndex++ {
				r.Apply(newMetric(
					"foo",
					nil,
					testCase.Samples[sampleIndex],
					t1.Add(time.Duration(testCase.Timestamps[sampleIndex])*time.Second)),
				)
			}

			m := newMetric("foo", nil,
				testCase.Samples[sampleIndex],
				t1.Add(time.Duration(testCase.Timestamps[sampleIndex])*time.Second),
			)

			result := r.Apply(m)[0]

			assert.Equal(t, result.Fields(), testCase.Result)
		})
	}
}

func TestFailsRateOnNonIncreasingTimestamp(t *testing.T) {
	r := newTransform()
	r.Fields = map[string]string{
		"/my/rate": "/my/rate",
	}
	r.Log = testutil.Logger{}
	assert.Nil(t, r.Init())

	t1 := time.Now()
	t2 := t1
	t3 := t1.Add(-time.Second * 1)

	apply := func(timestamp time.Time) []telegraf.Metric {
		return r.Apply(newMetric("foo", nil, map[string]interface{}{"/my/rate": 50}, timestamp))
	}

	for _, timestamp := range []time.Time{t1, t2, t3} {
		outputs := apply(timestamp)
		assert.Len(t, outputs, 1)
		assert.Len(t, outputs[0].FieldList(), 0)
	}
}

func TestFailsOnConflictingFieldMappings(t *testing.T) {
	for _, transformType := range transformTypes {
		t.Run(transformType, func(t *testing.T) {
			r := newTransformType(transformType)
			r.Fields = map[string]string{
				"/my/rate":       "/my/total",
				"/my/other/rate": "/my/total",
			}

			assert.EqualError(t, r.Init(), "both '/my/other/rate' and '/my/rate' are configured to be calculated from '/my/total'")
		})
	}
}

func TestFailsMissingPreviousSource(t *testing.T) {
	for _, transformType := range transformTypes {
		t.Run(transformType, func(t *testing.T) {
			r := newTransformType(transformType)
			r.Fields = map[string]string{
				"/my/rate": "/my/total",
			}
			r.PreviousFields = map[string]string{
				"/my/previous": "/my/missing",
			}

			assert.EqualError(t, r.Init(), "the previous field '/my/previous' references a transformed field '/my/missing' which does not exist")
		})
	}
}

func TestFailsOnInvalidTransform(t *testing.T) {
	r := newTransform()
	r.Fields = map[string]string{"/my/rate": "/my/rate"}
	r.Transform = "invalid"

	assert.EqualError(t, r.Init(), "'transform' is required and must be 'diff', 'rate', or 'state-change'")
}

func TestLoadsFromToml(t *testing.T) {

	plugin := &T128Transform{}
	exampleConfig := []byte(`
		expiration = "10s"
		transform = "diff"

		[fields]
			"/my/rate" = "/my/total"
			"/other/rate" = "/other/total"
	`)

	assert.NoError(t, toml.Unmarshal(exampleConfig, plugin))
	assert.Equal(t, map[string]string{"/my/rate": "/my/total", "/other/rate": "/other/total"}, plugin.Fields)
	assert.Equal(t, plugin.Expiration, config.Duration(10*time.Second))
	assert.Equal(t, "diff", plugin.Transform)
}
