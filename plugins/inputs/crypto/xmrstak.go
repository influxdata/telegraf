package crypto

import (
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type hashrate struct {
	Threads [][3]float64 `json:"threads"`
	Total   [3]float64   `json:"total"`
	Highest float64      `json:"highest"`
}

type errorLog struct {
	Count    int    `json:"count"`
	LastSeen int    `json:"last_seen"`
	Text     string `json:"text"`
}

type results struct {
	DiffCurrent int        `json:"diff_current"`
	SharesGood  int        `json:"shares_good"`
	SharesTotal int        `json:"shares_total"`
	AverageTime float64    `json:"avg_time"`
	HashesTotal int        `json:"hashes_total"`
	Best        []int      `json:"best"`
	ErrorLog    []errorLog `json:"error_log"`
}

type connection struct {
	Pool     string     `json:"pool"`
	Uptime   uint       `json:"uptime"`
	Ping     int        `json:"ping"`
	ErrorLog []errorLog `json:"error_log"`
}

type xmrStakResponse struct {
	Version    string     `json:"version"`
	Hashrate   hashrate   `json:"hashrate"`
	Results    results    `json:"results"`
	Connection connection `json:"connection"`
}

const xmrstakName = "xmr_stak"

var xmrstakSampleConf = `
  interval = "1m"
  ## Miner servers addresses and names
  servers = ["localhost:420"]
  names   = ["Rig1"]
  # number of threads per GPU
  threads = 2
`

// XMRStak miner
type XMRStak struct {
	serverBase
	Threads int `toml:"threads"`
}

// Description of XMRStak
func (*XMRStak) Description() string {
	return "Read XMR Stak's mining status from server(s)"
}

// SampleConfig of XMRStak
func (*XMRStak) SampleConfig() string {
	return xmrstakSampleConf
}

func (*XMRStak) getURL(address string) string {
	return "http://" + address + "/api.json"
}

func (m *XMRStak) serverGather(acc telegraf.Accumulator, i int, tags map[string]string) error {
	var reply xmrStakResponse
	if !getResponse(m.getURL(m.getAddress(i)), &reply, xmrstakName) {
		return nil
	}
	if reply.Connection.Pool == "not connected" || reply.Hashrate.Total[0] == 0 {
		log.Println(xmrstakName+" error: ", reply.Connection.Pool, reply.Hashrate.Total[0])
		return nil
	}

	version := strings.Split(reply.Version, "/")[1]

	tags["source"] = MINER.String()
	tags["algorithm"] = cryptonightv7.String()
	tags["version"] = version
	tags["pool"] = reply.Connection.Pool

	fields := map[string]interface{}{
		"uptime":          reply.Connection.Uptime, // in seconds
		"hashrate":        uint64(reply.Hashrate.Total[0]),
		"ping":            reply.Connection.Ping,
		"diff_current":    reply.Results.DiffCurrent,
		"shares_total":    reply.Results.SharesTotal,
		"shares_accepted": reply.Results.SharesGood,
		"shares_rejected": reply.Results.SharesTotal - reply.Results.SharesGood,
		"shares_rate":     100 * reply.Results.SharesGood / reply.Results.SharesTotal,
		"avg_time":        reply.Results.AverageTime,
		"hashes_total":    reply.Results.HashesTotal,
		"gpu":             len(reply.Hashrate.Threads) / m.Threads,
		"thread":          len(reply.Hashrate.Threads),
	}
	acc.AddFields(xmrstakName, fields, tags)

	delete(tags, "version")
	delete(tags, "pool")
	var gpu float64
	var x int
	fields = map[string]interface{}{}
	for i, hashStat := range reply.Hashrate.Threads {
		hash := hashStat[0]
		tags["source"] = THREAD.String()
		tags["unit"] = fmt.Sprintf("%d", i+1)
		fields["hashrate"] = uint64(hash)
		acc.AddFields(xmrstakName, fields, tags)

		gpu += hash
		if i%m.Threads == (m.Threads - 1) {
			tags["source"] = GPU.String()
			tags["unit"] = fmt.Sprintf("%d", i/m.Threads+1)
			fields["hashrate"] = uint64(gpu)
			acc.AddFields(xmrstakName, fields, tags)
			gpu = 0
		}
		x = i
	}
	if x%m.Threads != (m.Threads - 1) {
		tags["source"] = GPU.String()
		tags["unit"] = fmt.Sprintf("%d", x/m.Threads+1)
		fields["hashrate"] = uint64(gpu)
		acc.AddFields(xmrstakName, fields, tags)
	}
	for i, best := range reply.Results.Best {
		fields[fmt.Sprintf("best_%d", i)] = best
	}
	return nil
}

// Gather for XMRStak
func (m *XMRStak) Gather(acc telegraf.Accumulator) error {
	return m.minerGather(acc, m)
}

func init() {
	inputs.Add(xmrstakName, func() telegraf.Input { return &XMRStak{} })
}
