//go:generate ../../../tools/readme_config_includer/generator
package fibaro

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/fibaro/hc2"
	"github.com/influxdata/telegraf/plugins/inputs/fibaro/hc3"
)

//go:embed sample.conf
var sampleConfig string

const defaultTimeout = 5 * time.Second

// Fibaro contains connection information
type Fibaro struct {
	URL        string          `toml:"url"`
	Username   string          `toml:"username"`
	Password   string          `toml:"password"`
	Timeout    config.Duration `toml:"timeout"`
	DeviceType string          `toml:"device_type"`

	client *http.Client
}

// getJSON connects, authenticates and reads JSON payload returned by Fibaro box
func (f *Fibaro) getJSON(path string) ([]byte, error) {
	var requestURL = f.URL + path

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(f.Username, f.Password)
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("response from url %q has status code %d (%s), expected %d (%s)",
			requestURL,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	return bodyBytes, nil
}

func (f *Fibaro) Init() error {
	switch f.DeviceType {
	case "":
		f.DeviceType = "HC2"
	case "HC2", "HC3":
	default:
		return fmt.Errorf("invalid option for device type")
	}

	return nil
}

func (*Fibaro) SampleConfig() string {
	return sampleConfig
}

// Gather fetches all required information to output metrics
func (f *Fibaro) Gather(acc telegraf.Accumulator) error {
	if f.client == nil {
		f.client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
			Timeout: time.Duration(f.Timeout),
		}
	}

	sections, err := f.getJSON("/api/sections")
	if err != nil {
		return err
	}
	rooms, err := f.getJSON("/api/rooms")
	if err != nil {
		return err
	}
	devices, err := f.getJSON("/api/devices")
	if err != nil {
		return err
	}

	switch f.DeviceType {
	case "HC2":
		return hc2.Parse(acc, sections, rooms, devices)
	case "HC3":
		return hc3.Parse(acc, sections, rooms, devices)
	}

	return nil
}

func init() {
	inputs.Add("fibaro", func() telegraf.Input {
		return &Fibaro{
			Timeout: config.Duration(defaultTimeout),
		}
	})
}
