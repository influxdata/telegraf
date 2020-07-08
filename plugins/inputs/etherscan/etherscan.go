package etherscan

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Etherscan struct {
	log telegraf.Logger

	client          *http.Client
	Timeout         string `toml:"req_timeout_duration" default:"3s"`
	timeoutDuration time.Duration

	networkRequests         requestQueue
	RequestCount            int
	RequestLimit            int    `toml:"api_rate_limit_calls" default:"5"`
	RequestInterval         string `toml:"api_rate_limit_duration" default:"1s"`
	RequestIntervalDuration time.Duration
	Key                     string `toml:"api_key" required:"true"`

	Balances []balance `toml:"balance"`

	initializeOnce sync.Once
}

func NewEtherscanInputPlugin() *Etherscan {
	nep := Etherscan{
		client:          &http.Client{},
		log:             models.NewLogger("input", "etherscan", ""),
		networkRequests: make(requestQueue),
		Key:             "",
		RequestLimit:    0,
		RequestInterval: "",
		Balances:        make([]balance, 0),
		initializeOnce:  sync.Once{},
	}

	return &nep
}

const ExampleEtherscanConfig = `
[[inputs.etherscan]]
  api_key = "ETHERSCANAPIKEYHERE"
  api_rate_limit_calls = 5
  api_rate_limit_duration = "1s"
  req_timeout_duration = "3s"
[[inputs.etherscan.balance]]
  address = "0x2c02a0977E8BFee1628D0c4712B0841c0C2D36a7"
  tag = "testAddr"
  network = "Mainnet"
`

func (m *Etherscan) Description() string {
	return `Gather information from Ethereum networks via the Etherscan Block Explorer API`
}

func (m *Etherscan) SampleConfig() string {
	return ExampleEtherscanConfig
}

func (m *Etherscan) Setup() error {
	reqInvDur, err := time.ParseDuration(m.RequestInterval)
	if err != nil {
		return err
	}
	m.RequestIntervalDuration = reqInvDur

	reqTimeoutDur, err := time.ParseDuration(m.Timeout)
	if err != nil {
		return err
	}
	m.timeoutDuration = reqTimeoutDur

	client := &http.Client{
		Timeout: m.timeoutDuration,
	}
	m.client = client

	// Building request list for all eth balance inquiries
	var balanceReqExist bool
	for _, v := range m.Balances {
		balanceReqExist = false
		networkAPI, err := NetworkLookup(v.Network)
		if err != nil {
			return fmt.Errorf("invalid network for [%s] %w", v.Address, err)
		}

		if networkRequests, ok := m.networkRequests[networkAPI]; ok {
			for _, req := range networkRequests {
				if balReq, ok := req.(*BalanceRequest); ok {
					balanceReqExist = true
					err := balReq.AddAddress(v.Address)
					if v.Tag != "" {
						balReq.AddTag(v.Address, v.Tag)
					}
					if errors.Is(err, ErrBalanceAddressCountExhausted) {
						balanceReqExist = false
						break
					}
					if err != nil {
						return fmt.Errorf("error adding address [%s] %w", v.Address, err)
					}
				}
			}
		}

		if !balanceReqExist {
			apiURL, err := url.Parse(networkAPI.URL())
			if err != nil {
				return fmt.Errorf(
					"error parsing network api [%s] into url %w",
					networkAPI.URL(),
					err,
				)
			}

			nbr := NewBalanceRequest(apiURL, m.Key, m.log)

			err = nbr.AddAddress(v.Address)
			if err != nil {
				return fmt.Errorf("error adding address [%s] %w", v.Address, err)
			}

			if v.Tag != "" {
				nbr.AddTag(v.Address, v.Tag)
			}

			m.networkRequests[networkAPI] = append(m.networkRequests[networkAPI], nbr)
		}
	}

	return nil
}

func (m *Etherscan) Initialize() {
	if m.RequestCount > 0 {
		m.log.Warnf(
			"number of requests starting to queue [%d] requests left", m.RequestCount,
		)
	}

	for _, requestList := range m.networkRequests {
		m.RequestCount += len(requestList)
	}
}

func (m *Etherscan) Gather(acc telegraf.Accumulator) error {
	m.initializeOnce.Do(
		func() {
			if err := m.Setup(); err != nil {
				acc.AddError(err)
			}
		},
	)
	m.Initialize()

	for _, requestList := range m.networkRequests {
		sort.Sort(requestList)
	}

	var wg sync.WaitGroup
	for _, requestList := range m.networkRequests {
		for _, request := range requestList {
			<-time.After(m.RequestIntervalDuration)
			for i := 0; i < m.RequestLimit && i < len(requestList); i++ {
				wg.Add(1)
				go m.Fetch(request, acc, &wg)
			}
		}
	}

	wg.Wait()

	return nil
}

func (m *Etherscan) Fetch(request Request, acc telegraf.Accumulator, wg *sync.WaitGroup) {
	etherscanRequest := request.(Request)
	etherscanRequest.MarkTriggered()

	results, err := etherscanRequest.Send(m.client)
	if err != nil {
		acc.AddError(err)
	}
	m.log.Debugf("results [%+v]", results)

	tags := etherscanRequest.Tags()

	m.RequestCount--

	acc.AddFields("balances", results, tags)

	wg.Done()
}

func init() {
	inputs.Add("etherscan", func() telegraf.Input { return NewEtherscanInputPlugin() })
}
