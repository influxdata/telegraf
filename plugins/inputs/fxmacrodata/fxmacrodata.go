//go:generate ../../../tools/readme_config_includer/generator
package fxmacrodata

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const defaultBaseURL = "https://api.fxmacrodata.com"

type FXMacroData struct {
	APIKey          config.Secret   `toml:"api_key"`
	Currencies      []string        `toml:"currencies"`
	Indicators      []string        `toml:"indicators"`
	FXPairs         []string        `toml:"fx_pairs"`
	BaseURL         string          `toml:"base_url"`
	ResponseTimeout config.Duration `toml:"response_timeout"`
	Log             telegraf.Logger `toml:"-"`
	tls.ClientConfig

	client *http.Client
}

type observation struct {
	Date                 string   `json:"date"`
	Value                *float64 `json:"val"`
	AnnouncementDatetime *int64   `json:"announcement_datetime"`
	Source               string   `json:"source"`
}

type seriesResponse struct {
	Data []observation `json:"data"`
}

func (*FXMacroData) SampleConfig() string {
	return sampleConfig
}

func (f *FXMacroData) Init() error {
	if f.BaseURL == "" {
		f.BaseURL = defaultBaseURL
	}
	if _, err := url.Parse(f.BaseURL); err != nil {
		return fmt.Errorf("invalid base_url %q: %w", f.BaseURL, err)
	}

	if len(f.Currencies) == 0 {
		f.Currencies = []string{"USD"}
	}
	for i, currency := range f.Currencies {
		f.Currencies[i] = strings.ToLower(strings.TrimSpace(currency))
	}
	for i, indicator := range f.Indicators {
		f.Indicators[i] = strings.ToLower(strings.TrimSpace(indicator))
	}

	for _, pair := range f.FXPairs {
		if !strings.Contains(pair, "/") {
			return fmt.Errorf("invalid fx_pair %q: expected the form BASE/QUOTE", pair)
		}
	}

	if len(f.Indicators) == 0 && len(f.FXPairs) == 0 {
		return errors.New("at least one of indicators or fx_pairs must be set")
	}

	if f.ResponseTimeout < config.Duration(time.Second) {
		f.ResponseTimeout = config.Duration(5 * time.Second)
	}

	tlsCfg, err := f.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	f.client = &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
		Timeout:   time.Duration(f.ResponseTimeout),
	}

	return nil
}

func (f *FXMacroData) Gather(acc telegraf.Accumulator) error {
	for _, currency := range f.Currencies {
		for _, indicator := range f.Indicators {
			acc.AddError(f.gatherIndicator(acc, currency, indicator))
		}
	}

	for _, pair := range f.FXPairs {
		base, quote, _ := strings.Cut(pair, "/")
		acc.AddError(f.gatherFX(acc, strings.ToLower(base), strings.ToLower(quote)))
	}

	return nil
}

func (f *FXMacroData) gatherIndicator(acc telegraf.Accumulator, currency, indicator string) error {
	endpoint := fmt.Sprintf("%s/v1/announcements/%s/%s?limit=1", f.BaseURL, currency, indicator)

	series, err := f.fetch(endpoint)
	if err != nil {
		return err
	}
	if len(series.Data) == 0 {
		return nil
	}

	row := series.Data[0]
	if row.Value == nil {
		// The API returns a null value for a period that was not reported. A
		// zero would be indistinguishable from a real reading of zero.
		return nil
	}

	tags := map[string]string{
		"currency":  strings.ToUpper(currency),
		"indicator": indicator,
	}
	if row.Source != "" {
		tags["source"] = row.Source
	}

	fields := map[string]interface{}{"value": *row.Value}
	if row.Date != "" {
		fields["reference_date"] = row.Date
	}

	acc.AddFields("fxmacrodata_indicator", fields, tags, timestamp(row))

	return nil
}

func (f *FXMacroData) gatherFX(acc telegraf.Accumulator, base, quote string) error {
	endpoint := fmt.Sprintf("%s/v1/forex/%s/%s?limit=1", f.BaseURL, base, quote)

	series, err := f.fetch(endpoint)
	if err != nil {
		return err
	}
	if len(series.Data) == 0 || series.Data[0].Value == nil {
		return nil
	}

	row := series.Data[0]
	tags := map[string]string{
		"base":  strings.ToUpper(base),
		"quote": strings.ToUpper(quote),
	}
	fields := map[string]interface{}{"rate": *row.Value}
	if row.Date != "" {
		fields["reference_date"] = row.Date
	}

	acc.AddFields("fxmacrodata_fx", fields, tags, timestamp(row))

	return nil
}

// timestamp prefers the instant the figure was published over collection time.
//
// A macro figure is only meaningful alongside when it became known: the value
// for a given month is not knowable during that month. Stamping the point with
// the publication instant keeps that usable in a time series, and falls back to
// now only when the API does not report one.
func timestamp(row observation) time.Time {
	if row.AnnouncementDatetime != nil && *row.AnnouncementDatetime > 0 {
		return time.Unix(*row.AnnouncementDatetime, 0).UTC()
	}
	return time.Now()
}

func (f *FXMacroData) fetch(endpoint string) (*seriesResponse, error) {
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")

	if !f.APIKey.Empty() {
		key, err := f.APIKey.Get()
		if err != nil {
			return nil, fmt.Errorf("getting api_key failed: %w", err)
		}
		// A header rather than a query parameter, so the key cannot reach a
		// proxy log or an error message quoting the URL.
		request.Header.Set("X-API-Key", key.String())
		key.Destroy()
	}

	resp, err := f.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		// The key does not cover this currency. That is a configuration fact,
		// so report it once per interval and keep gathering the rest.
		return nil, fmt.Errorf("%s is not covered by the configured api_key", endpoint)
	case resp.StatusCode == http.StatusNotFound:
		return &seriesResponse{}, nil
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%s returned status %d", endpoint, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var series seriesResponse
	if err := json.Unmarshal(body, &series); err != nil {
		return nil, fmt.Errorf("parsing response from %s failed: %w", endpoint, err)
	}

	return &series, nil
}

func init() {
	inputs.Add("fxmacrodata", func() telegraf.Input {
		return &FXMacroData{}
	})
}
