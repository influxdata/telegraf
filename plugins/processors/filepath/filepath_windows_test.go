package filepath

import (
	"testing"
)

var (
	samplePath       = "c:\\my\\test\\\\c\\..\\path\\file.log"
	smokePathResults = pathResults{
		path:        samplePath,
		inputTags:   getMetricTags(samplePath),
		inputFields: getMetricFields(samplePath),
		Options:     newOptions("c:\\my\\test\\"),
		mustIncludeTags: map[string]string{
			"baseTag":  "file.log",
			"dirTag":   "c:\\my\\test\\path",
			"stemTag":  "file",
			"cleanTag": "c:\\my\\test\\path\\file.log",
			"relTag":   "path\\file.log",
			"slashTag": "c:/my/test//c/../path/file.log",
		},
		mustIncludeFields: map[string]string{
			"baseField":  "file.log",
			"dirField":   "c:\\my\\test\\path",
			"stemField":  "file",
			"cleanField": "c:\\my\\test\\path\\file.log",
			"relField":   "path\\file.log",
			"slashField": "c:/my/test//c/../path/file.log",
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
	}
	runTestOptionsApply(t, tests)
}
