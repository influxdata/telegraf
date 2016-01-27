package models

import (
	"testing"
)

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
		if !f.ShouldPass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}
}

func TestFilter_Pass(t *testing.T) {
	f := Filter{
		Pass: []string{"foo*", "cpu_usage_idle"},
	}

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
		if !f.ShouldPass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.ShouldPass(measurement) {
			t.Errorf("Expected measurement %s to drop", measurement)
		}
	}
}

func TestFilter_Drop(t *testing.T) {
	f := Filter{
		Drop: []string{"foo*", "cpu_usage_idle"},
	}

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
		if !f.ShouldPass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.ShouldPass(measurement) {
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
		if !f.ShouldTagsPass(tags) {
			t.Errorf("Expected tags %v to pass", tags)
		}
	}

	for _, tags := range drops {
		if f.ShouldTagsPass(tags) {
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
		if !f.ShouldTagsPass(tags) {
			t.Errorf("Expected tags %v to pass", tags)
		}
	}

	for _, tags := range drops {
		if f.ShouldTagsPass(tags) {
			t.Errorf("Expected tags %v to drop", tags)
		}
	}
}
