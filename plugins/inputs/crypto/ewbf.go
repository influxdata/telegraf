package crypto

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ewbfResponse struct {
	//Version string  `json:"jsonrpc"`
	ID     uint64    `json:"id"`
	Error  string    `json:"error,omitempty"`
	Result [9]string `json:"result,omitempty"`
}

const (
	ewbfName    = "ewbf"
	ewbfRequest = "{\"id\":1,\"method\":\"getstat\"}\n"
)

var ewbfSampleConf = `
  #interval = "1s"
  ## Miner servers addresses and names
  servers = ["localhost:42000"]
`

// EWBF miner
type EWBF struct {
	serverBase
}

// Description of EWBF
func (*EWBF) Description() string {
	return "Read EWBF's mining status"
}

// SampleConfig of EWBF
func (*EWBF) SampleConfig() string {
	return ewbfSampleConf
}

func (m *EWBF) getAlgorithm(i int) string {
	return "equihash"
}

func (m *EWBF) serverGather(acc telegraf.Accumulator, i int, tags map[string]string) error {
	// var reply ewbfResponse
	// err := jsonReader(address, ewbfRequest, &reply)
	// if err != nil {
	// 	return err
	// }
	// if len(reply.Error) != 0 {
	// 	return errors.New(reply.Error)
	// }

	// results := reply.Result

	// tags := map[string]string{
	// 	"name":    name,
	// 	"address": address,
	// }
	return nil
}

// Gather for EWBF
func (m *EWBF) Gather(acc telegraf.Accumulator) error {
	return m.minerGather(acc, m)
}

func init() {
	inputs.Add(ewbfName, func() telegraf.Input { return &EWBF{} })
}
