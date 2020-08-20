/*
 * Modbus Gateway plugin
 * Developed by Christopher Piggott under the InfluxData CLA
 * August, 2020
 */

package modbusgw

import (
	mb "github.com/goburrow/modbus"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ModbusGateway struct {
	Name     string    `toml:"name"`
	Gateway  string    `toml:"gateway"`
	Requests []Request `toml:"requests"`

	Timeout     internal.Duration `toml:"timeout"`
	tcpHandler  *mb.TCPClientHandler
	isConnected bool
	client      mb.Client

	Log telegraf.Logger
}

type Request struct {
	Unit        uint8      `toml:"unit"`
	Address     uint16     `toml:"address"`
	Count       uint16     `toml:"count"`
	Type        string     `toml:"type"`
	Measurement string     `toml:"measurement"`
	Tags        []string   `toml:"tags"`
	Fields      []FieldDef `toml:"fields"`
}

type FieldDef struct {
	Name   string  `toml:"name"`
	Omit   bool    `toml:"omit"`
	Scale  float32 `toml:"scale"`
	Offset float32 `toml:"offset"`
	Type   string  `toml:"type"`
}

const description = `Expert mode MODBUS telegraf input`

func (m *ModbusGateway) Description() string {
	return description
}

// Add this plugin to telegraf
func init() {
	inputs.Add("modbusgw", func() telegraf.Input { return &ModbusGateway{} })
}
