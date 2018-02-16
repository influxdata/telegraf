// +build linux,arm

package dht22

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/d2r2/go-dht"
)

var DHTConfig = `
  ## Set the GPIO pin
  pin = 14
  ## use boostPerfFlag
  boost = false
  ## how many times to retry for a good reading.
  retry = 10
`

type DHT struct {
	Pin   int
	Boost bool
	Retry int
}

func (*DHT) Description() string {
	return "Monitor DHT22 connected to GPIO"
}

func (*DHT) SampleConfig() string {
	return DHTConfig
}

func (s *DHT) Gather(acc telegraf.Accumulator) error {
	temperature, humidity, retries, err := dht.ReadDHTxxWithRetry(dht.DHT22, s.Pin, s.Boost, s.Retry)
	if err != nil {
		return err
	}
	fields := make(map[string]interface{})
	fields["temperature"] = temperature
	fields["humidity"] = humidity
	fields["retries"] = retries

	acc.AddFields("dht22", fields, nil)

	return nil
}

func init() {
	inputs.Add("dht22", func() telegraf.Input {
		return &DHT{}
	})
}
