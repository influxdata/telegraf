package t128_transform

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"

	"github.com/stretchr/testify/assert"
)

func TestStateChangeSendFirstSample(t *testing.T) {
	r := newTransformType("state-change")
	r.Fields = map[string]string{"/my/state": "/my/state"}
	assert.Nil(t, r.Init())

	m := newMetric("foo", nil, map[string]interface{}{"/my/state": "state1"}, time.Now())

	rate := r.Apply(m)
	assert.Len(t, rate, 1)
	assert.Equal(t, rate[0].Fields()["/my/state"], "state1")
}

func TestStateChangeSendsOnExpired(t *testing.T) {
	r := newTransformType("state-change")
	r.Fields = map[string]string{"/my/state": "/my/state"}
	r.Expiration = config.Duration(10 * time.Second)
	assert.Nil(t, r.Init())

	t1 := time.Now()
	t2 := t1.Add(10 * time.Second)
	t3 := t2.Add(10*time.Second + 1*time.Nanosecond)

	m1 := newMetric("foo", nil, map[string]interface{}{"/my/state": "state1"}, t1)
	m2 := newMetric("foo", nil, map[string]interface{}{"/my/state": "state1"}, t2)
	m3 := newMetric("foo", nil, map[string]interface{}{"/my/state": "state1"}, t3)

	r.Apply(m1)
	assert.Len(t, r.Apply(m2)[0].FieldList(), 0)
	assert.Len(t, r.Apply(m3)[0].FieldList(), 1)
}

func TestStateChangeLeavesUnmarkedFieldsInTact(t *testing.T) {
	r := newTransformType("state-change")
	r.Fields = map[string]string{"/my/state": "/my/state"}
	assert.Nil(t, r.Init())

	t1 := time.Now()
	t2 := t1.Add(time.Second * 1)

	m1 := newMetric("foo", nil, map[string]interface{}{"/my/state": 50, "/unmarked": 50}, t1)
	m2 := newMetric("foo", nil, map[string]interface{}{"/my/state": 60, "/unmarked": 60}, t2)

	outputs := r.Apply(m1)
	assert.Len(t, outputs, 1)
	assert.Len(t, outputs[0].FieldList(), 2)

	v, ok := outputs[0].GetField("/my/state")
	assert.True(t, ok)
	assert.Equal(t, int64(50), v)

	v, ok = outputs[0].GetField("/unmarked")
	assert.True(t, ok)
	assert.Equal(t, int64(50), v)

	outputs = r.Apply(m2)
	assert.Len(t, outputs, 1)

	v, ok = outputs[0].GetField("/my/state")
	assert.True(t, ok)
	assert.Equal(t, int64(60), v)

	v, ok = outputs[0].GetField("/unmarked")
	assert.True(t, ok)
	assert.Equal(t, int64(60), v)
}

type sample = map[string]interface{}

func TestStateChangeRemoveOriginalAndRename(t *testing.T) {
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
			Fields: map[string]string{"/state": "/state"},
			Samples: []sample{
				{},
				{"/state": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{"/state": int64(50)},
		},
		{
			Name:   "remove-matching-existing-same",
			Fields: map[string]string{"/state": "/state"},
			Samples: []sample{
				{"/state": 45},
				{"/state": 45},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{},
		},
		{
			Name:   "remove-matching-existing-changed",
			Fields: map[string]string{"/state": "/state"},
			Samples: []sample{
				{"/state": 45},
				{"/state": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{"/state": int64(50)},
		},
		{
			Name:   "remove-matching-expired",
			Fields: map[string]string{"/state": "/state"},
			Samples: []sample{
				{"/state": 45},
				{"/state": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{"/state": int64(50)},
		},
		{
			Name:   "remove-mismatching-new",
			Fields: map[string]string{"/state": "/non-state"},
			Samples: []sample{
				{},
				{"/non-state": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{"/state": int64(50)},
		},
		{
			Name:   "remove-mismatching-existing-same",
			Fields: map[string]string{"/state": "/non-state"},
			Samples: []sample{
				{"/non-state": 45},
				{"/non-state": 45},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{},
		},
		{
			Name:   "remove-mismatching-existing-changed",
			Fields: map[string]string{"/state": "/non-state"},
			Samples: []sample{
				{"/non-state": 45},
				{"/non-state": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{"/state": int64(50)},
		},
		{
			Name:   "remove-mismatching-expired",
			Fields: map[string]string{"/state": "/non-state"},
			Samples: []sample{
				{"/non-state": 45},
				{"/non-state": 50},
			},
			Timestamps:     []int{0, 6},
			RemoveOriginal: true,
			Result:         sample{"/state": int64(50)},
		},
		{
			Name:   "leave-matching-new",
			Fields: map[string]string{"/state": "/state"},
			Samples: []sample{
				{},
				{"/state": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{"/state": int64(50)},
		},
		{
			Name:   "leave-matching-existing-same",
			Fields: map[string]string{"/state": "/state"},
			Samples: []sample{
				{"/state": 45},
				{"/state": 45},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{},
		},
		{
			Name:   "leave-matching-existing-changed",
			Fields: map[string]string{"/state": "/state"},
			Samples: []sample{
				{"/state": 45},
				{"/state": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{"/state": int64(50)},
		},
		{
			Name:   "leave-matching-expired",
			Fields: map[string]string{"/state": "/state"},
			Samples: []sample{
				{"/state": 45},
				{"/state": 50},
			},
			Timestamps:     []int{0, 6},
			RemoveOriginal: false,
			Result:         sample{"/state": int64(50)},
		},
		{
			Name:   "leave-mismatching-new",
			Fields: map[string]string{"/state": "/non-state"},
			Samples: []sample{
				{},
				{"/non-state": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{"/non-state": int64(50), "/state": int64(50)},
		},
		{
			Name:   "leave-mismatching-existing-same",
			Fields: map[string]string{"/state": "/non-state"},
			Samples: []sample{
				{"/non-state": 45},
				{"/non-state": 45},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{"/non-state": int64(45)},
		},
		{
			Name:   "leave-mismatching-existing-changed",
			Fields: map[string]string{"/state": "/non-state"},
			Samples: []sample{
				{"/non-state": 45},
				{"/non-state": 50},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: false,
			Result:         sample{"/non-state": int64(50), "/state": int64(50)},
		},
		{
			Name:   "leave-mismatching-expired",
			Fields: map[string]string{"/state": "/non-state"},
			Samples: []sample{
				{"/non-state": 45},
				{"/non-state": 50},
			},
			Timestamps:     []int{0, 6},
			RemoveOriginal: false,
			Result:         sample{"/non-state": int64(50), "/state": int64(50)},
		},

		{
			Name:           "previous-new",
			Fields:         map[string]string{"/state": "/state"},
			PreviousFields: map[string]string{"/previous": "/state"},
			Samples: []sample{
				{"/state": "s1"},
			},
			Timestamps:     []int{0},
			RemoveOriginal: true,
			Result:         sample{"/state": "s1"},
		},
		{
			Name:           "previous-with-previous-value",
			Fields:         map[string]string{"/state": "/state"},
			PreviousFields: map[string]string{"/previous": "/state"},
			Samples: []sample{
				{"/state": "s1"},
				{"/state": "s2"},
			},
			Timestamps:     []int{0, 5},
			RemoveOriginal: true,
			Result:         sample{"/state": "s2", "/previous": "s1"},
		},
		{
			Name:           "previous-after-multiple-matching",
			Fields:         map[string]string{"/state": "/state"},
			PreviousFields: map[string]string{"/previous": "/state"},
			Samples: []sample{
				{"/state": "s1"},
				{"/state": "s1"},
				{"/state": "s2"},
			},
			Timestamps:     []int{0, 5, 10},
			RemoveOriginal: true,
			Result:         sample{"/state": "s2", "/previous": "s1"},
		},
		{
			Name:           "previous-matching-expired",
			Fields:         map[string]string{"/state": "/state"},
			PreviousFields: map[string]string{"/previous": "/state"},
			Samples: []sample{
				{"/state": "s1"},
				{"/state": "s2"},
			},
			Timestamps:     []int{0, 6},
			RemoveOriginal: true,
			Result:         sample{"/state": "s2", "/previous": "s1"},
		},
		{
			Name:           "previous-after-expired",
			Fields:         map[string]string{"/state": "/state"},
			PreviousFields: map[string]string{"/previous": "/state"},
			Samples: []sample{
				{"/state": "s1"},
				{"/state": "s2"},
				{"/state": "s3"},
			},
			Timestamps:     []int{0, 6, 11},
			RemoveOriginal: true,
			Result:         sample{"/state": "s3", "/previous": "s2"},
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
			r := newTransformType("state-change")
			r.Fields = testCase.Fields
			r.PreviousFields = testCase.PreviousFields
			r.RemoveOriginal = testCase.RemoveOriginal
			r.Expiration = config.Duration(5 * time.Second)
			assert.Nil(t, r.Init())

			t1 := time.Now()
			sampleIndex := 0
			for sampleIndex = 0; sampleIndex < len(testCase.Samples)-1; sampleIndex++ {
				r.Apply(newMetric("foo", nil, testCase.Samples[sampleIndex], t1.Add(time.Duration(testCase.Timestamps[sampleIndex])*time.Second)))
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

func TestStateChangeWithPersistence(t *testing.T) {
	testDir, err := ioutil.TempDir("/tmp/", "t128-transform-test")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", testDir)
	}
	defer os.RemoveAll(testDir)

	persistToPath := filepath.Join(testDir, "last-state")

	persistenceFile, err := os.Create(persistToPath)
	if err != nil {
		t.Fatal(err)
	}
	persistenceFile.Close()

	t1 := time.Now()
	r1 := newTranformWithPersistence(t, persistToPath)
	m1 := newMetric("foo", nil, map[string]interface{}{"/state": "state1"}, t1)

	rate1 := r1.Apply(m1)
	assert.Len(t, rate1, 1)
	assert.Equal(t, rate1[0].Fields(), sample{"/state": "state1"})

	r2 := newTranformWithPersistence(t, persistToPath)
	m2 := newMetric("foo", nil, map[string]interface{}{"/state": "state1"}, t1.Add(time.Duration(3)*time.Second))

	rate2 := r2.Apply(m2)
	assert.Len(t, rate2, 1)
	assert.Equal(t, rate2[0].Fields(), sample{})
}

func newTranformWithPersistence(t *testing.T, path string) *T128Transform {
	r := newTransformType("state-change")
	r.Fields = map[string]string{"/state": "/state"}
	r.RemoveOriginal = true
	r.Expiration = config.Duration(5 * time.Second)
	r.PersistTo = path
	r.Log = models.NewLogger("test", "test", "")
	assert.Nil(t, r.Init())

	return r
}
