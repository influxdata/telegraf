package crypto

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type simplemultialgo struct {
	Paying float64 `json:"paying,string"`
	Port   int     `json:"port"`
	Name   string  `json:"name"`
	Algo   int     `json:"algo"`
}

type profitabilityRespponse struct {
	Result struct {
		Simplemultialgo []simplemultialgo `json:"simplemultialgo"`
	} `json:"result"`
	Method string `json:"method"`
}

type speedData struct {
	Accepted          float64 `json:"a,string"`
	RejectedTarget    float64 `json:"rt,string"`
	RejectedStale     float64 `json:"rs,string"`
	RejectedDuplicate float64 `json:"rd,string"`
	RejectedOther     float64 `json:"ro,string"`
}

type algoSpeed struct {
	speed   speedData
	balance float64
}

type current struct {
	Profitability float64   `json:"profitability,string"`
	Data          algoSpeed `json:"data"`
	Name          string    `json:"name"`
	Suffix        string    `json:"suffix"`
	Algo          int       `json:"algo"`
}

type pastData struct {
	timestamp int64
	algoSpeed
}

type past struct {
	Data []pastData `json:"data"`
	Algo int        `json:"algo"`
}

type providerRespponse struct {
	Result struct {
		Current  []current     `json:"current"`
		NhWallet bool          `json:"nh_wallet"`
		Past     []past        `json:"past"`
		Payments []interface{} `json:"payments"`
		Addr     string        `json:"addr"`
	} `json:"result"`
	Method string `json:"method"`
}

type worker struct {
	Rigname    string
	Speed      speedData
	Uptime     int64
	XNSub      int
	Difficulty float64
	Location   int
	Algorithm  int
}

type workerResult struct {
	Addr    string   `json:"addr"`
	Workers []worker `json:"workers"`
	Algo    int      `json:"algo"`
}

type workerRespponse struct {
	Result workerResult `json:"result"`
	Method string       `json:"method"`
}

// UnmarshalJSON for past has a special handling since it not a vaild JSON
func (s *algoSpeed) UnmarshalJSON(b []byte) error {
	var err error
	var a []*json.RawMessage
	if err = json.Unmarshal(b, &a); err != nil {
		return err
	}
	if err = json.Unmarshal(*a[0], &s.speed); err != nil {
		return err
	}
	var tmp string
	if err = json.Unmarshal(*a[1], &tmp); err != nil {
		return err
	}
	s.balance, err = strconv.ParseFloat(tmp, 64)
	return err
}

// UnmarshalJSON for past has a special handling since it not a vaild JSON
func (p *pastData) UnmarshalJSON(b []byte) error {
	var err error
	var a []*json.RawMessage
	if err = json.Unmarshal(b, &a); err != nil {
		return err
	}
	if err = json.Unmarshal(*a[0], &p.timestamp); err != nil {
		return err
	}
	p.timestamp *= 300 //ugly nicehash feature?
	if err = json.Unmarshal(*a[1], &p.speed); err != nil {
		return err
	}
	var tmp string
	if err = json.Unmarshal(*a[2], &tmp); err != nil {
		return err
	}
	p.balance, err = strconv.ParseFloat(tmp, 64)
	return err
}

// UnmarshalJSON for worker has a special handling since it not a vaild JSON
func (w *worker) UnmarshalJSON(b []byte) error {
	var err error
	var a []*json.RawMessage
	if err = json.Unmarshal(b, &a); err != nil {
		return err
	}
	if err = json.Unmarshal(*a[0], &w.Rigname); err != nil {
		return err
	}
	if err = json.Unmarshal(*a[1], &w.Speed); err != nil {
		return err
	}
	if err = json.Unmarshal(*a[2], &w.Uptime); err != nil {
		return err
	}
	w.Uptime *= 60
	if err = json.Unmarshal(*a[3], &w.XNSub); err != nil {
		return err
	}
	var tmp string
	if err = json.Unmarshal(*a[4], &tmp); err != nil {
		return err
	}
	if w.Difficulty, err = strconv.ParseFloat(tmp, 64); err != nil {
		return err
	}
	if err := json.Unmarshal(*a[5], &w.Location); err != nil {
		return err
	}
	return json.Unmarshal(*a[6], &w.Algorithm)
}

// algorithm constants the order and id is the same as for nicehash
var algoUnit = map[string]string{
	"scrypt":         "TH",
	"sha256":         "PH",
	"scryptnf":       "GH",
	"x11":            "TH",
	"x13":            "GH",
	"keccak":         "TH",
	"x15":            "GH",
	"nist5":          "GH",
	"neoscrypt":      "GH",
	"lyra2re":        "GH",
	"whirlpoolx":     "GH",
	"qubit":          "TH",
	"quark":          "TH",
	"axiom":          "kH",
	"lyra2rev2":      "TH",
	"scryptjanenf16": "MH",
	"blake256r8":     "TH",
	"blake256r14":    "TH",
	"blake256r8vnl":  "TH",
	"hodl":           "kH",
	"ethash":         "GH",
	"decred":         "TH",
	"cryptonight":    "MH",
	"lbry":           "TH",
	"equihash":       "MSol",
	"pascal":         "TH",
	"x11gost":        "GH",
	"sia":            "TH",
	"blake2s":        "TH",
	"skunk":          "GH",
	"cryptonightv7":  "MH",
	//"daggerhashimoto": "GH",
}

// algorithm id, name mapoping for nicehash
var algoID = map[int]string{
	0:  "scrypt",
	1:  "sha256",
	2:  "scryptnf",
	3:  "x11",
	4:  "x13",
	5:  "keccak",
	6:  "x15",
	7:  "nist5",
	8:  "neoscrypt",
	9:  "lyra2re",
	10: "whirlpoolx",
	11: "qubit",
	12: "quark",
	13: "axiom",
	14: "lyra2rev2",
	15: "scryptjanenf16",
	16: "blake256r8",
	17: "blake256r14",
	18: "blake256r8vnl",
	19: "hodl",
	20: "ethash",
	//20: "daggerhashimoto"
	21: "decred",
	22: "cryptonight",
	23: "lbry",
	24: "equihash",
	25: "pascal",
	26: "x11gost",
	27: "sia",
	28: "blake2s",
	29: "skunk",
	30: "cryptonightv7",
}

const (
	nicehashName        = "nicehash"
	nicehashAPI         = "https://api.nicehash.com/api?method="
	profitabilityMethod = "simplemultialgo.info"
	providerMethod      = "stats.provider.ex&addr="
	workerMethod        = "stats.provider.workers&addr="
)

// Nicehash api docs: https://www.Nicehash.com/doc-api
type Nicehash struct {
	Addr          []string `toml:"addr"`
	Profitability []bool   `toml:"profitability"`
	Worker        []bool   `toml:"worker"`
	accAlgoSpeed  map[int]string
}

var nicehashSampleConf = `
  interval = "1m"
  ## each profitability is in BTC/GH/day
  # addr          = [ <Nicehash BTC address> ]
  # profitability = [ false ]
  # worker        = [ false ]
`

// Description of Nicehash
func (*Nicehash) Description() string {
	return "Read Nicehash's pool parameters"
}

// SampleConfig of Nicehash
func (*Nicehash) SampleConfig() string {
	return nicehashSampleConf
}

func (n *Nicehash) getStandardAlgoName(niceName string) string {
	name := strings.ToLower(niceName)
	if name == "daggerhashimoto" {
		return ethash.String()
	}
	return name
}

func (n *Nicehash) getCurrentPaying(acc telegraf.Accumulator, tags map[string]string) error {
	var reply profitabilityRespponse
	if err := getResponseSimple(nicehashAPI+profitabilityMethod, &reply); err != nil {
		return err
	}

	fields := map[string]interface{}{}
	for _, algo := range reply.Result.Simplemultialgo {
		name := n.getStandardAlgoName(algo.Name)
		// n.algoID[name] = algo.Algo
		// n.algoPaying[name] = algo.Paying
		// it's always in BTC/GH/day
		fields[name] = algo.Paying
	}

	acc.AddFields(nicehashName, fields, tags)
	return nil
}

func (n *Nicehash) getAccount(acc telegraf.Accumulator, i int, tags map[string]string) error {
	var reply providerRespponse
	if err := getResponseSimple(nicehashAPI+providerMethod+n.Addr[i], &reply); err != nil {
		return err
	}

	tags["base_currency"] = "btc"
	fields := map[string]interface{}{}
	for _, currency := range reply.Result.Current {
		data := currency.Data.speed
		if data.Accepted != 0 {
			tags["algorithm"] = n.getStandardAlgoName(currency.Name)
			tags["unit"] = currency.Suffix
			n.accAlgoSpeed[currency.Algo] = currency.Suffix
			mul := float64(unitMultilier(currency.Suffix))
			fields["profitability"] = currency.Profitability
			fields["hashrate"] = data.Accepted * mul
			fields["hashrate_rejected"] = (data.RejectedTarget + data.RejectedStale + data.RejectedDuplicate + data.RejectedOther) * mul
			fields["daily"] = currency.Profitability * data.Accepted
			acc.AddFields(nicehashName, fields, tags)
		}
	}
	return nil
}

func (n *Nicehash) getWorker(acc telegraf.Accumulator, i int, tags map[string]string) error {
	var reply workerRespponse
	if err := getResponseSimple(nicehashAPI+workerMethod+n.Addr[i], &reply); err != nil {
		return nil
	}

	fields := map[string]interface{}{}
	for _, worker := range reply.Result.Workers {
		if worker.Speed.Accepted != 0 {
			tags["worker"] = worker.Rigname
			tags["xnsub"] = strconv.Itoa(worker.XNSub)
			tags["algorithm"] = algoID[worker.Algorithm]
			unit := n.accAlgoSpeed[worker.Algorithm]
			//tags["unit"] = unit
			mul := float64(unitMultilier(unit))
			switch worker.Location {
			case 0:
				tags["location"] = "EU"
			case 1:
				tags["location"] = "US"
			case 2:
				tags["location"] = "HK"
			case 3:
				tags["location"] = "JP"
			default:
				tags["location"] = "Unkown"
			}
			fields["uptime"] = worker.Uptime
			speed := worker.Speed
			fields["hashrate"] = speed.Accepted * mul
			fields["hashrate_rejected"] = (speed.RejectedTarget + speed.RejectedStale + speed.RejectedDuplicate + speed.RejectedOther) * mul
			acc.AddFields(nicehashName, fields, tags)
		}
	}
	return nil
}

// Gather of Nicehash
func (n *Nicehash) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	wg.Add(len(n.Addr))
	for i := 0; i < len(n.Addr); i++ {
		tags := map[string]string{}

		go func(i int, tags map[string]string) {
			defer wg.Done()
			if len(n.Addr[i]) > 0 {
				tags["source"] = ACCOUNT.String()
				acc.AddError(n.getAccount(acc, i, tags))
			}
			if n.Worker[i] {
				tags["source"] = MINER.String()
				delete(tags, "unit")
				acc.AddError(n.getWorker(acc, i, tags))
			}
			if n.Profitability[i] {
				tags["source"] = POOL.String()
				tags["name"] = "paying"
				acc.AddError(n.getCurrentPaying(acc, tags))
			}
		}(i, tags)
	}
	wg.Wait()
	return nil
}

// NewNicehash constructor
func NewNicehash() *Nicehash {
	return &Nicehash{accAlgoSpeed: map[int]string{}}
}

func init() {
	inputs.Add(nicehashName, func() telegraf.Input { return NewNicehash() })
}
