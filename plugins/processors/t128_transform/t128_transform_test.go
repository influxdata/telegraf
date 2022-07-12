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
	testCases := []struct {
		Name            string
		Fields          map[string]string
		LastSample      map[string]interface{}
		CurrentSample   map[string]interface{}
		TimeDelta       time.Duration
		RemoveOriginal  bool
		RemainingFields []string
	}{
		{
			Name:            "remove-matching-new",
			Fields:          map[string]string{"/rate": "/rate"},
			LastSample:      map[string]interface{}{},
			CurrentSample:   map[string]interface{}{"/rate": 50},
			TimeDelta:       5 * time.Second,
			RemoveOriginal:  true,
			RemainingFields: []string{},
		},
		{
			Name:            "remove-matching-existing",
			Fields:          map[string]string{"/rate": "/rate"},
			LastSample:      map[string]interface{}{"/rate": 45},
			CurrentSample:   map[string]interface{}{"/rate": 50},
			TimeDelta:       5 * time.Second,
			RemoveOriginal:  true,
			RemainingFields: []string{"/rate"},
		},
		{
			Name:            "remove-matching-expired",
			Fields:          map[string]string{"/rate": "/rate"},
			LastSample:      map[string]interface{}{"/rate": 45},
			CurrentSample:   map[string]interface{}{"/rate": 50},
			TimeDelta:       6 * time.Second,
			RemoveOriginal:  true,
			RemainingFields: []string{},
		},
		{
			Name:            "remove-mismatching-new",
			Fields:          map[string]string{"/rate": "/non-rate"},
			LastSample:      map[string]interface{}{},
			CurrentSample:   map[string]interface{}{"/non-rate": 50},
			TimeDelta:       5 * time.Second,
			RemoveOriginal:  true,
			RemainingFields: []string{},
		},
		{
			Name:            "remove-mismatching-existing",
			Fields:          map[string]string{"/rate": "/non-rate"},
			LastSample:      map[string]interface{}{"/non-rate": 45},
			CurrentSample:   map[string]interface{}{"/non-rate": 50},
			TimeDelta:       5 * time.Second,
			RemoveOriginal:  true,
			RemainingFields: []string{"/rate"},
		},
		{
			Name:            "remove-mismatching-expired",
			Fields:          map[string]string{"/rate": "/non-rate"},
			LastSample:      map[string]interface{}{"/non-rate": 45},
			CurrentSample:   map[string]interface{}{"/non-rate": 50},
			TimeDelta:       6 * time.Second,
			RemoveOriginal:  true,
			RemainingFields: []string{},
		},

		{
			Name:            "leave-matching-new",
			Fields:          map[string]string{"/rate": "/rate"},
			LastSample:      map[string]interface{}{},
			CurrentSample:   map[string]interface{}{"/rate": 50},
			TimeDelta:       5 * time.Second,
			RemoveOriginal:  false,
			RemainingFields: []string{},
		},
		{
			Name:            "leave-matching-existing",
			Fields:          map[string]string{"/rate": "/rate"},
			LastSample:      map[string]interface{}{"/rate": 45},
			CurrentSample:   map[string]interface{}{"/rate": 50},
			TimeDelta:       5 * time.Second,
			RemoveOriginal:  false,
			RemainingFields: []string{"/rate"},
		},
		{
			Name:            "leave-matching-expired",
			Fields:          map[string]string{"/rate": "/rate"},
			LastSample:      map[string]interface{}{"/rate": 45},
			CurrentSample:   map[string]interface{}{"/rate": 50},
			TimeDelta:       6 * time.Second,
			RemoveOriginal:  false,
			RemainingFields: []string{},
		},
		{
			Name:            "leave-mismatching-new",
			Fields:          map[string]string{"/rate": "/non-rate"},
			LastSample:      map[string]interface{}{},
			CurrentSample:   map[string]interface{}{"/non-rate": 50},
			TimeDelta:       5 * time.Second,
			RemoveOriginal:  false,
			RemainingFields: []string{"/non-rate"},
		},
		{
			Name:            "leave-mismatching-existing",
			Fields:          map[string]string{"/rate": "/non-rate"},
			LastSample:      map[string]interface{}{"/non-rate": 45},
			CurrentSample:   map[string]interface{}{"/non-rate": 50},
			TimeDelta:       5 * time.Second,
			RemoveOriginal:  false,
			RemainingFields: []string{"/non-rate", "/rate"},
		},
		{
			Name:            "leave-mismatching-expired",
			Fields:          map[string]string{"/rate": "/non-rate"},
			LastSample:      map[string]interface{}{"/non-rate": 45},
			CurrentSample:   map[string]interface{}{"/non-rate": 50},
			TimeDelta:       6 * time.Second,
			RemoveOriginal:  false,
			RemainingFields: []string{"/non-rate"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := newTransform()
			r.Fields = testCase.Fields
			r.RemoveOriginal = testCase.RemoveOriginal
			r.Expiration = config.Duration(5 * time.Second)
			assert.Nil(t, r.Init())

			t1 := time.Now()
			m1 := newMetric("foo", nil, testCase.LastSample, t1)
			m2 := newMetric("foo", nil, testCase.CurrentSample, t1.Add(testCase.TimeDelta))

			r.Apply(m1)
			result := r.Apply(m2)[0]

			assert.Len(t, result.FieldList(), len(testCase.RemainingFields))

			for _, field := range testCase.RemainingFields {
				_, exists := result.GetField(field)
				assert.Truef(t, exists, "the field '%v' doesn't exist", field)
			}
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
	r := newTransform()
	r.Fields = map[string]string{
		"/my/rate":       "/my/total",
		"/my/other/rate": "/my/total",
	}

	assert.EqualError(t, r.Init(), "both '/my/other/rate' and '/my/rate' are configured to be calculated from '/my/total'")
}

func TestFailsOnInvalidTransform(t *testing.T) {
	r := newTransform()
	r.Fields = map[string]string{"/my/rate": "/my/rate"}
	r.Transform = "invalid"

	assert.EqualError(t, r.Init(), "'transform' is required and must be 'diff' or 'rate'")
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
