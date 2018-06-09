package crypto

import (
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type luxorUserResponse struct {
	Address                 string  `json:"address"`
	Balance                 float64 `json:"balance"`
	EstimatedBalance        float64 `json:"estimated_balance,string"`
	Hashrate1H              float64 `json:"hashrate_1h"`
	Hashrate24H             float64 `json:"hashrate_24h"`
	ValidSharesFiveMin      float64 `json:"valid_shares_five_min"`
	ValidSharesFifteenMin   float64 `json:"valid_shares_fifteen_min"`
	ValidSharesOneHour      float64 `json:"valid_shares_one_hour"`
	ValidSharesSixHour      float64 `json:"valid_shares_six_hour"`
	ValidSharesOneDay       float64 `json:"valid_shares_one_day"`
	InvalidSharesFiveMin    int     `json:"invalid_shares_five_min"`
	InvalidSharesFifteenMin int     `json:"invalid_shares_fifteen_min"`
	InvalidSharesOneHour    int     `json:"invalid_shares_one_hour"`
	InvalidSharesSixHour    int     `json:"invalid_shares_six_hour"`
	InvalidSharesOneDay     int     `json:"invalid_shares_one_day"`
	StaleSharesFiveMin      int     `json:"stale_shares_five_min"`
	StaleSharesFifteenMin   int     `json:"stale_shares_fifteen_min"`
	StaleSharesOneHour      int     `json:"stale_shares_one_hour"`
	StaleSharesSixHour      int     `json:"stale_shares_six_hour"`
	StaleSharesOneDay       int     `json:"stale_shares_one_day"`
	PayoutsFiveMin          float64 `json:"payouts_five_min"`
	PayoutsFifteenMin       float64 `json:"payouts_fifteen_min"`
	PayoutsOneHour          float64 `json:"payouts_one_hour"`
	PayoutsSixHour          float64 `json:"payouts_six_hour"`
	PayoutsOneDay           float64 `json:"payouts_one_day"`
	TotalPayouts            float64 `json:"total_payouts"`
	BlocksFound             int     `json:"blocks_found"`
	LastShareTime           int     `json:"last_share_time"`
	Miners                  []struct {
		Name               string  `json:"name"`
		Affinity           string  `json:"affinity"`
		MinerType          string  `json:"miner_type"`
		LastShareTime      int     `json:"last_share_time"`
		TotalShares        float64 `json:"total_shares"`
		HashrateFiveMin    float64 `json:"hashrate_five_min"`
		HashrateFifteenMin float64 `json:"hashrate_fifteen_min"`
		HashrateOneHour    float64 `json:"hashrate_one_hour"`
		HashrateSixHour    float64 `json:"hashrate_six_hour"`
		HashrateOneDay     float64 `json:"hashrate_one_day"`
		StaleFiveMin       int     `json:"stale_five_min"`
		StaleFifteenMin    int     `json:"stale_fifteen_min"`
		StaleOneHour       int     `json:"stale_one_hour"`
		StaleSixHour       int     `json:"stale_six_hour"`
		StaleOneDay        int     `json:"stale_one_day"`
	} `json:"miners"`
	Payouts  []interface{} `json:"payouts"`
	Hashrate []interface{} `json:"hashrate"`
	// Payouts                 []struct {
	// 	Username string    `json:"username"`
	// 	Amount   string    `json:"amount"`
	// 	TxID     string    `json:"tx_id"`
	// 	Time     time.Time `json:"time"`
	// } `json:"payouts"`
	// Hashrate []struct {
	// 	Time     int     `json:"time"`
	// 	Hashrate float64 `json:"hashrate"`
	// } `json:"hashrate"`
}

type luxorPricesResponse []struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Symbol           string      `json:"symbol"`
	Rank             string      `json:"rank"`
	PriceUsd         float64     `json:"price_usd,string"`
	PriceBtc         float64     `json:"price_btc,string"`
	Two4HVolumeUsd   float64     `json:"24h_volume_usd,string"`
	MarketCapUsd     float64     `json:"market_cap_usd,string"`
	AvailableSupply  float64     `json:"available_supply,string"`
	TotalSupply      float64     `json:"total_supply,string"`
	MaxSupply        interface{} `json:"max_supply,string"`
	PercentChange1H  float64     `json:"percent_change_1h,string"`
	PercentChange24H float64     `json:"percent_change_24h,string"`
	PercentChange7D  float64     `json:"percent_change_7d,string"`
	LastUpdated      uint64      `json:"last_updated,string"`
}

const (
	luxorName      = "luxor"
	luxorAPI       = "https://mining.luxor.tech/api/"
	luxorUserPath  = "/user/"
	luxorPricePath = "/price"
)

var luxorCoinUnit = map[string]float64{
	"SC":  1000000000000000000000000.0,
	"DCR": 1000000000000000000000000.0,
	"LBC": 1000000000000000000000000.0,
}

// Luxor api docs?
type Luxor struct {
	Coin     []string `toml:"coin"`
	Addr     []string `toml:"addr"`
	Earnings []string `toml:"earnings"`
}

var luxorSampleConf = `
  interval = "1m"
  coin     = [ <coin type: SC, DCR, LBC> ]
  addr     = [ <coin address> ]
  # earnings = [ <usd,btc> ]
`

// Description of Luxor
func (*Luxor) Description() string {
	return "Read Luxor's pool parameters"
}

// SampleConfig of Luxor
func (*Luxor) SampleConfig() string {
	return luxorSampleConf
}

func (n *Luxor) luxorAPIURL(path string, i int) string {
	return luxorAPI + n.Coin[i] + path
}

func (n *Luxor) getAccount(acc telegraf.Accumulator, i int, tags map[string]string) error {
	var reply luxorUserResponse
	if err := getResponseSimple(n.luxorAPIURL(luxorUserPath, i)+n.Addr[i], &reply); err != nil {
		return err
	}

	tags["source"] = ACCOUNT.String()
	tags["coin"] = n.Coin[i]
	daily := reply.PayoutsOneDay / luxorCoinUnit[n.Coin[i]]
	fields := map[string]interface{}{
		"balance":       reply.Balance / luxorCoinUnit[n.Coin[i]],
		"hashrate":      reply.Hashrate1H,
		"daily_coin":    daily,
		"payouts_total": reply.TotalPayouts / luxorCoinUnit[n.Coin[i]],
	}
	// if len(n.Earnings) > 0 {
	// 	tags["base_currency"] = n.Earnings[i]
	// 	var prices luxorPricesResponse
	// 	if getResponse(n.luxorAPIURL(luxorPricePath, i), &prices, luxorName) {
	// 		switch n.Earnings[i] {
	// 		case "usd":
	// 			fields["daily"] = prices[0].PriceUsd * daily
	// 		case "btc":
	// 			fields["daily"] = prices[0].PriceBtc * daily
	// 		}
	// 	}
	// }
	acc.AddFields(luxorName, fields, tags)

	tags["source"] = MINER.String()
	for _, worker := range reply.Miners {
		tags["name"] = worker.Name
		fields := map[string]interface{}{
			"hashrate":     worker.HashrateFiveMin,
			"shares_total": worker.TotalShares,
		}
		acc.AddFields(luxorName, fields, tags)
	}
	return nil
}

// Gather of Luxor
func (n *Luxor) Gather(acc telegraf.Accumulator) error {
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
	inputs.Add(luxorName, func() telegraf.Input { return &Luxor{} })
}
