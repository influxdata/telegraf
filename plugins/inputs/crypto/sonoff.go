package crypto

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type sonoffRespponse struct {
	StatusSNS struct {
		Time   string `json:"Time"`
		ENERGY struct {
			Total     float64 `json:"Total"`
			Yesterday float64 `json:"Yesterday"`
			Today     float64 `json:"Today"`
			Power     int     `json:"Power"`
			Factor    float64 `json:"Factor"`
			Voltage   int     `json:"Voltage"`
			Current   float64 `json:"Current"`
		} `json:"ENERGY"`
	} `json:"StatusSNS"`
}

const sonoffName = "sonoff"

// Sonoff Tasmota firmware
// https://github.com/arendst/Sonoff-Tasmota/wiki/Commands
// curl -s --connect-timeout 1 -m 1 http://192.168.210.121/cm?cmnd=Status%208 |jq .
type Sonoff struct {
	serverBase
}

var sonoffPowSampleConf = serverSampleConf

// Description of SonoffPow
func (*Sonoff) Description() string {
	return "Read Sonoff Pow's energy status"
}

// SampleConfig of SonoffPow
func (*Sonoff) SampleConfig() string {
	return sonoffPowSampleConf
}

func (*Sonoff) getURL(address string) string {
	return "http://" + address + "/cm?cmnd=Status%208"
}

func (m *Sonoff) serverGather(acc telegraf.Accumulator, i int, tags map[string]string) error {
	var reply sonoffRespponse
	if !getResponse(m.getURL(m.getAddress(i)), &reply, sonoffName) {
		return nil
	}

	fields := map[string]interface{}{
		"power":        reply.StatusSNS.ENERGY.Power,
		"power_factor": reply.StatusSNS.ENERGY.Factor,
		"current":      reply.StatusSNS.ENERGY.Current,
		"voltage":      reply.StatusSNS.ENERGY.Voltage,
	}
	acc.AddFields(sonoffName, fields, tags)
	return nil
}

// Gather of SonoffPow
func (m *Sonoff) Gather(acc telegraf.Accumulator) error {
	return m.minerGather(acc, m)
}

func init() {
	inputs.Add(sonoffName, func() telegraf.Input { return &Sonoff{} })
}
