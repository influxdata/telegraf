// +build !windows

package filepath

import (
	"testing"
)

var (
	samplePath = "/my/test//c/../path/file.log"

	smokePathResults = pathResults{
		path:        samplePath,
		inputTags:   getMetricTags(samplePath),
		inputFields: getMetricFields(samplePath),
		Options:     newOptions("/my/test/"),
		mustIncludeTags: map[string]string{
			"baseTag":  "file.log",
			"dirTag":   "/my/test/path",
			"stemTag":  "file",
			"cleanTag": "/my/test/path/file.log",
			"relTag":   "path/file.log",
			"slashTag": "/my/test//c/../path/file.log",
		},
		mustIncludeFields: map[string]string{
			"baseField":  "file.log",
			"dirField":   "/my/test/path",
			"stemField":  "file",
			"cleanField": "/my/test/path/file.log",
			"relField":   "path/file.log",
			"slashField": "/my/test//c/../path/file.log",
		},
	}

	destPathResults = pathResults{
		path: samplePath,
		inputTags: map[string]string{
			"sourcePath": samplePath,
		},
		inputFields: map[string]interface{}{
			"sourcePath": samplePath,
		},
		Options: &Options{
			BaseName: []BaseOpts{
				{
					Field: "sourcePath",
					Tag:   "sourcePath",
					Dest:  "basePath",
				},
			},
		},
		mustIncludeTags: map[string]string{
			"sourcePath": samplePath,
			"basePath":   "file.log",
		},
		mustIncludeFields: map[string]string{
			"sourcePath": samplePath,
			"basePath":   "file.log",
		},
	}
)

func TestOptions_Apply(t *testing.T) {
	tests := []struct {
		name string
		pr   pathResults
	}{
		{
			name: "Smoke Test",
			pr:   smokePathResults,
		},
		{
			name: "Dest Test",
			pr:   destPathResults,
		},
	}
	runTestOptionsApply(t, tests)
}
