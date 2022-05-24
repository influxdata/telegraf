package processes

import _ "embed"

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embedd the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

func (*Processes) SampleConfig() string {
	return sampleConfig
}
