package crypto

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type tasmotaRespponse struct {
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

const tasmotaName = "tasmota"

// Tasmota firmware
// https://github.com/arendst/tasmota-Tasmota/wiki/Commands
// curl -s --connect-timeout 1 -m 1 http://192.168.210.121/cm?cmnd=Status%208 | jq .
type tasmota struct {
	serverBase
}

var tasmotaSampleConf = serverSampleConf

// Description of tasmota
func (*tasmota) Description() string {
	return "Read tasmota sensor's status"
}

// SampleConfig of tasmota
func (*tasmota) SampleConfig() string {
	return tasmotaSampleConf
}

func (*tasmota) getURL(address string) string {
	return "http://" + address + "/cm?cmnd=Status%208"
}

func (m *tasmota) serverGather(acc telegraf.Accumulator, i int, tags map[string]string) error {
	var reply tasmotaRespponse
	if err := getResponseSimple(m.getURL(m.getAddress(i)), &reply); err != nil {
		return err
	}

	fields := map[string]interface{}{
		"power":        reply.StatusSNS.ENERGY.Power,
		"power_factor": reply.StatusSNS.ENERGY.Factor,
		"current":      reply.StatusSNS.ENERGY.Current,
		"voltage":      reply.StatusSNS.ENERGY.Voltage,
	}
	acc.AddFields(tasmotaName, fields, tags)
	return nil
}

// Gather of tasmota
func (m *tasmota) Gather(acc telegraf.Accumulator) error {
	return m.minerGather(acc, m)
}

func init() {
	inputs.Add(tasmotaName, func() telegraf.Input { return &tasmota{} })
}
