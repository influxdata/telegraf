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

type EtherscanOption func(*Etherscan)

func WithHTTPClient(newClient HTTPClient) EtherscanOption {
	return func(plugin *Etherscan) {
		plugin.client = newClient
	}
}

func WithLogger(newLogger telegraf.Logger) EtherscanOption {
	return func(plugin *Etherscan) {
		plugin.log = newLogger
	}
}

type Etherscan struct {
	log telegraf.Logger

	client          HTTPClient
	Timeout         string `toml:"req_timeout_duration"`
	timeoutDuration time.Duration

	networkRequests requestQueue
	Key             string `toml:"api_key" required:"true"`
	requestCount    int

	RequestLimit            int    `toml:"api_rate_limit_calls"`
	RequestInterval         string `toml:"api_rate_limit_duration"`
	requestIntervalDuration time.Duration
	burstLimiter            chan time.Time

	Balances []balance `toml:"balance"`

	initializeOnce sync.Once
}

func NewEtherscanInputPlugin(opts ...EtherscanOption) *Etherscan {
	nep := Etherscan{
		client:          &http.Client{},
		log:             models.NewLogger("input", "etherscan", ""),
		networkRequests: make(requestQueue),
		Key:             "",
		Timeout:         _RequestTimeout,
		RequestLimit:    _RequestLimit,
		RequestInterval: _RequestInterval,
		Balances:        make([]balance, 0),
		initializeOnce:  sync.Once{},
	}

	for _, opt := range opts {
		opt(&nep)
	}

	return &nep
}

type _ telegraf.Input

func (m *Etherscan) Description() string {
	return _Description
}

func (m *Etherscan) SampleConfig() string {
	return _ExampleEtherscanConfig
}

func (m *Etherscan) Setup() error {
	reqInvDur, err := time.ParseDuration(m.RequestInterval)
	if err != nil {
		return err
	}
	m.requestIntervalDuration = reqInvDur

	reqTimeoutDur, err := time.ParseDuration(m.Timeout)
	if err != nil {
		return err
	}
	m.timeoutDuration = reqTimeoutDur

	client := &http.Client{
		Timeout: m.timeoutDuration,
	}
	m.client = client

	m.burstLimiter = make(chan time.Time, m.RequestLimit)
	go func() {
		for t := range time.Tick(m.requestIntervalDuration) {
			m.burstLimiter <- t
		}
	}()

	err = m.BuildAndQueueBalanceRequests()
	if err != nil {
		return err
	}

	return nil
}

func (m *Etherscan) PrepareInterval() error {
	if m.requestCount > 0 {
		m.log.Warn("flushing")

		m.networkRequests = make(requestQueue)

		err := m.BuildAndQueueBalanceRequests()
		if err != nil {
			return err
		}
	}

	for _, requestList := range m.networkRequests {
		m.requestCount += len(requestList)
	}

	return nil
}

func (m *Etherscan) Gather(acc telegraf.Accumulator) error {
	m.initializeOnce.Do(
		func() {
			if err := m.Setup(); err != nil {
				acc.AddError(err)
			}
		},
	)

	if err := m.PrepareInterval(); err != nil {
		acc.AddError(err)
	}

	for _, requestList := range m.networkRequests {
		sort.Sort(requestList)
	}

	for _, requestList := range m.networkRequests {
		for _, request := range requestList {
			<-m.burstLimiter
			go m.Fetch(request, acc)
		}
	}

	return nil
}

func (m *Etherscan) Fetch(request Request, acc telegraf.Accumulator) {
	etherscanRequest := request.(Request)

	etherscanRequest.MarkTriggered()

	results, err := etherscanRequest.Send(m.client)
	if err != nil {
		m.requestCount--
		acc.AddError(err)
		return
	}
	m.log.Debugf("results [%+v]", results)

	tags := etherscanRequest.Tags()

	m.requestCount--

	acc.AddFields("balances", results, tags)
}

func (m *Etherscan) BuildAndQueueBalanceRequests() error {
	var balanceReqExist bool
	for _, balanceAccount := range m.Balances {
		balanceReqExist = false
		networkAPI, err := NetworkLookup(balanceAccount.Network)
		if err != nil {
			return fmt.Errorf(
				"invalid network [%s] for [%s] %w",
				balanceAccount.Network,
				balanceAccount.Address,
				err,
			)
		}

		if networkRequests, ok := m.networkRequests[networkAPI]; ok {
			for idx, req := range networkRequests {
				if balReq, ok := req.(*BalanceRequest); ok {
					balanceReqExist = true
					err := balReq.AddAddress(balanceAccount.Address)
					if balanceAccount.Tag != "" {
						balReq.AddTag(balanceAccount.Address, balanceAccount.Tag)
					}
					if errors.Is(err, ErrBalanceAddressCountExhausted) &&
						idx+1 >= len(m.networkRequests[networkAPI]) {
						balanceReqExist = false
						break
					}
					if errors.Is(err, ErrBalanceAddressCountExhausted) {
						continue
					}
					if err != nil {
						return fmt.Errorf("error adding address [%s] %w", balanceAccount.Address, err)
					}
				}
			}
		}

		if !balanceReqExist {
			apiURL, err := url.Parse(networkAPI.URL())
			if err != nil {
				return err
			}

			nbr := NewBalanceRequest(apiURL, m.Key, m.log)

			err = nbr.AddAddress(balanceAccount.Address)
			if err != nil {
				return fmt.Errorf("error adding address [%s] %w", balanceAccount.Address, err)
			}

			if balanceAccount.Tag != "" {
				nbr.AddTag(balanceAccount.Address, balanceAccount.Tag)
			}

			m.networkRequests[networkAPI] = append(m.networkRequests[networkAPI], nbr)
		}
	}

	return nil
}

func init() {
	inputs.Add("etherscan", func() telegraf.Input { return NewEtherscanInputPlugin() })
}
