package crypto

import (
	"log"
	"strconv"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type nanopoolUser struct {
	Account            string  `json:"account"`
	UnconfirmedBalance float64 `json:"unconfirmed_balance,string"`
	Balance            float64 `json:"balance,string"`
	Hashrate           float64 `json:"hashrate,string"`
	AvgHashrate        struct {
		H1  float64 `json:"h1,string"`
		H3  float64 `json:"h3,string"`
		H6  float64 `json:"h6,string"`
		H12 float64 `json:"h12,string"`
		H24 float64 `json:"h24,string"`
	} `json:"avgHashrate"`
	Workers []struct {
		ID        string  `json:"id"`
		UID       int     `json:"uid"`
		Hashrate  float64 `json:"hashrate,string"`
		Lastshare int     `json:"lastshare"`
		Rating    int     `json:"rating"`
		H1        float64 `json:"h1,string"`
		H3        float64 `json:"h3,string"`
		H6        float64 `json:"h6,string"`
		H12       float64 `json:"h12,string"`
		H24       float64 `json:"h24,string"`
	} `json:"workers"`
}

type nanopoolUserRespponse struct {
	Status bool         `json:"status"`
	Data   nanopoolUser `json:"data"`
	Error  string       `json:"error"`
}

type nanopoolReported struct {
	Worker   string  `json:"worker"`
	Hashrate float64 `json:"hashrate"`
}

type nanopoolReportedRespponse struct {
	Status bool               `json:"status"`
	Data   []nanopoolReported `json:"data"`
	Error  string             `json:"error"`
}

type nanopoolEarning struct {
	Coins    float64 `json:"coins"`
	Dollars  float64 `json:"dollars"`
	Yuan     float64 `json:"yuan"`
	Euros    float64 `json:"euros"`
	Rubles   float64 `json:"rubles"`
	Bitcoins float64 `json:"bitcoins"`
}

type nanopoolEarnings struct {
	Minute nanopoolEarning `json:"minute"`
	Hour   nanopoolEarning `json:"hour"`
	Day    nanopoolEarning `json:"day"`
	Week   nanopoolEarning `json:"week"`
	Month  nanopoolEarning `json:"month"`
}

type nanopoolEarningsRespponse struct {
	Status bool             `json:"status"`
	Data   nanopoolEarnings `json:"data"`
	Error  string           `json:"error"`
}

const (
	nanopoolName         = "nanopool"
	nanopoolAPI          = "https://api.nanopool.org/v1/"
	nanopoolUserPath     = "/user/"
	nanopoolReportedPath = "/reportedhashrates/"
	nanopoolEarningsPath = "/approximated_earnings/"
)

var nanopoolAlgoUnit = map[string]string{
	"eth":  "MH",
	"etc":  "MH",
	"etn":  "H",
	"pasc": "MH",
	"sia":  "MH",
	"xmr":  "H",
	"zec":  "Sol",
}

var nanopoolEarnings2Fiat = map[string]string{
	"eth":  "MH",
	"etc":  "MH",
	"etn":  "H",
	"pasc": "MH",
	"sia":  "MH",
	"xmr":  "H",
	"zec":  "Sol",
}

// Nanopool api docs: https://eth.nanopool.org/api
type Nanopool struct {
	Coin     []string `toml:"coin"`
	Addr     []string `toml:"addr"`
	Earnings []string `toml:"earnings"`
	Worker   []bool   `toml:"worker"`
	// Profitability []bool   `toml:"profitability"`

	// algoID map[string]int
	// algoPaying map[string]float64
}

var nanopoolSampleConf = `
  interval = "1m"
  coin = [ <coin type: eth, etc, sia, zec, xmr, asc, etn> ]
  addr = [ <coin addresses> ]
  # earnings = [ <coins,dollars,yuan,euros,rubles,bitcoins> ]
  # worker = [ false ]
`

// Description of Nanopool
func (*Nanopool) Description() string {
	return "Read Nanopool's pool parameters"
}

// SampleConfig of Nanopool
func (*Nanopool) SampleConfig() string {
	return nanopoolSampleConf
}

func (n *Nanopool) getDaily(i int, day nanopoolEarning) (currency string, daily float64) {
	switch n.Earnings[i] {
	case "coins":
		return n.Coin[i], day.Coins
	case "dollars":
		return "usd", day.Dollars
	case "yuan":
		return "cny", day.Yuan
	case "euros":
		return "eur", day.Euros
	case "rubles":
		return "rub", day.Rubles
	case "bitcoins":
		return "btc", day.Bitcoins
	}
	return n.Earnings[i], -1.0
}

func (n *Nanopool) getAccount(acc telegraf.Accumulator, i int, tags map[string]string) error {
	var user nanopoolUserRespponse
	url := nanopoolAPI + n.Coin[i] + nanopoolUserPath + n.Addr[i]
	if !getResponse(url, &user, nanopoolName) {
		return nil
	}
	if !user.Status {
		log.Println(nanopoolName+" error: ", user.Error, url)
		return nil
	}

	tags["source"] = ACCOUNT.String()
	tags["coin"] = n.Coin[i]
	unit := nanopoolAlgoUnit[n.Coin[i]]
	tags["unit"] = unit
	mul := float64(unitMultilier(unit))
	fields := map[string]interface{}{
		"balance":  user.Data.Balance,
		"hashrate": user.Data.Hashrate * mul,
	}
	var day nanopoolEarning
	if user.Data.Hashrate > 0 {
		var earnings nanopoolEarningsRespponse
		url = nanopoolAPI + n.Coin[i] + nanopoolEarningsPath + strconv.FormatFloat(user.Data.Hashrate, 'f', -1, 64)
		if getResponse(url, &earnings, nanopoolName) {
			if !earnings.Status {
				log.Println(nanopoolName+" error: ", earnings.Error, url)
			} else {
				day = earnings.Data.Day
			}
		}
	}
	tags["base_currency"], fields["daily"] = n.getDaily(i, day)
	acc.AddFields(nanopoolName, fields, tags)

	delete(tags, "base_currency")
	tags["source"] = MINER.String()
	if n.Worker[i] {
		var reported nanopoolReportedRespponse
		url = nanopoolAPI + n.Coin[i] + nanopoolReportedPath + n.Addr[i]
		if !getResponse(url, &reported, nanopoolName) {
			return nil
		}
		if !reported.Status {
			log.Println(nanopoolName+" error: ", reported.Error, url)
			return nil
		}

		reportedHash := map[string]float64{}
		for _, worker := range reported.Data {
			reportedHash[worker.Worker] = worker.Hashrate
		}
		for _, worker := range user.Data.Workers {
			tags["name"] = worker.ID
			fields := map[string]interface{}{
				"hashrate":          worker.Hashrate * mul,
				"hashrate_reported": reportedHash[worker.ID] * mul,
			}
			acc.AddFields(nanopoolName, fields, tags)
		}
	}
	return nil
}

// Gather of Nanopool
func (n *Nanopool) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	wg.Add(len(n.Coin))
	for i := 0; i < len(n.Coin); i++ {
		tags := map[string]string{
			"coin": n.Coin[i],
		}
		go func(i int, tags map[string]string) {
			defer wg.Done()
			if len(n.Addr[i]) > 0 {
				tags["source"] = ACCOUNT.String()
				acc.AddError(n.getAccount(acc, i, tags))
			}
		}(i, tags)
	}
	wg.Wait()
	return nil
}

func init() {
	inputs.Add(nanopoolName, func() telegraf.Input { return &Nanopool{} })
}
