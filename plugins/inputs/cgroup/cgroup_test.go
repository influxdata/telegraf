// +build linux

package cgroup

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
)

var cg1 = &CGroup{
	Paths: []string{"testdata/memory"},
	Files: []string{
		"memory.empty",
		"memory.max_usage_in_bytes",
		"memory.limit_in_bytes",
		"memory.stat",
		"memory.use_hierarchy",
		"notify_on_release",
	},
}

func assertContainsFields(a *testutil.Accumulator, t *testing.T, measurement string, fieldSet []map[string]interface{}) {
	a.Lock()
	defer a.Unlock()

	numEquals := 0
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for _, fields := range fieldSet {
				if reflect.DeepEqual(fields, p.Fields) {
					numEquals++
				}
			}
		}
	}

	if numEquals != len(fieldSet) {
		assert.Fail(t, fmt.Sprintf("only %d of %d are equal", numEquals, len(fieldSet)))
	}
}

func TestCgroupStatistics_1(t *testing.T) {
	var acc testutil.Accumulator

	err := cg1.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"memory.stat.cache":           1739362304123123123,
		"memory.stat.rss":             1775325184,
		"memory.stat.rss_huge":        778043392,
		"memory.stat.mapped_file":     421036032,
		"memory.stat.dirty":           -307200,
		"memory.max_usage_in_bytes.0": 0,
		"memory.max_usage_in_bytes.1": -1,
		"memory.max_usage_in_bytes.2": 2,
		"memory.limit_in_bytes":       223372036854771712,
		"memory.use_hierarchy":        "12-781",
		"notify_on_release":           0,
		"path":                        "testdata/memory",
	}
	assertContainsFields(&acc, t, "cgroup", []map[string]interface{}{fields})
}

// ======================================================================

var cg2 = &CGroup{
	Paths: []string{"testdata/cpu"},
	Files: []string{"cpuacct.usage_percpu"},
}

func TestCgroupStatistics_2(t *testing.T) {
	var acc testutil.Accumulator

	err := cg2.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"cpuacct.usage_percpu.0": -1452543795404,
		"cpuacct.usage_percpu.1": 1376681271659,
		"cpuacct.usage_percpu.2": 1450950799997,
		"cpuacct.usage_percpu.3": -1473113374257,
		"path": "testdata/cpu",
	}
	assertContainsFields(&acc, t, "cgroup", []map[string]interface{}{fields})
}

// ======================================================================

var cg3 = &CGroup{
	Paths: []string{"testdata/memory/*"},
	Files: []string{"memory.limit_in_bytes"},
}

func TestCgroupStatistics_3(t *testing.T) {
	var acc testutil.Accumulator

	err := cg3.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"memory.limit_in_bytes": 223372036854771712,
		"path":                  "testdata/memory/group_1",
	}

	fieldsTwo := map[string]interface{}{
		"memory.limit_in_bytes": 223372036854771712,
		"path":                  "testdata/memory/group_2",
	}
	assertContainsFields(&acc, t, "cgroup", []map[string]interface{}{fields, fieldsTwo})
}

// ======================================================================

var cg4 = &CGroup{
	Paths: []string{"testdata/memory/*/*", "testdata/memory/group_2"},
	Files: []string{"memory.limit_in_bytes"},
}

func TestCgroupStatistics_4(t *testing.T) {
	var acc testutil.Accumulator

	err := cg4.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"memory.limit_in_bytes": 223372036854771712,
		"path":                  "testdata/memory/group_1/group_1_1",
	}

	fieldsTwo := map[string]interface{}{
		"memory.limit_in_bytes": 223372036854771712,
		"path":                  "testdata/memory/group_1/group_1_2",
	}

	fieldsThree := map[string]interface{}{
		"memory.limit_in_bytes": 223372036854771712,
		"path":                  "testdata/memory/group_2",
	}

	assertContainsFields(&acc, t, "cgroup", []map[string]interface{}{fields, fieldsTwo, fieldsThree})
}

// ======================================================================

var cg5 = &CGroup{
	Paths: []string{"testdata/memory/*/group_1_1"},
	Files: []string{"memory.limit_in_bytes"},
}

func TestCgroupStatistics_5(t *testing.T) {
	var acc testutil.Accumulator

	err := cg5.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"memory.limit_in_bytes": 223372036854771712,
		"path":                  "testdata/memory/group_1/group_1_1",
	}

	fieldsTwo := map[string]interface{}{
		"memory.limit_in_bytes": 223372036854771712,
		"path":                  "testdata/memory/group_2/group_1_1",
	}
	assertContainsFields(&acc, t, "cgroup", []map[string]interface{}{fields, fieldsTwo})
}

// ======================================================================

var cg6 = &CGroup{
	Paths: []string{"testdata/memory"},
	Files: []string{"memory.us*", "*/memory.kmem.*"},
}

func TestCgroupStatistics_6(t *testing.T) {
	var acc testutil.Accumulator

	err := cg6.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"memory.usage_in_bytes":      3513667584,
		"memory.use_hierarchy":       "12-781",
		"memory.kmem.limit_in_bytes": 9223372036854771712,
		"path": "testdata/memory",
	}
	assertContainsFields(&acc, t, "cgroup", []map[string]interface{}{fields})
}
