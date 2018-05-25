package crypto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

type stats struct {
	CGMiner     string `json:"CGMiner,omitempty"`
	Miner       string `json:"Miner,omitempty"`
	CompileTime string `json:"CompileTime,omitempty"`
	Type        string `json:"Type,omitempty"`
	// above this line is the mix from json bug
	STATS          int     `json:"STATS,omitempty"`
	ID             string  `json:"ID,omitempty"`
	ChainID        int     `json:"Chain ID,omitempty"`
	Elapsed        int     `json:"Elapsed,omitempty"`
	Calls          int     `json:"Calls,omitempty"`
	Wait           float64 `json:"Wait,omitempty"`
	Max            float64 `json:"Max,omitempty"`
	Min            float64 `json:"Min,omitempty"`
	NumChips       int     `json:"Num chips,omitempty"`
	NumCores       int     `json:"Num cores,omitempty"`
	NumActiveChips int     `json:"Num active chips,omitempty"`
	GHS5S          float64 `json:"GHS 5s,omitempty,string"`
	GHSAv          float64 `json:"GHS av,omitempty"`
	MinerCount     int     `json:"miner_count,omitempty"`
	Frequency      int     `json:"frequency,omitempty,string"`
	FanNum         int     `json:"fan_num,omitempty"`
	FanDuty        int     `json:"fan duty,omitempty"`
	Fan1           int     `json:"fan1,omitempty"`
	Fan2           int     `json:"fan2,omitempty"`
	TempNum        int     `json:"temp_num,omitempty"`
	Temp           int     `json:"Temp,omitempty"`
	Temp1          int     `json:"temp1,omitempty"`
	Temp2          int     `json:"temp2,omitempty"`
	Temp3          int     `json:"temp3,omitempty"`
	Temp4          int     `json:"temp4,omitempty"`
	Temp21         int     `json:"temp2_1,omitempty"`
	Temp22         int     `json:"temp2_2,omitempty"`
	Temp23         int     `json:"temp2_3,omitempty"`
	Temp24         int     `json:"temp2_4,omitempty"`
	TempMax        int     `json:"temp_max,omitempty"`
	DeviceHardware float64 `json:"Device Hardware%,omitempty"`
	NoMatchingWork int     `json:"no_matching_work,omitempty"`
	ChainAcn1      int     `json:"chain_acn1,omitempty"`
	ChainAcn2      int     `json:"chain_acn2,omitempty"`
	ChainAcn3      int     `json:"chain_acn3,omitempty"`
	ChainAcn4      int     `json:"chain_acn4,omitempty"`
	ChainAcs1      string  `json:"chain_acs1,omitempty"`
	ChainAcs2      string  `json:"chain_acs2,omitempty"`
	ChainAcs3      string  `json:"chain_acs3,omitempty"`
	ChainAcs4      string  `json:"chain_acs4,omitempty"`
	ChainHw1       int     `json:"chain_hw1,omitempty"`
	ChainHw2       int     `json:"chain_hw2,omitempty"`
	ChainHw3       int     `json:"chain_hw3,omitempty"`
	ChainHw4       int     `json:"chain_hw4,omitempty"`
	ChainRate1     string  `json:"chain_rate1,omitempty"`
	ChainRate2     string  `json:"chain_rate2,omitempty"`
	ChainRate3     string  `json:"chain_rate3,omitempty"`
	ChainRate4     string  `json:"chain_rate4,omitempty"`
}

type statsResponse struct {
	Status []status
	Stats  []stats
	ID     int
}

type pool struct {
	POOL                int     `json:"POOL"`
	URL                 string  `json:"URL"`
	Status              string  `json:"Status"`
	Priority            int     `json:"Priority"`
	Quota               int     `json:"Quota"`
	LongPoll            string  `json:"Long Poll"`
	Getworks            int     `json:"Getworks"`
	Accepted            int     `json:"Accepted"`
	Rejected            int     `json:"Rejected"`
	Works               int     `json:"Works"`
	Discarded           int     `json:"Discarded"`
	Stale               int     `json:"Stale"`
	GetFailures         int     `json:"Get Failures"`
	RemoteFailures      int     `json:"Remote Failures"`
	User                string  `json:"User"`
	LastShareTime       string  `json:"Last Share Time"`
	Diff                string  `json:"Diff"`
	Diff1Shares         int     `json:"Diff1 Shares"`
	ProxyType           string  `json:"Proxy Type"`
	Proxy               string  `json:"Proxy"`
	DifficultyAccepted  float64 `json:"Difficulty Accepted"`
	DifficultyRejected  float64 `json:"Difficulty Rejected"`
	DifficultyStale     float64 `json:"Difficulty Stale"`
	LastShareDifficulty float64 `json:"Last Share Difficulty"`
	WorkDifficulty      float64 `json:"Work Difficulty"`
	HasStratum          bool    `json:"Has Stratum"`
	StratumActive       bool    `json:"Stratum Active"`
	StratumURL          string  `json:"Stratum URL"`
	StratumDifficulty   int     `json:"Stratum Difficulty"`
	HasVmask            bool    `json:"Has Vmask"`
	HasGBT              bool    `json:"Has GBT"`
	BestShare           string  `json:"Best Share"`
	PoolRejected        float64 `json:"Pool Rejected%"`
	PoolStale           float64 `json:"Pool Stale%"`
	CurrentBlockHeight  int     `json:"Current Block Height"`
	CurrentBlockVersion int     `json:"Current Block Version"`
}

type poolResponse struct {
	Status []status
	Pools  []pool
	ID     int
}

/*
type summary struct {
	Elapsed            uint64
	GHS5s              float64 `json:"GHS 5s,string"`
	GHSav              float64 `json:"GHS av"`
	MHSav              float64 `json:"MHS av"`
	MHS5s              float64 `json:"MHS 5s"`
	MHS1m              float64 `json:"MHS 1m"`
	MHS5m              float64 `json:"MHS 5m"`
	MHS15m             float64 `json:"MHS 15m"`
	FoundBlocks        int     `json:"Found Blocks"`
	Getworks           uint64
	Accepted           string
	Rejected           string
	HardwareErrors     int `json:"Hardware Errors"`
	Utility            float64
	Discarded          uint64
	Stale              int
	GetFailures        int     `json:"Get Failures"`
	LocalWork          uint64  `json:"Local Work"`
	RemoteFailures     int     `json:"Remote Failures"`
	NetworkBlocks      int     `json:"Network Blocks"`
	TotalMH            uint64  `json:"Total MH"`
	WorkUtility        float64 `json:"Work Utility"`
	DifficultyAccepted float64 `json:"Difficulty Accepted"`
	DifficultyRejected float64 `json:"Difficulty Rejected"`
	DifficultyStale    int     `json:"Difficulty Stale"`
	BestShare          uint64  `json:"Best Share"`
	DeviceHardware     float64 `json:"Device Hardware%"`
	DeviceRejected     float64 `json:"Device Rejected%"`
	PoolRejected       float64 `json:"Pool Rejected%"`
	PoolStale          int     `json:"Pool Stale%"`
	Lastgetwork        uint64  `json:"Last getwork"`
}

type summaryResponse struct {
	Status  []status
	Summary []summary
	ID      int
}
*/

type cgMinerResponse struct {
	Stats []statsResponse
	Pools []poolResponse
	//Summary []summaryResponse
	ID int
}

const (
	cgMinerName = "cgminer"
	//cgMinerRequest = "{\"command\":\"stats+pools+summary\"}"
	cgMinerRequest = "{\"command\":\"stats+pools\"}"
)

var cgMinerSampleConf = `
  interval = "1m"
  ## Miner servers addresses and names
  servers    = ["localhost:4028"]
  names      = ["Rig1"]
  algorithms = ["sha256"]
`

// CGMiner API docs: https://github.com/ckolivas/cgminer/blob/master/API-README
// nc 127.0.0.1 4028 <<< '{"command":"stats+pools+summary"}' | sed "s/}{/,/" | jq .
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
// replace "}{" to "," in response to "stats" command
func jsonWorkaround(buf bytes.Buffer) []byte {
	real := buf.Bytes()[:buf.Len()-1]
	s := strings.Replace(string(real), "}{", ",", -1)
	return []byte(s)
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
	//summary := reply.Summary[0].Summary[0]
	stats := reply.Stats[0].Stats[0]
	pools := reply.Pools[0].Pools
	if err != nil || statusS.Status != "S" || statusP.Status != "S" {
		log.Println(cgMinerName+" error: ", err, statusS.Status, statusP.Status)
		return nil
	}

	tags["algorithm"] = m.getAlgorithm(i)
	tags["source"] = MINER.String()
	tags["version"] = stats.CGMiner
	fields := map[string]interface{}{
		"uptime":   stats.Elapsed,                      // in seconds
		"hashrate": uint64(stats.GHS5S * 1000000000.0), // was in GH/s
		"chain":    stats.MinerCount,
		"fan":      stats.FanNum,
	}
	for _, pool := range pools {
		if pool.StratumActive {
			tags["pool"] = pool.URL
			fields["shares_total"] = pool.Accepted + pool.Discarded
			fields["shares_good"] = pool.Accepted
			fields["shares_bad"] = pool.Rejected
			acc.AddFields(cgMinerName, fields, tags)
		}
	}

	delete(tags, "pool")
	delete(tags, "version")
	delete(tags, "algorithm")
	var responseMap map[string]*json.RawMessage
	json.Unmarshal(response, &responseMap)
	var statsMap []map[string]*json.RawMessage
	json.Unmarshal(*responseMap["stats"], &statsMap)
	var statsMap2 []map[string]*json.RawMessage
	json.Unmarshal(*statsMap[0]["STATS"], &statsMap2)
	statMap := statsMap2[0]
	for i := 1; i <= stats.MinerCount; i++ {
		tags["source"] = CHAIN.String()
		tags["unit"] = fmt.Sprintf("%d", i)
		var hash float64
		json.Unmarshal(*statMap[fmt.Sprintf("chain_rate%d", i)], &hash)
		// Normally, the PCB temp should be 40℃-85℃. The chips temp should be 85℃-115℃.
		var temp, tempPcb int
		json.Unmarshal(*statMap[fmt.Sprintf("temp%d", i)], &tempPcb)
		json.Unmarshal(*statMap[fmt.Sprintf("temp2_%d", i)], &temp)
		var acs string
		json.Unmarshal(*statMap[fmt.Sprintf("chain_acs%d", i)], &acs)
		fields := map[string]interface{}{
			"hashrate":        uint64(hash),
			"temperature_pcb": tempPcb,
			"temperature":     temp,
			"chain_state":     strings.Count(acs, "x"),
		}
		acc.AddFields(cgMinerName, fields, tags)
	}
	for i := 1; i <= stats.FanNum; i++ {
		tags["source"] = FAN.String()
		tags["unit"] = fmt.Sprintf("%d", i)
		var t int
		json.Unmarshal(*statMap[fmt.Sprintf("fan%d", i)], &t)
		fields := map[string]interface{}{
			"fan": t,
		}
		acc.AddFields(cgMinerName, fields, tags)
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
