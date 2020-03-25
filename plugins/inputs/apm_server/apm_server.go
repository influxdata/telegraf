package apm_server

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// APM Server is a input plugin that listens for requests sent by Elastic APM Agents.
type APMServer struct {
	ServiceAddress string `toml:"service_address"`
}

func (s *APMServer) Description() string {
	return "APM Server is a input plugin that listens for requests sent by Elastic APM Agents."
}

func (s *APMServer) SampleConfig() string {
	return `
   ## Address and port to list APM Agents
   service_address = ":8200"
`
}

func (s *APMServer) Init() error {
	return nil
}

func (s *APMServer) Gather(acc telegraf.Accumulator) error {
	acc.AddFields("apm_server", map[string]interface{}{"service_address": s.ServiceAddress}, nil)

	return nil
}

func init() {
	inputs.Add("apm_server", func() telegraf.Input {
		return &APMServer{
			ServiceAddress: ":8200",
		}
	})
}
