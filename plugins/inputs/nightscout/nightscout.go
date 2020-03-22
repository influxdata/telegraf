package nightscout

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// sampleConfig will provide and example and documentation for configuring the plugin
var sampleConfig = `
[[outputs.influxdb]]
  servers = ["http://137.135.67.239:8086"]
  database = "mydb"
  skip_database_creation = true
  timeout = "10s"
`

// BGData will hold all stats fetched from the database of a nightscout site
type BGData struct {
	Type       string `json:"type"`
	DateString string `json:"dateString"`
	Date       int    `json:"date"`
	SGV        int    `json:"sgv"`
	Direction  string `json:"direction"`
	Noise      int    `json:"noise"`
	Filtered   int    `json:"filtered"`
	Unfiltered int    `json:"unfiltered"`
	RSSI       int    `json:"rssi"`
}

// Nightscout holds configuration and query parameters
type Nightscout struct {
	Servers []string // Slice of servers to query
	Token   string   // Web token generated from /admin
	Secret  string   // API_SECRET as a SHA1 hash
	Count   string   // how many results to fetch

	tls.ClientConfig // TLS config

	client *http.Client // HTTP client
}

// SampleConfig returns the SampleConfig to aid in configuration of the plugin
func (ns *Nightscout) SampleConfig() string {
	return sampleConfig
}

// Description is a one-liner about the plugin
func (ns *Nightscout) Description() string {
	return "Fetch BG data from a nightscout site"
}

// Gather reads blood glucose data from all configured servers accumulates
// stats. Returns any errors encountered while gather stats (if any).
func (ns *Nightscout) Gather(acc telegraf.Accumulator) error {

	for _, s := range ns.Servers {

		acc.AddFields("server", map[string]interface{}{"Server": s}, nil)
		req, err := http.NewRequest("GET", s, nil)
		if err != nil {
			acc.AddError(err)
		}

		req.Header.Set("api_secret", ns.Secret)
		req.Header.Set("accept", "application/json")

		q := req.URL.Query()
		q.Add("count", ns.Count)
		q.Add("token", ns.Token)
		req.URL.RawQuery = q.Encode()

		if ns.client == nil {
			tlsCfg, err := ns.ClientConfig.TLSConfig()
			if err != nil {
				return err
			}
			tr := &http.Transport{
				ResponseHeaderTimeout: time.Duration(3 * time.Second),
				TLSClientConfig:       tlsCfg,
			}
			client := &http.Client{
				Transport: tr,
				Timeout:   time.Duration(4 * time.Second),
			}
			ns.client = client
		}

		res, err := ns.client.Do(req)

		if err != nil {
			acc.AddError(err)
		} else {
			err := ns.importNSResult(res.Body, acc)
			if err != nil {
				acc.AddError(err)
			}
			return nil
		}
	}
	return nil
}

// importNSResult will parse the json body into a standard format
func (ns *Nightscout) importNSResult(r io.Reader, acc telegraf.Accumulator) error {
	now := time.Now()

	body, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(body))

	fields := make(map[string]interface{})

	tags := map[string]string{
		"person": "batman",
	}

	bgData := make([]BGData, 0)

	err = json.Unmarshal(body, &bgData)
	if err != nil {
		fmt.Println("whoops:", err)
	}

	fields["type"] = bgData[0].Type
	fields["dateString"] = bgData[0].DateString
	fields["date"] = bgData[0].Date
	fields["sgv"] = bgData[0].SGV
	fields["direction"] = bgData[0].Direction
	fields["noise"] = bgData[0].Noise
	fields["filtered"] = bgData[0].Filtered
	fields["unfiltered"] = bgData[0].Unfiltered
	fields["rssi"] = bgData[0].RSSI
	fields["direction_num"] = directionMapping(bgData[0].Direction)

	acc.AddFields("nightscout", fields, tags, now)

	return err
}

// init registers the plugin with telegraf
func init() {
	inputs.Add("nightscout", func() telegraf.Input { return &Nightscout{} })
}

// directionMapping converts the text direction to a number for easy processing in dashboards
func directionMapping(direction string) string {
	ans := ""
	switch direction {
	case "DoubleUp":
		ans = "3"
	case "SingleUp":
		ans = "2"
	case "FortyFiveUp":
		ans = "1"
	case "Flat":
		ans = "0"
	case "FortyFiveDown":
		ans = "-1"
	case "SingleDown":
		ans = "-2"
	case "DoubleDown":
		ans = "-3"
	default:
		ans = ""
	}
	return ans
}
