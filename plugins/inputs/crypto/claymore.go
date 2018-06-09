package crypto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type claymoreResponse struct {
	//Version string  `json:"jsonrpc"`
	ID     uint64    `json:"id"`
	Error  string    `json:"error,omitempty"`
	Result [9]string `json:"result,omitempty"`
}

const (
	claymoreName    = "claymore"
	claymoreRequest = "{\"id\":0,\"jsonrpc\":\"2.0\",\"method\":\"miner_getstat1\"}"
)

var claymoreAlgoMap = map[string]string{
	"eth": ethash.String(),
	"ns":  neoscrypt.String(),
	"xmr": cryptonightv7.String(),
	"zec": equihash.String(),
}

var claymoreAlgoUnit = map[string]string{
	"eth": "kH", // hashrate is in KH/s API docs are wrongly said MH/s
	"ns":  "H",
	"xmr": "H",
	"zec": "Sol",
}

var claymoreSampleConf = `
  interval = "1m"
  ## Miner servers addresses and names
  servers = ["localhost:3333"]
  names   = ["Rig1"]
`

// Claymore API docs: https://github.com/abuisine/docker-claymore/blob/master/API.txt
type Claymore struct {
	serverBase
}

// Description of Claymore
func (*Claymore) Description() string {
	return "Read Claymore's mining status"
}

// SampleConfig of Claymore
func (*Claymore) SampleConfig() string {
	return claymoreSampleConf
}

func toInt(val string) int64 {
	res, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return res
}

func (m *Claymore) serverGather(acc telegraf.Accumulator, i int, tags map[string]string) error {
	var buf bytes.Buffer
	if err := jsonReader(m.getAddress(i), claymoreRequest, &buf); err != nil {
		log.Println(claymoreName+" error: ", err)
		return nil
	}

	var reply claymoreResponse
	err := json.Unmarshal(buf.Bytes(), &reply)
	if err != nil || len(reply.Error) != 0 {
		log.Println(claymoreName+" error: ", err, reply.Error)
		return nil
	}
	results := reply.Result

	tags["source"] = MINER.String()
	pools := strings.Split(results[7], ";")
	invalids := strings.Split(results[8], ";")
	hasDcr := len(pools) > 1
	if len(pools) > 0 {
		tags["pool"] = pools[0]
	}
	version := strings.TrimSpace(strings.Split(results[0], "-")[0])
	algo := strings.TrimSpace(strings.ToLower(strings.Split(results[0], "-")[1]))
	tags["version"] = version
	tags["algorithm"] = claymoreAlgoMap[algo]

	mul := unitMultilier(claymoreAlgoUnit[algo])

	shares := strings.Split(results[2], ";")
	invalid := toInt(invalids[0])
	total := toInt(shares[1])
	bad := toInt(shares[2])
	hashrates := strings.Split(results[3], ";")
	tempFans := strings.Split(results[6], ";")
	rate := int64(100)
	if total != 0 {
		rate = 100 * (total - bad - invalid) / total
	}
	fields := map[string]interface{}{
		"uptime":           toInt(results[1]) * 60, // was in minutes
		"hashrate":         toInt(shares[0]) * mul,
		"shares_total":     total,
		"shares_rejected":  bad,
		"shares_discarded": invalid,
		"shares_accepted":  total - bad - invalid,
		"shares_rate":      rate,
		"pool_switch":      toInt(invalids[1]),
		"gpu":              len(hashrates),
	}
	acc.AddFields(claymoreName, fields, tags)

	delete(tags, "version")
	delete(tags, "pool")
	for i, hash := range hashrates {
		tags["source"] = GPU.String()
		tags["unit"] = fmt.Sprintf("%d", i+1)
		fields := map[string]interface{}{
			"hashrate":    toInt(hash) * mul,
			"temperature": toInt(tempFans[2*i]),
			"fan":         toInt(tempFans[2*i+1]),
		}
		acc.AddFields(claymoreName, fields, tags)
	}
	if hasDcr {
		delete(tags, "unit")
		tags["version"] = version
		tags["algorithm"] = "dcr"
		tags["pool"] = pools[1]
		tags["source"] = MINER.String()

		shares = strings.Split(results[4], ";")
		invalid = toInt(invalids[2])
		total = toInt(shares[1])
		bad = toInt(shares[2])
		hashrates = strings.Split(results[5], ";")
		rate := int64(100)
		if total != 0 {
			rate = 100 * (total - bad - invalid) / total
		}
		fields := map[string]interface{}{
			"uptime":           toInt(results[1]) * 60, // was in minutes
			"hashrate":         toInt(shares[0]),       // * mul ???
			"shares_total":     total,
			"shares_rejected":  bad,
			"shares_accepted":  total - bad - invalid,
			"shares_discarded": invalid,
			"shares_rate":      rate,
			"pool_switch":      toInt(invalids[3]),
			"gpu":              len(hashrates),
		}
		acc.AddFields(claymoreName, fields, tags)

		delete(tags, "version")
		delete(tags, "pool")
		for i, hash := range hashrates {
			tags["source"] = GPU.String()
			tags["unit"] = fmt.Sprintf("%d", i+1)
			fields := map[string]interface{}{
				"hashrate":    toInt(hash), // * mul ???
				"temperature": toInt(tempFans[2*i]),
				"fan":         toInt(tempFans[2*i+1]),
			}
			acc.AddFields(claymoreName, fields, tags)
		}
	}
	return nil
}

// Gather for Claymore
func (m *Claymore) Gather(acc telegraf.Accumulator) error {
	return m.minerGather(acc, m)
}

func init() {
	inputs.Add(claymoreName, func() telegraf.Input { return &Claymore{} })
}
