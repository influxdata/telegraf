package etherscan

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrBalanceAddressCountExhausted = errors.New("maximum address limit reached for batch request")
	ErrReturnCodeNotOK              = errors.New("return code was not OK")
	ErrNotValidEthAddress           = errors.New("invalid Ethereum address")
)

type ValidatingBalanceResponder interface {
	ValidateResponseStatus() error
	Result() []AccountBalance
}

type balance struct {
	Network string `toml:"network"`
	Address string `toml:"address"`
	Tag     string `toml:"tag"`
}

type MultiBalanceResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Results []AccountBalance `json:"result"`
}

func (mbr *MultiBalanceResponse) ValidateResponseStatus() error {
	status, err := strconv.ParseInt(mbr.Status, 10, 64)
	if err != nil {
		return fmt.Errorf("error processing response status [%w]", err)
	}

	if err := validateReturnCode(status); err != nil {
		if mbr.Message == "OK" {
			return nil
		}

		return fmt.Errorf("%s error: [%s]", err.Error(), mbr.Message)
	}

	return nil
}

func (mbr *MultiBalanceResponse) Result() []AccountBalance {
	return mbr.Results
}

type SingleBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Results string `json:"result"`
}

func (sbr *SingleBalanceResponse) ValidateResponseStatus() error {
	status, err := strconv.ParseInt(sbr.Status, 10, 64)
	if err != nil {
		return fmt.Errorf("error processing response status [%w]", err)
	}

	if err := validateReturnCode(status); err != nil {
		if sbr.Message == "OK" {
			return nil
		}

		return fmt.Errorf("%s error: [%s]", err.Error(), sbr.Message)
	}

	return nil
}

func (sbr *SingleBalanceResponse) Result() []AccountBalance {
	return []AccountBalance{
		{
			Account: "",
			Balance: sbr.Results,
		},
	}
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
	if ok := common.IsHexAddress(addr); !ok {
		return ErrNotValidEthAddress
	}

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

func (br *BalanceRequest) Send(client HTTPClient) (map[string]interface{}, error) {
	resp, err := client.Get(br.URL().String())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			br.log.Errorf("response body close error: [%s]", err.Error())
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	reqType := br.URL().Query().Get("action")
	switch reqType {
	case "balance":
		balanceResponse := &SingleBalanceResponse{}
		return br.assignBalances(body, balanceResponse)

	case "balancemulti":
		balanceResponse := &MultiBalanceResponse{}
		return br.assignBalances(body, balanceResponse)

	default:
		return nil, fmt.Errorf("unknown request [%s] action [%s]", br.URL().String(), reqType)
	}
}

func (br *BalanceRequest) assignBalances(
	body []byte,
	responder ValidatingBalanceResponder,
) (map[string]interface{}, error) {
	err := json.Unmarshal(body, &responder)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling: [%w]", err)
	}

	err = responder.ValidateResponseStatus()
	if err != nil {
		return nil, fmt.Errorf("error parsing status: [%w]", err)
	}

	balances := make(map[string]interface{})
	for _, account := range responder.Result() {
		if br.URL().Query().Get("action") == "balance" {
			account.Account = br.URL().Query().Get("address")
		}

		floatEth, err := parseEthSum(account.Balance)
		if err != nil {
			return nil, fmt.Errorf(
				"parsing balance [%s] error: [%w]",
				account.Balance,
				err,
			)
		}

		balances[account.Account] = floatEth
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

func validateReturnCode(code int64) error {
	if code != 0 {
		return ErrReturnCodeNotOK
	}

	return nil
}
