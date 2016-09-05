package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter_ApplyEmpty(t *testing.T) {
	f := Filter{}
	require.NoError(t, f.Compile())
	assert.False(t, f.IsActive())

	assert.True(t, f.Apply("m", map[string]interface{}{"value": int64(1)}, map[string]string{}))
}

func TestFilter_ApplyTagsDontPass(t *testing.T) {
	filters := []TagFilter{
		TagFilter{
			Name:   "cpu",
			Filter: []string{"cpu-*"},
		},
	}
	f := Filter{
		TagDrop: filters,
	}
	require.NoError(t, f.Compile())
	require.NoError(t, f.Compile())
	assert.True(t, f.IsActive())

	assert.False(t, f.Apply("m",
		map[string]interface{}{"value": int64(1)},
		map[string]string{"cpu": "cpu-total"}))
}

func TestFilter_ApplyDeleteFields(t *testing.T) {
	f := Filter{
		FieldDrop: []string{"value"},
	}
	require.NoError(t, f.Compile())
	require.NoError(t, f.Compile())
	assert.True(t, f.IsActive())

	fields := map[string]interface{}{"value": int64(1), "value2": int64(2)}
	assert.True(t, f.Apply("m", fields, nil))
	assert.Equal(t, map[string]interface{}{"value2": int64(2)}, fields)
}

func TestFilter_ApplyDeleteAllFields(t *testing.T) {
	f := Filter{
		FieldDrop: []string{"value*"},
	}
	require.NoError(t, f.Compile())
	require.NoError(t, f.Compile())
	assert.True(t, f.IsActive())

	fields := map[string]interface{}{"value": int64(1), "value2": int64(2)}
	assert.False(t, f.Apply("m", fields, nil))
}

func TestFilter_Empty(t *testing.T) {
	f := Filter{}

	measurements := []string{
		"foo",
		"bar",
		"barfoo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"supercalifradjulisticexpialidocious",
	}

	for _, measurement := range measurements {
		if !f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}
}

func TestFilter_NamePass(t *testing.T) {
	f := Filter{
		NamePass: []string{"foo*", "cpu_usage_idle"},
	}
	require.NoError(t, f.Compile())

	passes := []string{
		"foo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"cpu_usage_idle",
	}

	drops := []string{
		"bar",
		"barfoo",
		"bar_foo",
		"cpu_usage_busy",
	}

	for _, measurement := range passes {
		if !f.shouldNamePass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.shouldNamePass(measurement) {
			t.Errorf("Expected measurement %s to drop", measurement)
		}
	}
}

func TestFilter_NameDrop(t *testing.T) {
	f := Filter{
		NameDrop: []string{"foo*", "cpu_usage_idle"},
	}
	require.NoError(t, f.Compile())

	drops := []string{
		"foo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"cpu_usage_idle",
	}

	passes := []string{
		"bar",
		"barfoo",
		"bar_foo",
		"cpu_usage_busy",
	}

	for _, measurement := range passes {
		if !f.shouldNamePass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.shouldNamePass(measurement) {
			t.Errorf("Expected measurement %s to drop", measurement)
		}
	}
}

func TestFilter_FieldPass(t *testing.T) {
	f := Filter{
		FieldPass: []string{"foo*", "cpu_usage_idle"},
	}
	require.NoError(t, f.Compile())

	passes := []string{
		"foo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"cpu_usage_idle",
	}

	drops := []string{
		"bar",
		"barfoo",
		"bar_foo",
		"cpu_usage_busy",
	}

	for _, measurement := range passes {
		if !f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to drop", measurement)
		}
	}
}

func TestFilter_FieldDrop(t *testing.T) {
	f := Filter{
		FieldDrop: []string{"foo*", "cpu_usage_idle"},
	}
	require.NoError(t, f.Compile())

	drops := []string{
		"foo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"cpu_usage_idle",
	}

	passes := []string{
		"bar",
		"barfoo",
		"bar_foo",
		"cpu_usage_busy",
	}

	for _, measurement := range passes {
		if !f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to drop", measurement)
		}
	}
}

func TestFilter_TagPass(t *testing.T) {
	filters := []TagFilter{
		TagFilter{
			Name:   "cpu",
			Filter: []string{"cpu-*"},
		},
		TagFilter{
			Name:   "mem",
			Filter: []string{"mem_free"},
		}}
	f := Filter{
		TagPass: filters,
	}
	require.NoError(t, f.Compile())

	passes := []map[string]string{
		{"cpu": "cpu-total"},
		{"cpu": "cpu-0"},
		{"cpu": "cpu-1"},
		{"cpu": "cpu-2"},
		{"mem": "mem_free"},
	}

	drops := []map[string]string{
		{"cpu": "cputotal"},
		{"cpu": "cpu0"},
		{"cpu": "cpu1"},
		{"cpu": "cpu2"},
		{"mem": "mem_used"},
	}

	for _, tags := range passes {
		if !f.shouldTagsPass(tags) {
			t.Errorf("Expected tags %v to pass", tags)
		}
	}

	for _, tags := range drops {
		if f.shouldTagsPass(tags) {
			t.Errorf("Expected tags %v to drop", tags)
		}
	}
}

func TestFilter_TagDrop(t *testing.T) {
	filters := []TagFilter{
		TagFilter{
			Name:   "cpu",
			Filter: []string{"cpu-*"},
		},
		TagFilter{
			Name:   "mem",
			Filter: []string{"mem_free"},
		}}
	f := Filter{
		TagDrop: filters,
	}
	require.NoError(t, f.Compile())

	drops := []map[string]string{
		{"cpu": "cpu-total"},
		{"cpu": "cpu-0"},
		{"cpu": "cpu-1"},
		{"cpu": "cpu-2"},
		{"mem": "mem_free"},
	}

	passes := []map[string]string{
		{"cpu": "cputotal"},
		{"cpu": "cpu0"},
		{"cpu": "cpu1"},
		{"cpu": "cpu2"},
		{"mem": "mem_used"},
	}

	for _, tags := range passes {
		if !f.shouldTagsPass(tags) {
			t.Errorf("Expected tags %v to pass", tags)
		}
	}

	for _, tags := range drops {
		if f.shouldTagsPass(tags) {
			t.Errorf("Expected tags %v to drop", tags)
		}
	}
}

func TestFilter_FilterTagsNoMatches(t *testing.T) {
	pretags := map[string]string{
		"host":  "localhost",
		"mytag": "foobar",
	}
	f := Filter{
		TagExclude: []string{"nomatch"},
	}
	require.NoError(t, f.Compile())

	f.filterTags(pretags)
	assert.Equal(t, map[string]string{
		"host":  "localhost",
		"mytag": "foobar",
	}, pretags)

	f = Filter{
		TagInclude: []string{"nomatch"},
	}
	require.NoError(t, f.Compile())

	f.filterTags(pretags)
	assert.Equal(t, map[string]string{}, pretags)
}

func TestFilter_FilterTagsMatches(t *testing.T) {
	pretags := map[string]string{
		"host":  "localhost",
		"mytag": "foobar",
	}
	f := Filter{
		TagExclude: []string{"ho*"},
	}
	require.NoError(t, f.Compile())

	f.filterTags(pretags)
	assert.Equal(t, map[string]string{
		"mytag": "foobar",
	}, pretags)

	pretags = map[string]string{
		"host":  "localhost",
		"mytag": "foobar",
	}
	f = Filter{
		TagInclude: []string{"my*"},
	}
	require.NoError(t, f.Compile())

	f.filterTags(pretags)
	assert.Equal(t, map[string]string{
		"mytag": "foobar",
	}, pretags)
}
