//go:generate ../../../tools/readme_config_includer/generator
package processes

import _ "embed"

//go:embed sample.conf
var sampleConfig string

func (*Processes) SampleConfig() string {
	return sampleConfig
}
