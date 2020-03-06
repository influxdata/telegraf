package salesforce

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig = `
  ## specify your credentials
  ##
  username = "your_username"
  password = "your_password"
  ##
  ## (optional) security token
  # security_token = "your_security_token"
  ##
  ## (optional) environment type (sandbox or production)
  ## default is: production
  ##
  # environment = "production"
  ##
  ## (optional) API version (default: "39.0")
  ##
  # version = "39.0"
`

type limit struct {
	Max       int
	Remaining int
}

type limits map[string]limit

type Salesforce struct {
	Username       string
	Password       string
	SecurityToken  string
	Environment    string
	SessionID      string
	ServerURL      *url.URL
	OrganizationID string
	Version        string

	client *http.Client
}

const defaultVersion = "39.0"
const defaultEnvironment = "production"

// returns a new Salesforce plugin instance
func NewSalesforce() *Salesforce {
	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(5 * time.Second),
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(10 * time.Second),
	}
	return &Salesforce{
		client:      client,
		Version:     defaultVersion,
		Environment: defaultEnvironment}
}

func (s *Salesforce) SampleConfig() string {
	return sampleConfig
}

func (s *Salesforce) Description() string {
	return "Read API usage and limits for a Salesforce organisation"
}

// Reads limits values from Salesforce API
func (s *Salesforce) Gather(acc telegraf.Accumulator) error {
	limits, err := s.fetchLimits()
	if err != nil {
		return err
	}

	tags := map[string]string{
		"organization_id": s.OrganizationID,
		"host":            s.ServerURL.Host,
	}

	fields := make(map[string]interface{})
	for k, v := range limits {
		key := internal.SnakeCase(k)
		fields[key+"_max"] = v.Max
		fields[key+"_remaining"] = v.Remaining
	}

	acc.AddFields("salesforce", fields, tags)
	return nil
}

// query the limits endpoint
func (s *Salesforce) queryLimits() (*http.Response, error) {
	endpoint := fmt.Sprintf("%s://%s/services/data/v%s/limits", s.ServerURL.Scheme, s.ServerURL.Host, s.Version)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "encoding/json")
	req.Header.Add("Authorization", "Bearer "+s.SessionID)
	return s.client.Do(req)
}

func (s *Salesforce) isAuthenticated() bool {
	return s.SessionID != ""
}

func (s *Salesforce) fetchLimits() (limits, error) {
	var l limits
	if !s.isAuthenticated() {
		if err := s.login(); err != nil {
			return l, err
		}
	}

	resp, err := s.queryLimits()
	if err != nil {
		return l, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		if err = s.login(); err != nil {
			return l, err
		}
		resp, err = s.queryLimits()
		if err != nil {
			return l, err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return l, fmt.Errorf("Salesforce responded with unexpected status code %d", resp.StatusCode)
	}

	l = limits{}
	err = json.NewDecoder(resp.Body).Decode(&l)
	return l, err
}

func (s *Salesforce) getLoginEndpoint() (string, error) {
	switch s.Environment {
	case "sandbox":
		return fmt.Sprintf("https://test.salesforce.com/services/Soap/c/%s/", s.Version), nil
	case "production":
		return fmt.Sprintf("https://login.salesforce.com/services/Soap/c/%s/", s.Version), nil
	default:
		return "", fmt.Errorf("unknown environment type: %s", s.Environment)
	}
}

// Authenticate with Salesfroce
func (s *Salesforce) login() error {
	if s.Username == "" || s.Password == "" {
		return errors.New("missing username or password")
	}

	body := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
		<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
		  xmlns:urn="urn:enterprise.soap.sforce.com">
		  <soapenv:Body>
			<urn:login>
			  <urn:username>%s</urn:username>
			  <urn:password>%s%s</urn:password>
			</urn:login>
		  </soapenv:Body>
		</soapenv:Envelope>`,
		s.Username, s.Password, s.SecurityToken)

	loginEndpoint, err := s.getLoginEndpoint()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, loginEndpoint, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "text/xml")
	req.Header.Add("SOAPAction", "login")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// ignore the err here; LimitReader returns io.EOF and we're not interested in read errors.
		body, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 200))
		return fmt.Errorf("%s returned HTTP status %s: %q", loginEndpoint, resp.Status, body)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	soapFault := struct {
		Code    string `xml:"Body>Fault>faultcode"`
		Message string `xml:"Body>Fault>faultstring"`
	}{}

	err = xml.Unmarshal(respBody, &soapFault)
	if err != nil {
		return err
	}

	if soapFault.Code != "" {
		return fmt.Errorf("login failed: %s", soapFault.Message)
	}

	loginResult := struct {
		ServerURL      string `xml:"Body>loginResponse>result>serverUrl"`
		SessionID      string `xml:"Body>loginResponse>result>sessionId"`
		OrganizationID string `xml:"Body>loginResponse>result>userInfo>organizationId"`
	}{}

	err = xml.Unmarshal(respBody, &loginResult)
	if err != nil {
		return err
	}

	s.SessionID = loginResult.SessionID
	s.OrganizationID = loginResult.OrganizationID
	s.ServerURL, err = url.Parse(loginResult.ServerURL)

	return err
}

func init() {
	inputs.Add("salesforce", func() telegraf.Input {
		return NewSalesforce()
	})
}
