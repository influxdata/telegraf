package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSelectors(t *testing.T) {
	tests := []struct {
		name      string
		selectors []string
		want      [][]string
		wantErr   bool
	}{
		{
			name:      "Valid selectors",
			selectors: []string{"env=prod;app=api;kind=metrics", "region=dc-23;node=host-1;service=web"},
			want:      [][]string{{"env=prod", "app=api", "kind=metrics"}, {"region=dc-23", "node=host-1", "service=web"}},
		},
		{
			name:      "Empty selector",
			selectors: []string{""},
			wantErr:   true,
		},
		{
			name:      "empty key",
			selectors: []string{"=prod", "region=dc-23"},
			wantErr:   true,
		},
		{
			name:      "empty value",
			selectors: []string{"env=", "region=dc-23"},
			wantErr:   true,
		},
		{
			name:      "Invalid format",
			selectors: []string{"envprod"}, // missing '='
			wantErr:   true,
		},
		{
			name:      "Duplicate key",
			selectors: []string{"env=prod;app=api;env=staging"}, // duplicate 'env'
			wantErr:   true,
		},
		{
			name:      "duplicate keys but in different selectors",
			selectors: []string{"env=prod", "env=staging"}, // different selectors but same key
			want:      [][]string{{"env=prod"}, {"env=staging"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalSelectorFlags := SelectorFlags
			defer func() {
				SelectorFlags = originalSelectorFlags
				parsedSelectors = nil // Reset parsedSelectors after each test
			}()
			SelectorFlags = tt.selectors
			err := ParseSelectors()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.ElementsMatch(t, tt.want, parsedSelectors)
		})
	}
}

func TestShouldPluginRun(t *testing.T) {
	tests := []struct {
		name      string
		selectors [][]string
		labels    map[string]string
		want      bool
	}{
		{
			name:      "[Backward Compatibility] No selectors, should run",
			selectors: nil,
			labels:    map[string]string{"env": "prod"},
			want:      true,
		},
		{
			name:      "[Backward Compatibility] No labels, should run",
			selectors: [][]string{{"env=prod"}},
			labels:    nil,
			want:      true,
		},
		{
			name:      "Simple exact match",
			selectors: [][]string{{"env=prod"}},
			labels:    map[string]string{"env": "prod"},
			want:      true,
		},
		{
			name:      "Simple mismatch",
			selectors: [][]string{{"env=prod"}},
			labels:    map[string]string{"env": "dev"},
			want:      false,
		},
		{
			name:      "extra labels ignored",
			selectors: [][]string{{"env=prod"}},
			labels:    map[string]string{"env": "prod", "region": "us-east"},
			want:      true,
		},
		{
			name:      "AND inside selector (all match)",
			selectors: [][]string{{"env=prod", "region=dc-23"}},
			labels:    map[string]string{"env": "prod", "region": "dc-23"},
			want:      true,
		},
		{
			name:      "AND inside selector (partial match fail)",
			selectors: [][]string{{"env=prod", "region=dc-23"}},
			labels:    map[string]string{"env": "prod", "region": "dc-24"},
			want:      false,
		},
		{
			name:      "Simple Wildcard match",
			selectors: [][]string{{"region=dc-*"}},
			labels:    map[string]string{"region": "dc-23"},
			want:      true,
		},
		{
			name:      "Simple Wildcard no match",
			selectors: [][]string{{"region=us-*"}},
			labels:    map[string]string{"region": "eu-1"},
			want:      false,
		},
		{
			name:      "Simple Wildcard match with ?",
			selectors: [][]string{{"region=eu-dc-?-north"}},
			labels:    map[string]string{"region": "eu-dc-1-north"},
			want:      true,
		},
		{
			name:      "Simple Wildcard mismatch with ?",
			selectors: [][]string{{"region=eu-dc-?-north"}},
			labels:    map[string]string{"region": "eu-dc-fail-north"},
			want:      false,
		},
		{
			name:      "Multiple selectors (OR logic) - First matches",
			selectors: [][]string{{"app=web", "env=prod"}, {"region=eu-*"}},
			labels:    map[string]string{"app": "web", "env": "prod"},
			want:      true,
		},
		{
			name:      "Multiple selectors (OR logic) - Second matches",
			selectors: [][]string{{"app=web", "env=prod"}, {"region=eu-*"}},
			labels:    map[string]string{"app": "web", "env": "staging", "region": "eu-west"},
			want:      true,
		},
		{
			name:      "Multiple selectors (OR logic) - None matches",
			selectors: [][]string{{"app=web", "env=prod"}, {"region=eu-*", "app=api"}},
			labels:    map[string]string{"app": "api", "env": "staging", "region": "us-east"},
			want:      false,
		},
		{
			name: "Multiple labels and multiple selectors (AND logic)",
			selectors: [][]string{
				{"env=prod-*-dc-*", "region=eu-*456"}, // this one should not match
				{"simple=match"},                      // this one should match
			}, // OR logic
			labels: map[string]string{
				"env":    "prod-23-dc-1something",
				"region": "eu-central-123",
				"simple": "match",
			},
			want: true,
		},
		{
			name:      "Multiple labels and single selector(Selective AND)",
			selectors: [][]string{{"env=prod"}},
			labels: map[string]string{
				"env":    "prod",
				"region": "dc-23",
				"extra":  "value",
				"extra2": "value2",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldRun := shouldPluginRun(tt.selectors, tt.labels)
			require.Equal(t, tt.want, shouldRun)
		})
	}
}
