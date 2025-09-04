package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetPluginLabelSelections(t *testing.T) {
	tests := []struct {
		name           string
		selections     []string
		expectedGroups []int // denotes the length of each group
		wantErr        bool
	}{
		{
			name:           "single selectors",
			selections:     []string{"env=prod", "region=dc-23"},
			expectedGroups: []int{1, 1}, // two groups, each with one selector
		},
		{
			name:           "multiple selectors",
			selections:     []string{"env=prod;region=dc-23", "env=dev;app=backend;policy=web"},
			expectedGroups: []int{2, 3}, // two groups, one with two selectors and one with three
		},
		{
			name:       "invalid selector syntax",
			selections: []string{"env=prod;region=dc-23", "invalid-selector"},
			wantErr:    true,
		},
		{
			name:           "nil selectors",
			selections:     nil,
			expectedGroups: []int{0},
		},
		{
			name:           "empty selector",
			selections:     []string{""},
			expectedGroups: []int{0},
		},
		{
			name:           "multiple empty selectors",
			selections:     []string{"", "app=web;env=prod*", ""},
			expectedGroups: []int{2}, // only one valid group with 2 selectors
		},
		{
			name:       "duplicate within group",
			selections: []string{"env=prod;app=web;env=staging"},
			wantErr:    true,
		},
		{
			name:       "invalid key",
			selections: []string{"invalid$key=prod", "region=dc-23"},
			wantErr:    true,
		},
		{
			name:       "invalid value",
			selections: []string{"env=prod;app=web;invalid=value&()"},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pluginSelector labelSelector
			err := pluginSelector.setSelections(tt.selections)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			for i, group := range pluginSelector.groups {
				require.Equal(t, len(group), tt.expectedGroups[i])
			}
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name      string
		selectors []string
		labels    map[string]string
		expected  bool
	}{
		{
			name:      "[Backward Compatibility] No selectors, should run",
			selectors: nil,
			labels:    map[string]string{"env": "prod"},
			expected:  true,
		},
		{
			name:      "[Backward Compatibility] No labels, should run",
			selectors: []string{"env=prod"},
			labels:    nil,
			expected:  true,
		},
		{
			name:      "Simple exact match",
			selectors: []string{"env=prod"},
			labels:    map[string]string{"env": "prod"},
			expected:  true,
		},
		{
			name:      "Simple mismatch",
			selectors: []string{"env=prod"},
			labels:    map[string]string{"env": "dev"},
			expected:  false,
		},
		{
			name:      "extra labels ignored",
			selectors: []string{"env=prod"},
			labels:    map[string]string{"env": "prod", "region": "us-east"},
			expected:  true,
		},
		{
			name:      "AND inside selector (all match)",
			selectors: []string{"env=prod;region=dc-23"},
			labels:    map[string]string{"env": "prod", "region": "dc-23"},
			expected:  true,
		},
		{
			name:      "AND inside selector (partial match fail)",
			selectors: []string{"env=prod;region=dc-23"},
			labels:    map[string]string{"env": "prod", "region": "dc-24"},
			expected:  false,
		},
		{
			name:      "Simple Wildcard match",
			selectors: []string{"region=dc-*"},
			labels:    map[string]string{"region": "dc-23"},
			expected:  true,
		},
		{
			name:      "Simple Wildcard no match",
			selectors: []string{"region=us-*"},
			labels:    map[string]string{"region": "eu-1"},
			expected:  false,
		},
		{
			name:      "Simple Wildcard match with ?",
			selectors: []string{"region=eu-dc-?-north"},
			labels:    map[string]string{"region": "eu-dc-1-north"},
			expected:  true,
		},
		{
			name:      "Simple Wildcard mismatch with ?",
			selectors: []string{"region=eu-dc-?-north"},
			labels:    map[string]string{"region": "eu-dc-fail-north"},
			expected:  false,
		},
		{
			name:      "Multiple selectors (OR logic) - First matches",
			selectors: []string{"app=web;env=prod", "region=eu-*"},
			labels:    map[string]string{"app": "web", "env": "prod"},
			expected:  true,
		},
		{
			name:      "Multiple selectors (OR logic) - Second matches",
			selectors: []string{"app=web;env=prod", "region=eu-*"},
			labels:    map[string]string{"app": "web", "env": "staging", "region": "eu-west"},
			expected:  true,
		},
		{
			name:      "Multiple selectors (OR logic) - None matches",
			selectors: []string{"app=web;env=prod", "region=eu-*", "app=not-web"},
			labels:    map[string]string{"app": "api", "env": "staging", "region": "us-east"},
			expected:  false,
		},
		{
			name: "Multiple labels and multiple selectors (AND logic)",
			selectors: []string{
				"env=prod-*-dc-*;region=eu-*456", // this one should not match
				"simple=match",                   // this one should match
			}, // OR logic
			labels: map[string]string{
				"env":    "prod-23-dc-1something",
				"region": "eu-central-123",
				"simple": "match",
			},
			expected: true,
		},
		{
			name:      "Multiple labels and single selector(Selective AND)",
			selectors: []string{"env=prod"},
			labels: map[string]string{
				"env":    "prod",
				"region": "dc-23",
				"extra":  "value",
				"extra2": "value2",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pls := labelSelector{}
			require.NoError(t, pls.setSelections(tt.selectors))
			require.Equal(t, tt.expected, pls.matches(tt.labels))
		})
	}
}
