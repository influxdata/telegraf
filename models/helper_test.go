package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test that ShouldApplyPrefixToMetric is applied properly
// with respect to ConsecutiveNamePrefixLimit option
func TestRunningOutput_ShouldApplyPrefixToMetric(t *testing.T) {
	table := []struct {
		Description       string
		Name              string
		Prefix            string
		Limit             int
		ShouldApplyPrefix bool
	}{
		{
			Description:       "ConsecutiveNamePrefixLimit unset (default 0)",
			Name:              "prefix_foo",
			Prefix:            "prefix_",
			Limit:             0,
			ShouldApplyPrefix: true,
		},
		{
			Description:       "ConsecutiveNamePrefixLimit set, metric does not begin with prefix",
			Name:              "prefix_foo",
			Prefix:            "notprefix",
			Limit:             1,
			ShouldApplyPrefix: true,
		},
		{
			Description:       "ConsecutiveNamePrefixLimit set, metric name shorter than prefix",
			Name:              "foo",
			Prefix:            "prefix_",
			Limit:             1,
			ShouldApplyPrefix: true,
		},
		{
			Description:       "ConsecutiveNamePrefixLimit set, metric begins with prefix, limit not reached",
			Name:              "prefix_foo",
			Prefix:            "prefix_",
			Limit:             2,
			ShouldApplyPrefix: true,
		},
		{
			Description:       "ConsecutiveNamePrefixLimit set, metric begins with prefix, limit reached",
			Name:              "prefix_prefix_foo",
			Prefix:            "prefix_",
			Limit:             2,
			ShouldApplyPrefix: false,
		},
	}

	for _, test := range table {
		t.Run(test.Description, func(t *testing.T) {
			shouldApplyPrefix := shouldApplyPrefixToMetric(test.Name, test.Prefix, test.Limit)
			assert.Equal(t, shouldApplyPrefix, test.ShouldApplyPrefix)
		})
	}
}
