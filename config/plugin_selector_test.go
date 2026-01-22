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
		{
			name:       "empty value",
			selections: []string{"env=prod;app="},
			wantErr:    true,
		},
		{
			name:           "contains only _,-,*,?",
			selections:     []string{"env=__-_*?"},
			expectedGroups: []int{1},
		},
		{
			name:           "regex check- character set",
			selections:     []string{"Env123=456prOd;123=*456?", "p0-p1_p2.=_?*Env234_-"},
			expectedGroups: []int{2, 1},
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

func TestKeyRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid
		{name: "simple", input: "abc"},
		{name: "alphanumeric", input: "abc123"},
		{name: "with-dash", input: "abc-123"},
		{name: "with-dot", input: "abc.123"},
		{name: "with-underscore", input: "abc_123"},
		{name: "mixed", input: "A_z-9.X"},
		{name: "single-char", input: "a"},
		{name: "long-key", input: "abc.def-ghi_jkl.mno_123"},
		{name: "starts-with-dot", input: ".abc"},
		{name: "ends-with-dot", input: "abc."},
		{name: "two-dots", input: "a..b"},
		// Invalid
		{name: "empty", input: "", wantErr: true},
		{name: "wildcard-star", input: "abc*", wantErr: true},
		{name: "wildcard-question", input: "abc?", wantErr: true},
		{name: "space", input: "abc def", wantErr: true},
		{name: "unicode", input: "ümlaut", wantErr: true},
		{name: "symbols", input: "abc$", wantErr: true},
		{name: "slash", input: "abc/def", wantErr: true},
		{name: "colon", input: "abc:def", wantErr: true},
		{name: "comma", input: "abc,def", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSelectionKeyValuePairs(tt.input, "value")
			if tt.wantErr {
				require.Error(t, err)
			}
		})
	}
}

func TestSelectorValueRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid
		{name: "simple", input: "abc"},
		{name: "with-wildcards", input: "a*b?c"},
		{name: "only-star", input: "*"},
		{name: "only-question", input: "?"},
		{name: "mixed-wildcards", input: "*a?b*c*"},
		{name: "alphanumeric", input: "abc123"},
		{name: "combo", input: "A_z-9.X*?foo"},
		{name: "dash-dot-underscore-wildcards", input: "a_b-c.d*?"},
		{name: "ends-with-wildcard", input: "abc*"},
		{name: "starts-with-wildcard", input: "*abc"},
		{name: "wildcard-middle", input: "ab*cd?ef"},
		{name: "long-value", input: "abc.def-ghi*jkl?mno_123"},

		// Invalid
		{name: "empty", input: "", wantErr: true},
		{name: "space", input: "abc def", wantErr: true},
		{name: "unicode", input: "ümlaut", wantErr: true},
		{name: "control-char", input: "abc\x00def", wantErr: true},
		{name: "symbols", input: "abc$", wantErr: true},
		{name: "slash", input: "abc/def", wantErr: true},
		{name: "colon", input: "abc:def", wantErr: true},
		{name: "comma", input: "abc,def", wantErr: true},
		{name: "plus", input: "abc+def", wantErr: true},
		{name: "pipe", input: "abc|def", wantErr: true},
		{name: "caret", input: "abc^def", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSelectionKeyValuePairs("key", tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLabelValueRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid
		{name: "simple", input: "abc"},
		{name: "alphanumeric", input: "abc123"},
		{name: "with-dash", input: "abc-123"},
		{name: "with-dot", input: "abc.123"},
		{name: "with-underscore", input: "abc_123"},
		{name: "mixed", input: "A_z-9.X"},
		{name: "single-char", input: "a"},
		{name: "long", input: "abc.def-ghi_jkl.mno_123"},
		{name: "starts-with-dot", input: ".abc"},
		{name: "ends-with-dot", input: "abc."},
		{name: "two-dots", input: "a..b"},

		// Invalid
		{name: "empty", input: "", wantErr: true},
		{name: "wildcard-star", input: "abc*", wantErr: true},
		{name: "wildcard-question", input: "abc?", wantErr: true},
		{name: "space", input: "abc def", wantErr: true},
		{name: "unicode", input: "香港", wantErr: true},
		{name: "symbols", input: "abc$", wantErr: true},
		{name: "slash", input: "abc/def", wantErr: true},
		{name: "comma", input: "abc,def", wantErr: true},
		{name: "plus", input: "abc+def", wantErr: true},
		{name: "caret", input: "abc^def", wantErr: true},
		{name: "pipe", input: "abc|def", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckLabelKeyValuePairs("key", tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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
		{
			name:      "complex charset match",
			selectors: []string{"Env-123=prOd456;Env_456=_123prOd-", "Env-123=?456*;Env_456=_???pr0D*"},
			labels: map[string]string{
				"Env-123": "X456",
				"Env_456": "____pr0D123",
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
