package etherscan

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
)

var (
	ErrBalanceAddressCountExhausted = errors.New("maximum address limit reached for batch request")
)

type balance struct {
	Network string `toml:"network"`
	Address string `toml:"address"`
	Tag     string `toml:"tag"`
}

type MultiBalanceResponse struct {
	Status  int              `json:"status"`
	Message string           `json:"message"`
	Results []AccountBalance `json:"result"`
}

type SingleBalanceResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

type AccountBalance struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}

type BalanceRequest struct {
	log           telegraf.Logger
	lastTriggered time.Time
	request       *url.URL
	tags          map[string]string
}

func NewBalanceRequest(requestURL *url.URL, apiKey string, log telegraf.Logger) *BalanceRequest {
	nbr := BalanceRequest{
		log:           log,
		lastTriggered: time.Now(),
		request:       requestURL,
		tags:          make(map[string]string),
	}

	query := nbr.request.Query()
	query.Set("module", "account")
	query.Set("tag", "latest")
	query.Set("apikey", apiKey)
	nbr.request.RawQuery = query.Encode()

	return &nbr
}

func (br *BalanceRequest) AddAddress(addr string) error {
	query := br.URL().Query()
	action := query.Get("action")
	switch action {
	case "":
		query.Set("action", "balance")
		query.Set("address", addr)

	case "balance":
		query.Set("action", "balancemulti")
		fallthrough

	case "balancemulti":
		addresses := br.RequestAddressList()
		if len(addresses) >= 20 {
			return ErrBalanceAddressCountExhausted
		}

		addresses = append(addresses, addr)

		addressesCSV := strings.Join(addresses, ",")
		query.Set("address", addressesCSV)
	default:
		return fmt.Errorf("error building request unknown action type [%s]", action)
	}

	br.URL().RawQuery = query.Encode()

	br.log.Debugf("balance request created [%s]", br.URL().String())

	return nil
}

func (br *BalanceRequest) RequestAddressList() []string {
	var addresses []string
	addressesRaw := br.request.Query().Get("address")
	if addressesRaw != "" {
		addresses = strings.Split(addressesRaw, ",")
	}

	return addresses
}

func (br *BalanceRequest) AddTag(addr, tag string) {
	br.tags[addr] = tag
}

func (br *BalanceRequest) URL() *url.URL {
	return br.request
}

func (br *BalanceRequest) UpdateURL(url *url.URL) {
	br.request = url
}

func (br *BalanceRequest) LastRequest() time.Time {
	return br.lastTriggered
}

func (br *BalanceRequest) MarkTriggered() {
	br.lastTriggered = time.Now()
}

func (br *BalanceRequest) Tags() map[string]string {
	return br.tags
}

func (br *BalanceRequest) Send(client *http.Client) (map[string]interface{}, error) {
	resp, err := client.Get(br.URL().String())
	if err != nil {
		return nil, fmt.Errorf(
			"get request [%s] error: %w", br.URL().String(), err,
		)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			br.log.Errorf("response body close error: %s", err.Error())
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"request [%s] read body error: %w", br.URL().String(), err,
		)
	}

	balances := make(map[string]interface{})
	reqType := br.URL().Query().Get("action")
	switch reqType {
	case "balance":
		address := br.URL().Query().Get("address")
		singleBalance := &SingleBalanceResponse{}

		err = json.Unmarshal(body, &singleBalance)
		if err != nil {
			return nil, fmt.Errorf("request [%s] error unmarshaling: %w", br.URL().String(), err)
		}
		if singleBalance.Status != 0 {
			return nil, fmt.Errorf(
				"request [%s] error returned: [%s]",
				br.URL().String(),
				singleBalance.Message,
			)
		}

		br.log.Debugf("request [%s] response [%+v]", br.URL().String(), singleBalance)

		floatEth, err := parseEthSum(singleBalance.Result)
		if err != nil {
			br.log.Errorf("error in response [%s]", err.Error())
			return nil, fmt.Errorf(
				"request [%s] parsing balance [%s] error: [%w]",
				br.URL().String(),
				singleBalance.Result,
				err,
			)
		}
		balances[address] = floatEth

	case "balancemulti":
		multiBalance := &MultiBalanceResponse{}
		err = json.Unmarshal(body, &multiBalance)
		if err != nil {
			return nil, fmt.Errorf("request [%s] error unmarshaling: %w", br.URL().String(), err)
		}
		if multiBalance.Status != 0 {
			return nil, fmt.Errorf(
				"request [%s] error returned: [%s]",
				br.URL().String(),
				multiBalance.Message,
			)
		}

		br.log.Debugf("request [%s] response [%+v]", br.URL().String(), multiBalance)

		for _, account := range multiBalance.Results {
			floatEth, err := parseEthSum(account.Balance)
			if err != nil {
				return nil, fmt.Errorf(
					"request [%s] parsing balance [%s] error: [%w]",
					br.URL().String(),
					account.Balance,
					err,
				)
			}
			balances[account.Account] = floatEth
		}

	default:
		return nil, fmt.Errorf("unknown request [%s] action [%s]", br.URL().String(), reqType)
	}

	return balances, nil
}

func parseEthSum(balance string) (float64, error) {
	bal64, err := strconv.ParseFloat(balance, 64)
	if err != nil {
		return -1, err
	}

	bf := big.Float{}
	balFloat := bf.SetFloat64(bal64)
	ethValue := new(big.Float).Quo(balFloat, big.NewFloat(math.Pow10(18)))
	floatEth, _ := ethValue.Float64()

	return floatEth, nil
}
