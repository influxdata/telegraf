package mailchimp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

const (
	reports_endpoint          string = "/3.0/reports"
	reports_endpoint_campaign string = "/3.0/reports/%s"
)

var mailchimp_datacenter = regexp.MustCompile("[a-z]+[0-9]+$")

type ChimpAPI struct {
	Transport http.RoundTripper
	Debug     bool

	sync.Mutex

	url *url.URL
}

type ReportsParams struct {
	Count          string
	Offset         string
	SinceSendTime  string
	BeforeSendTime string
}

func (p *ReportsParams) String() string {
	v := url.Values{}
	if p.Count != "" {
		v.Set("count", p.Count)
	}
	if p.Offset != "" {
		v.Set("offset", p.Offset)
	}
	if p.BeforeSendTime != "" {
		v.Set("before_send_time", p.BeforeSendTime)
	}
	if p.SinceSendTime != "" {
		v.Set("since_send_time", p.SinceSendTime)
	}
	return v.Encode()
}

func NewChimpAPI(apiKey string) *ChimpAPI {
	u := &url.URL{}
	u.Scheme = "https"
	u.Host = fmt.Sprintf("%s.api.mailchimp.com", mailchimp_datacenter.FindString(apiKey))
	u.User = url.UserPassword("", apiKey)
	return &ChimpAPI{url: u}
}

type APIError struct {
	Status   int    `json:"status"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("ERROR %v: %v. See %v", e.Status, e.Title, e.Type)
}

func chimpErrorCheck(body []byte) error {
	var e APIError
	json.Unmarshal(body, &e)
	if e.Title != "" || e.Status != 0 {
		return e
	}
	return nil
}

func (a *ChimpAPI) GetReports(params ReportsParams) (ReportsResponse, error) {
	a.Lock()
	defer a.Unlock()
	a.url.Path = reports_endpoint

	var response ReportsResponse
	rawjson, err := runChimp(a, params)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(rawjson, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (a *ChimpAPI) GetReport(campaignID string) (Report, error) {
	a.Lock()
	defer a.Unlock()
	a.url.Path = fmt.Sprintf(reports_endpoint_campaign, campaignID)

	var response Report
	rawjson, err := runChimp(a, ReportsParams{})
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(rawjson, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func runChimp(api *ChimpAPI, params ReportsParams) ([]byte, error) {
	client := &http.Client{
		Transport: api.Transport,
		Timeout:   time.Duration(4 * time.Second),
	}

	var b bytes.Buffer
	req, err := http.NewRequest("GET", api.url.String(), &b)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.String()
	req.Header.Set("User-Agent", "Telegraf-MailChimp-Plugin")
	if api.Debug {
		log.Printf("D! Request URL: %s", req.URL.String())
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if api.Debug {
		log.Printf("D! Response Body:%s", string(body))
	}

	if err = chimpErrorCheck(body); err != nil {
		return nil, err
	}
	return body, nil
}

type ReportsResponse struct {
	Reports    []Report `json:"reports"`
	TotalItems int      `json:"total_items"`
}

type Report struct {
	ID            string `json:"id"`
	CampaignTitle string `json:"campaign_title"`
	Type          string `json:"type"`
	EmailsSent    int    `json:"emails_sent"`
	AbuseReports  int    `json:"abuse_reports"`
	Unsubscribed  int    `json:"unsubscribed"`
	SendTime      string `json:"send_time"`

	TimeSeries    []TimeSerie
	Bounces       Bounces       `json:"bounces"`
	Forwards      Forwards      `json:"forwards"`
	Opens         Opens         `json:"opens"`
	Clicks        Clicks        `json:"clicks"`
	FacebookLikes FacebookLikes `json:"facebook_likes"`
	IndustryStats IndustryStats `json:"industry_stats"`
	ListStats     ListStats     `json:"list_stats"`
}

type Bounces struct {
	HardBounces  int `json:"hard_bounces"`
	SoftBounces  int `json:"soft_bounces"`
	SyntaxErrors int `json:"syntax_errors"`
}

type Forwards struct {
	ForwardsCount int `json:"forwards_count"`
	ForwardsOpens int `json:"forwards_opens"`
}

type Opens struct {
	OpensTotal  int     `json:"opens_total"`
	UniqueOpens int     `json:"unique_opens"`
	OpenRate    float64 `json:"open_rate"`
	LastOpen    string  `json:"last_open"`
}

type Clicks struct {
	ClicksTotal            int     `json:"clicks_total"`
	UniqueClicks           int     `json:"unique_clicks"`
	UniqueSubscriberClicks int     `json:"unique_subscriber_clicks"`
	ClickRate              float64 `json:"click_rate"`
	LastClick              string  `json:"last_click"`
}

type FacebookLikes struct {
	RecipientLikes int `json:"recipient_likes"`
	UniqueLikes    int `json:"unique_likes"`
	FacebookLikes  int `json:"facebook_likes"`
}

type IndustryStats struct {
	Type       string  `json:"type"`
	OpenRate   float64 `json:"open_rate"`
	ClickRate  float64 `json:"click_rate"`
	BounceRate float64 `json:"bounce_rate"`
	UnopenRate float64 `json:"unopen_rate"`
	UnsubRate  float64 `json:"unsub_rate"`
	AbuseRate  float64 `json:"abuse_rate"`
}

type ListStats struct {
	SubRate   float64 `json:"sub_rate"`
	UnsubRate float64 `json:"unsub_rate"`
	OpenRate  float64 `json:"open_rate"`
	ClickRate float64 `json:"click_rate"`
}

type TimeSerie struct {
	TimeStamp       string `json:"timestamp"`
	EmailsSent      int    `json:"emails_sent"`
	UniqueOpens     int    `json:"unique_opens"`
	RecipientsClick int    `json:"recipients_click"`
}
