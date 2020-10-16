// Package circonus contains the output plugin used to write metric data to a
// Circonus broker.
package circonus

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	apiclient "github.com/circonus-labs/go-apiclient"
	apiconfig "github.com/circonus-labs/go-apiclient/config"
	"github.com/google/uuid"
)

// apiClient values can be used to communicate with the Circonus API.
type apiClient interface {
	Get(reqPath string) ([]byte, error)
	FetchBroker(cid apiclient.CIDType) (*apiclient.Broker, error)
	FetchBrokers() (*[]apiclient.Broker, error)
	SearchCheckBundles(searchCriteria *apiclient.SearchQueryType,
		filterCriteria *apiclient.SearchFilterType) (*[]apiclient.CheckBundle, error)
	CreateCheckBundle(cfg *apiclient.CheckBundle) (*apiclient.CheckBundle, error)
}

// getAPIClient creates a Circonus API client.
func (c *Circonus) getAPIClient() error {
	apiCfg := &apiclient.Config{
		TokenKey: c.APIToken,
		TokenApp: c.APIApp,
		URL:      c.APIURL,
	}

	if c.APITLSCA != "" {
		cert, err := ioutil.ReadFile(c.APITLSCA)
		if err != nil {
			return fmt.Errorf("unable to configure Circonus API client: %w", err)
		}

		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(cert) {
			return fmt.Errorf("unable to add Circonus API CA certificate to certtificate pool")
		}

		apiCfg.TLSConfig = &tls.Config{RootCAs: cp}
	}

	if c.APIInsecureSkipVerify {
		if apiCfg.TLSConfig == nil {
			apiCfg.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		} else {
			apiCfg.TLSConfig.InsecureSkipVerify = true
		}
	}

	cli, err := apiclient.New(apiCfg)
	if err != nil {
		return fmt.Errorf("unable to create Circonus API client: %w", err)
	}

	c.api = cli

	return nil
}

// getCheckBroker returns the broker CID for a Circonus check.
func (c *Circonus) getCheckBroker(name string) (string, error) {
	if c.api == nil {
		return "", fmt.Errorf("unable to get check broker: Circonus API not initialized")
	}

	search := apiclient.SearchQueryType("(active:1)(type:httptrap)(name:" + name + ")")
	checks, err := c.api.SearchCheckBundles(&search, nil)
	if err != nil {
		return "", fmt.Errorf("unable to search check bundles: %w", err)
	}

	if checks == nil {
		return "", nil
	}

	if len(*checks) == 0 {
		return "", nil
	}

	if len((*checks)[0].Brokers) == 0 {
		return "", nil
	}

	return (*checks)[0].Brokers[0], nil
}

// getRandomBroker returns the CID and CN of a random eligible broker.
func (c *Circonus) getRandomBroker() (string, string, error) {
	if c.api == nil {
		return "", "", fmt.Errorf("unable to get random broker CID: Circonus API not initialized")
	}

	brokers, err := c.api.FetchBrokers()
	if err != nil {
		return "", "", fmt.Errorf("unable to fetch brokers: %w", err)
	}

	if brokers == nil {
		return "", "", nil
	}

	if len(*brokers) == 0 {
		return "", "", nil
	}

	brokerCID := ""
	brokerCN := ""
	for _, broker := range *brokers {
		if len(broker.Details) == 0 {
			continue
		}

		if broker.Details[0].IP == nil || *broker.Details[0].IP == "" {
			continue
		}

		exclude := false
		for _, excid := range c.ExcludeBrokers {
			if broker.CID == excid {
				exclude = true
				break
			}
		}

		if exclude {
			continue
		}

		httpTrap := false

		for _, mod := range broker.Details[0].Modules {
			if mod == "httptrap" {
				httpTrap = true
				break
			}
		}

		if !httpTrap {
			continue
		}

		brokerCID = broker.CID
		brokerCN = broker.Details[0].CN
		break
	}

	return brokerCID, brokerCN, nil
}

// getBrokerInfo returns the CN of a broker given the broker CID.
func (c *Circonus) getBrokerInfo(cid string) (string, error) {
	if c.api == nil {
		return "", fmt.Errorf("unable to get broker CID: Circonus API not initialized")
	}

	broker, err := c.api.FetchBroker(apiclient.CIDType(&cid))
	if err != nil {
		return "", fmt.Errorf("unable to search brokers: %w", err)
	}

	if broker == nil {
		return "", nil
	}

	if len(broker.Details) == 0 {
		return "", fmt.Errorf("broker missing details: CN")
	}

	return broker.Details[0].CN, nil
}

// getCheckSubmissionURL returns the submission URL for a Circonus check.
func (c *Circonus) getCheckSubmissionURL(name, brokerCN string) (string, error) {
	if c.api == nil {
		return "", fmt.Errorf("unable to get check submission URL: Circonus API not initialized")
	}

	search := apiclient.SearchQueryType("(active:1)(type:httptrap)(name:" + name + ")")
	checks, err := c.api.SearchCheckBundles(&search, nil)
	if err != nil {
		return "", fmt.Errorf("unable to search check bundles: %w", err)
	}

	if checks == nil {
		return "", nil
	}

	if len(*checks) == 0 {
		return "", nil
	}

	url, ok := (*checks)[0].Config[apiconfig.Key("submission_url")]
	if !ok {
		return "", fmt.Errorf("check found but does not contain submission URL")
	}

	re, err := regexp.Compile(`^http.*\:\/\/\d*\.\d*\.\d*\.\d*\:\d*\/`)
	if err != nil {
		return "", fmt.Errorf("unable to compile url regexp: %s: %s",
			url, err.Error())
	}

	if re.MatchString(url) {
		url = url[:strings.Index(url, "//")+2] + brokerCN +
			url[strings.Index(url[strings.Index(url, ":")+1:], ":")+
				strings.Index(url, ":")+1:]
	}

	return url, nil
}

// createCheck creates a Circonus check to be used for telegraf metrics.
// It returns the submission URL for the new check.
func (c *Circonus) createCheck(brokerCID, brokerCN string) (string, error) {
	if c.api == nil {
		return "", fmt.Errorf("unable to create check: Circonus API not initialized")
	}

	secret, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("unable to generate secret: %w", err)
	}

	check := &apiclient.CheckBundle{
		Brokers: []string{brokerCID},
		Config: apiclient.CheckBundleConfig{
			apiconfig.Key("asynch_metrics"): "true",
			apiconfig.Key("secret"):         secret.String(),
		},
		DisplayName:   "telegraf-httptrap",
		Metrics:       []apiclient.CheckBundleMetric{},
		MetricFilters: [][]string{{"allow", ".", "default"}},
		Period:        60,
		Status:        "active",
		Tags:          []string{"source:telegraf"},
		Target:        "localhost",
		Timeout:       10,
		Type:          "httptrap",
	}

	chk, err := c.api.CreateCheckBundle(check)
	if err != nil {
		return "", fmt.Errorf("unable to create Circonus check: %w", err)
	}

	url, ok := chk.Config[apiconfig.Key("submission_url")]
	if !ok {
		return "", fmt.Errorf("check created but does not contain submission URL")
	}

	re, err := regexp.Compile(`^http.*\:\/\/\d*\.\d*\.\d*\.\d*\:\d*\/`)
	if err != nil {
		return "", fmt.Errorf("unable to compile url regexp: %s: %s",
			url, err.Error())
	}

	if re.MatchString(url) {
		url = url[:strings.Index(url, "//")+2] + brokerCN +
			url[strings.Index(url[strings.Index(url, ":")+1:], ":")+
				strings.Index(url, ":")+1:]
	}

	return url, nil
}

// getSubmissionURL returns the submission URL for the Circonus check used to
// automatically receive Telegraf data. If the check does not exist, it will
// be created.
func (c *Circonus) getSubmissionURL() (string, error) {
	if c.api == nil {
		return "", fmt.Errorf("unable to get submission URL: Circonus API not initialized")
	}

	var err error

	checkName := "telegraf-httptrap"

	brokerCID := c.Broker
	brokerCN := ""

	if brokerCID == "" || strings.ToLower(brokerCID) == "auto" {
		brokerCID, err = c.getCheckBroker(checkName)
		if err != nil {
			return "", err
		}

		if brokerCID == "" {
			brokerCID, brokerCN, err = c.getRandomBroker()
			if err != nil {
				return "", err
			}
		} else {
			brokerCN, err = c.getBrokerInfo(brokerCID)
			if err != nil {
				return "", err
			}
		}
	} else {
		brokerCN, err = c.getBrokerInfo(brokerCID)
		if err != nil {
			return "", err
		}
	}

	if brokerCID == "" || brokerCN == "" {
		return "", fmt.Errorf("unable to get broker")
	}

	url, err := c.getCheckSubmissionURL(checkName, brokerCN)
	if err != nil {
		return "", err
	}

	if url != "" {
		return url, nil
	}

	url, err = c.createCheck(brokerCID, brokerCN)
	if err != nil {
		return "", err
	}

	return url, nil
}
