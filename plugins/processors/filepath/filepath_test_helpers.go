package filepath

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

// structure holding the expected results for a given scenario according to their options
type pathResults struct {
	path              string
	inputTags         map[string]string
	inputFields       map[string]interface{}
	mustIncludeTags   map[string]string
	mustIncludeFields map[string]string
	*Options
}

func newOptions(basePath string) *Options {
	return &Options{
		BaseName: []BaseOpts{
			{
				Field: "baseField",
				Tag:   "baseTag",
			},
		},
		DirName: []BaseOpts{
			{
				Field: "dirField",
				Tag:   "dirTag",
			},
		},
		Stem: []BaseOpts{
			{
				Field: "stemField",
				Tag:   "stemTag",
			},
		},
		Clean: []BaseOpts{
			{
				Field: "cleanField",
				Tag:   "cleanTag",
			},
		},
		Rel: []RelOpts{
			{
				BaseOpts: BaseOpts{
					Field: "relField",
					Tag:   "relTag",
				},
				BasePath: basePath,
			},
		},
		ToSlash: []BaseOpts{
			{
				Field: "slashField",
				Tag:   "slashTag",
			},
		},
	}
}

func getMetricTags(path string) map[string]string {
	return map[string]string{
		"baseTag":  path,
		"dirTag":   path,
		"stemTag":  path,
		"cleanTag": path,
		"relTag":   path,
		"slashTag": path,
	}
}

func getMetricFields(path string) map[string]interface{} {
	return map[string]interface{}{
		"baseField":  path,
		"dirField":   path,
		"stemField":  path,
		"cleanField": path,
		"relField":   path,
		"slashField": path,
	}
}

func runTestOptionsApply(t *testing.T, tests []struct {
	name string
	pr   pathResults
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.pr.Options
			metric := testutil.MustMetric("testmetric",
				tt.pr.inputTags,
				tt.pr.inputFields,
				time.Now())

			got := o.Apply(metric)

			for k, v := range tt.pr.mustIncludeTags {
				gotTagValue, ok := got[0].GetTag(k)
				assert.True(t, ok, "Expected tag '%s' not found in processed metric '%s'",
					k, got[0].Name())
				assert.Equal(t, v, gotTagValue, "Expected value for tag '%s': %s, but got: %s",
					k, v, gotTagValue)
			}

			for k, v := range tt.pr.mustIncludeFields {
				gotFieldValue, ok := got[0].GetField(k)
				assert.True(t, ok, "Expected field '%s' not found in processed metric '%s'",
					k, got[0].Name())
				assert.Equal(t, v, gotFieldValue, "Expected value for field '%s': '%s', but got: '%s'",
					k, v, gotFieldValue)
			}

		})
	}
}
