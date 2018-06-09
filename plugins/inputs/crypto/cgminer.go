package crypto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type status struct {
	Status      string `json:"STATUS"`
	When        uint64 `json:"When"`
	Code        int    `json:"Code"`
	Msg         string `json:"Msg"`
	Description string `json:"Description"`
}

type version struct {
	CGMiner     string `json:"CGMiner"`
	API         string `json:"API"`
	Miner       string `json:"Miner"`
	CompileTime string `json:"CompileTime"`
	Type        string `json:"Type"`
}

type versionResponse struct {
	Status  []status
	Version []version
	ID      int
}

type stats struct {
	STATS   int    `json:"STATS,omitempty"`
	ID      string `json:"ID,omitempty"`
	ChainID int    `json:"Chain ID,omitempty"`
	Elapsed int    `json:"Elapsed,omitempty"`
}

type statsResponse struct {
	Status []status
	Stats  []stats
	ID     int
}

type pool struct {
	POOL          int    `json:"POOL"`
	URL           string `json:"URL"`
	Status        string `json:"Status"`
	Priority      int    `json:"Priority"`
	Accepted      int    `json:"Accepted"`
	Rejected      int    `json:"Rejected"`
	User          string `json:"User"`
	StratumActive bool   `json:"Stratum Active"`
}

type poolResponse struct {
	Status []status
	Pools  []pool
	ID     int
}

type summary struct {
	Elapsed uint64
	GHS5s   float64 `json:"GHS 5s,string"`
	MHS5s   float64 `json:"MHS 5s"`
	MHS1m   float64 `json:"MHS 1m"`
}

type summaryResponse struct {
	Status  []status
	Summary []summary
	ID      int
}

type devs struct {
	ASC         int     `json:"ASC"`
	Name        string  `json:"Name"`
	ID          int     `json:"ID"`
	Enabled     string  `json:"Enabled"`
	Status      string  `json:"Status"`
	Temperature float64 `json:"Temperature"`
	MHSAv       float64 `json:"MHS av"`
	MHS5s       float64 `json:"MHS 5s"`
	MHS1m       float64 `json:"MHS 1m"`
	MHS5m       float64 `json:"MHS 5m"`
	MHS15m      float64 `json:"MHS 15m"`
	Accepted    int     `json:"Accepted"`
	Rejected    int     `json:"Rejected"`
}

type devsResponse struct {
	Status []status
	Devs   []devs
	ID     int
}

type cgMinerResponse struct {
	Stats   []statsResponse
	Pools   []poolResponse
	Version []versionResponse
	Summary []summaryResponse
	Devs    []devsResponse
	ID      int
}

const (
	cgMinerName    = "cgminer"
	cgMinerRequest = "{\"command\":\"stats+pools+version+summary+devs\"}"
)

var cgMinerSampleConf = `
  interval = "1m"
  ## Miner servers addresses and names
  servers    = ["localhost:4028"]
  names      = ["Rig1"]
  algorithms = ["sha256"]
`

// CGMiner API docs: https://github.com/ckolivas/cgminer/blob/master/API-README
// nc 127.0.0.1 4028 <<< '{"command":"stats+pools+summary"}' | sed "s/{[^}]+}{/{/" | jq .
type CGMiner struct {
	serverBase
	Algorithms []string `toml:"algorithms"`
}

// Description of CGMiner
func (*CGMiner) Description() string {
	return "Read CGMiner's mining status"
}

// SampleConfig of CGMiner
func (*CGMiner) SampleConfig() string {
	return cgMinerSampleConf
}

func (m *CGMiner) getAlgorithm(i int) string {
	return m.Algorithms[i]
}

// drop the last '\0' byte
// drop the version info ends with "}{" in response to "stats" command
func jsonWorkaround(buf bytes.Buffer) []byte {
	real := buf.Bytes()[:buf.Len()-1]
	re := regexp.MustCompile("{[^}]+}{")
	return re.ReplaceAll(real, []byte("{"))
}

func (m *CGMiner) fieldChainGather(response []byte, acc telegraf.Accumulator, tags map[string]string) {
	var responseMap map[string]*json.RawMessage
	json.Unmarshal(response, &responseMap)
	var statsMap []map[string]*json.RawMessage
	json.Unmarshal(*responseMap["stats"], &statsMap)
	var statsMap2 []map[string]*json.RawMessage
	json.Unmarshal(*statsMap[0]["STATS"], &statsMap2)
	stats := statsMap2[0]
	var row int
	json.Unmarshal(*stats["temp_num"], &row)
	for i := 1; i <= row; i++ {
		tags["source"] = CHAIN.String()
		tags["unit"] = fmt.Sprintf("%d", i)

		var rate string
		json.Unmarshal(*stats[fmt.Sprintf("chain_rate%d", i)], &rate)
		hash, _ := strconv.ParseFloat(rate, 64)

		// Normally, the PCB temp should be 40℃-85℃. The chips temp should be 85℃-115℃.
		var temp, tempPcb int
		json.Unmarshal(*stats[fmt.Sprintf("temp%d", i)], &tempPcb)
		json.Unmarshal(*stats[fmt.Sprintf("temp2_%d", i)], &temp)
		var acs string
		json.Unmarshal(*stats[fmt.Sprintf("chain_acs%d", i)], &acs)
		fields := map[string]interface{}{
			"hashrate":        uint64(hash * 1000000000.0), // was in GH/s
			"temperature_pcb": tempPcb,
			"temperature":     temp,
			"failed":          strings.Count(acs, "x"),
		}
		acc.AddFields(cgMinerName, fields, tags)
	}
	json.Unmarshal(*stats["fan_num"], &row)
	for i := 1; i <= row; i++ {
		tags["source"] = FAN.String()
		tags["unit"] = fmt.Sprintf("%d", i)
		var t int
		json.Unmarshal(*stats[fmt.Sprintf("fan%d", i)], &t)
		fields := map[string]interface{}{
			"fan": t,
		}
		acc.AddFields(cgMinerName, fields, tags)
	}
}

func (m *CGMiner) arrayChainGather(devsArray []devs, acc telegraf.Accumulator, tags map[string]string) {
	for _, dev := range devsArray {
		tags["source"] = CHAIN.String()
		tags["unit"] = fmt.Sprintf("%d", dev.ID)
		failed := 0
		if dev.Status != "Alive" {
			failed = 1
		}
		total := dev.Accepted + dev.Rejected
		rate := 100
		if total != 0 {
			rate = 100 * dev.Accepted / total
		}
		fields := map[string]interface{}{
			"hashrate":        uint64(dev.MHS1m * 1000000.0), // was in MH/s
			"temperature":     int(dev.Temperature),
			"failed":          failed,
			"shares_total":    total,
			"shares_accepted": dev.Accepted,
			"shares_rejected": dev.Rejected,
			"shares_rate":     rate,
		}
		acc.AddFields(cgMinerName, fields, tags)
	}
}

func (m *CGMiner) serverGather(acc telegraf.Accumulator, i int, tags map[string]string) error {
	var buf bytes.Buffer
	if err := jsonReader(m.getAddress(i), cgMinerRequest, &buf); err != nil {
		log.Println(cgMinerName+" error: ", err)
		return nil
	}

	response := jsonWorkaround(buf)
	var reply cgMinerResponse
	err := json.Unmarshal(response, &reply)

	statusS := reply.Stats[0].Status[0]
	statusP := reply.Pools[0].Status[0]
	statusV := reply.Version[0].Status[0]
	statusY := reply.Summary[0].Status[0]
	statusD := reply.Devs[0].Status[0]
	pools := reply.Pools[0].Pools
	version := reply.Version[0].Version[0]
	summary := reply.Summary[0].Summary[0]
	devs := reply.Devs[0].Devs
	if err != nil || statusS.Status != "S" || statusP.Status != "S" || statusV.Status != "S" || statusY.Status != "S" {
		log.Println(cgMinerName+" error: ", err, statusS.Status, statusP.Status, statusV.Status, statusY.Status)
		return nil
	}

	tags["algorithm"] = m.getAlgorithm(i)
	tags["source"] = MINER.String()
	tags["version"] = version.CGMiner
	tags["api"] = version.API

	fields := map[string]interface{}{
		"uptime": summary.Elapsed, // in seconds
	}
	if summary.GHS5s != 0.0 {
		fields["hashrate"] = uint64(summary.GHS5s * 1000000000.0) // was in GH/s
	} else if summary.MHS1m != 0.0 {
		fields["hashrate"] = uint64(summary.MHS1m * 1000000.0) // was in MH/s
	}
	acc.AddFields(cgMinerName, fields, tags)
	delete(tags, "chain")
	delete(tags, "version")
	delete(tags, "api")

	tags["source"] = POOL.String()
	for _, pool := range pools {
		if pool.StratumActive {
			tags["pool"] = pool.URL
			total := pool.Accepted + pool.Rejected
			rate := 100
			if total != 0 {
				rate = 100 * pool.Accepted / total
			}
			fields := map[string]interface{}{
				"shares_total":    total,
				"shares_accepted": pool.Accepted,
				"shares_rejected": pool.Rejected,
				"shares_rate":     rate,
			}
			acc.AddFields(cgMinerName, fields, tags)
		}
	}
	delete(tags, "pool")

	if statusD.Status != "S" {
		m.fieldChainGather(response, acc, tags)
	} else {
		m.arrayChainGather(devs, acc, tags)
	}

	return nil
}

// Gather for CGMiner
func (m *CGMiner) Gather(acc telegraf.Accumulator) error {
	return m.minerGather(acc, m)
}

func init() {
	inputs.Add(cgMinerName, func() telegraf.Input { return &CGMiner{} })
}
