package processes

import _ "embed"

//go:embed sample.conf
var sampleConfig string

func (*Processes) SampleConfig() string {
	return sampleConfig
}
