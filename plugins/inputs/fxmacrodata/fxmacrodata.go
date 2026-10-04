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
	"slices"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type FXMacroData struct {
	URL             string          `toml:"url"`
	APIKey          config.Secret   `toml:"api_key"`
	Currencies      []string        `toml:"currencies"`
	Indicators      []string        `toml:"indicators"`
	FXPairs         []string        `toml:"fx_pairs"`
	ResponseTimeout config.Duration `toml:"response_timeout"`
	Log             telegraf.Logger `toml:"-"`
	tls.ClientConfig

	client *http.Client
	series []series
}

func (*FXMacroData) SampleConfig() string {
	return sampleConfig
}

func (f *FXMacroData) Init() error {
	// Set defaults
	if f.URL == "" {
		f.URL = "https://api.fxmacrodata.com/v1"
	}
	if len(f.Currencies) == 0 {
		f.Currencies = []string{"USD"}
	}

	// Check settings
	baseURL, err := url.Parse(f.URL)
	if err != nil {
		return fmt.Errorf("parsing url %q failed: %w", f.URL, err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return fmt.Errorf("invalid scheme %q in url, expected http or https", baseURL.Scheme)
	}

	if len(f.Indicators) == 0 && len(f.FXPairs) == 0 {
		return errors.New("at least one of indicators or fx_pairs must be set")
	}

	for _, currency := range f.Currencies {
		if isInvalidCurrency(currency) {
			return fmt.Errorf("invalid currency %q, expected an uppercase three-letter code like \"USD\"", currency)
		}
	}

	if slices.Contains(f.Indicators, "") {
		return errors.New("empty entry in indicators")
	}

	if f.ResponseTimeout < 0 {
		return errors.New("response_timeout must not be negative")
	}

	// Build the list of series to query, so gathering does not need to
	// assemble the endpoints again in every cycle
	f.series = make([]series, 0, len(f.Currencies)*len(f.Indicators)+len(f.FXPairs))
	for _, currency := range f.Currencies {
		for _, indicator := range f.Indicators {
			endpoint := baseURL.JoinPath("announcements", strings.ToLower(currency), indicator)
			endpoint.RawQuery = "limit=1"
			f.series = append(f.series, series{
				endpoint: endpoint.String(),
				name:     "fxmacrodata_indicator",
				field:    "value",
				tags:     map[string]string{"currency": currency, "indicator": indicator},
			})
		}
	}

	for _, pair := range f.FXPairs {
		base, quote, found := strings.Cut(pair, "/")
		if !found || isInvalidCurrency(base) || isInvalidCurrency(quote) {
			return fmt.Errorf("invalid fx_pair %q, expected the form \"BASE/QUOTE\" in uppercase like \"EUR/USD\"", pair)
		}

		endpoint := baseURL.JoinPath("forex", strings.ToLower(base), strings.ToLower(quote))
		endpoint.RawQuery = "limit=1"
		f.series = append(f.series, series{
			endpoint: endpoint.String(),
			name:     "fxmacrodata_fx",
			field:    "rate",
			tags:     map[string]string{"base": base, "quote": quote},
		})
	}

	tlsCfg, err := f.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("creating TLS configuration failed: %w", err)
	}

	f.client = &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
		Timeout:   time.Duration(f.ResponseTimeout),
	}

	return nil
}

func (f *FXMacroData) Gather(acc telegraf.Accumulator) error {
	// Series that do not exist are dropped from the list, so they are not
	// queried again in the next cycles
	active := make([]series, 0, len(f.series))
	for _, s := range f.series {
		exists, err := f.gatherSeries(acc, s)
		acc.AddError(err)
		if !exists {
			f.Log.Warnf("Series %s does not exist, removing it from the list of queried series", s.endpoint)
			continue
		}
		active = append(active, s)
	}
	f.series = active

	return nil
}

// gatherSeries queries a single series and adds the latest data point to the
// accumulator. The returned flag is false if the series does not exist.
func (f *FXMacroData) gatherSeries(acc telegraf.Accumulator, s series) (bool, error) {
	request, err := http.NewRequest("GET", s.endpoint, nil)
	if err != nil {
		return true, fmt.Errorf("creating request for %s failed: %w", s.endpoint, err)
	}
	request.Header.Set("Accept", "application/json")

	if !f.APIKey.Empty() {
		key, err := f.APIKey.Get()
		if err != nil {
			return true, fmt.Errorf("getting api_key failed: %w", err)
		}

		// Send the key as a header rather than a query parameter, so it
		// cannot end up in a proxy log or in an error message quoting the URL
		request.Header.Set("X-API-Key", key.String())
		key.Destroy()
	}

	resp, err := f.client.Do(request)
	if err != nil {
		return true, fmt.Errorf("querying %s failed: %w", s.endpoint, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// Do nothing, everything is alright
	case http.StatusUnauthorized, http.StatusForbidden:
		// The key does not cover this series. Report it and keep gathering
		// the remaining series.
		return true, fmt.Errorf("%s is not covered by the configured api_key", s.endpoint)
	case http.StatusNotFound:
		// The series does not exist for this currency
		return false, nil
	default:
		return true, fmt.Errorf("querying %s returned %q", s.endpoint, http.StatusText(resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return true, fmt.Errorf("reading response from %s failed: %w", s.endpoint, err)
	}

	var data seriesResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return true, fmt.Errorf("parsing response from %s failed: %w", s.endpoint, err)
	}

	if len(data.Data) == 0 {
		return true, nil
	}

	point := data.Data[0]
	if point.Value == nil {
		// A null value means the period was not reported. There is no value
		// we could write instead that is distinguishable from a real reading,
		// so skip the point.
		f.Log.Debugf("Skipping %s for %q as the value is not reported", s.endpoint, point.Date)
		return true, nil
	}
	value := *point.Value

	tags := make(map[string]string, len(s.tags)+1)
	for k, v := range s.tags {
		tags[k] = v
	}
	if s.name == "fxmacrodata_indicator" {
		tags["source"] = data.Source
	}

	fields := map[string]any{
		s.field:          value,
		"reference_date": point.Date,
	}

	// Use the instant the figure was published as the metric time and only
	// fall back to the collection time if the API does not report one
	timestamp := time.Now()
	if point.AnnouncementDatetime != nil && *point.AnnouncementDatetime > 0 {
		timestamp = time.Unix(*point.AnnouncementDatetime, 0).UTC()
	}

	acc.AddFields(s.name, fields, tags, timestamp)

	return true, nil
}

func isInvalidCurrency(code string) bool {
	if len(code) != 3 {
		return true
	}
	for _, c := range code {
		if c < 'A' || c > 'Z' {
			return true
		}
	}

	return false
}

func init() {
	inputs.Add("fxmacrodata", func() telegraf.Input {
		return &FXMacroData{
			ResponseTimeout: config.Duration(5 * time.Second),
		}
	})
}
