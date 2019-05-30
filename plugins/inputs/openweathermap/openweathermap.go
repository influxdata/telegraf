package openweathermap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type OpenWeatherMap struct {
	BaseUrl string
	AppId   string
	CityId  []string

	client *http.Client

	ResponseTimeout internal.Duration
	Fetch           []string
	Units           string
}

// https://openweathermap.org/current#severalid
// Call for several city IDs
// The limit of locations is 20.
const owmRequestSeveralCityId int = 20
const defaultResponseTimeout time.Duration = time.Second * 5
const defaultUnits string = "metric"

var sampleConfig = `
  ## Root url of weather map REST API
  base_url = "https://api.openweathermap.org/"
  ## Your personal user token from openweathermap.org
  app_id = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  city_id = ["2988507", "2988588"]

  ## HTTP response timeout (default: 5s)
  response_timeout = "5s"
  fetch = ["weather", "forecast"]
  units = "metric"
  ## Limit OpenWeatherMap query interval. See calls per minute info at: https://openweathermap.org/price
  interval = "10m"
`

func (n *OpenWeatherMap) SampleConfig() string {
	return sampleConfig
}

func (n *OpenWeatherMap) Description() string {
	return "Read current weather and forecasts data from openweathermap.org"
}

func (n *OpenWeatherMap) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	var strs []string

	base, err := url.Parse(n.BaseUrl)
	if err != nil {
		return err
	}

	// Create an HTTP client that is re-used for each
	// collection interval

	if n.client == nil {
		client, err := n.createHttpClient()
		if err != nil {
			return err
		}
		n.client = client
	}
	units := n.Units
	if units == "" {
		units = defaultUnits
	}
	for _, fetch := range n.Fetch {
		if fetch == "forecast" {
			var u *url.URL
			var addr *url.URL

			for _, city := range n.CityId {
				u, err = url.Parse(fmt.Sprintf("/data/2.5/forecast?id=%s&APPID=%s&units=%s", city, n.AppId, units))
				if err != nil {
					acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
					continue
				}
				addr = base.ResolveReference(u)
				wg.Add(1)
				go func(addr *url.URL) {
					defer wg.Done()
					acc.AddError(n.gatherUrl(addr, acc, true))
				}(addr)
			}
		} else if fetch == "weather" {
			j := 0
			for j < len(n.CityId) {
				var u *url.URL
				var addr *url.URL
				strs = make([]string, 0)
				for i := 0; j < len(n.CityId) && i < owmRequestSeveralCityId; i++ {
					strs = append(strs, n.CityId[j])
					j++
				}
				cities := strings.Join(strs, ",")

				u, err = url.Parse(fmt.Sprintf("/data/2.5/group?id=%s&APPID=%s&units=%s", cities, n.AppId, units))
				if err != nil {
					acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
					continue
				}

				addr = base.ResolveReference(u)
				wg.Add(1)
				go func(addr *url.URL) {
					defer wg.Done()
					acc.AddError(n.gatherUrl(addr, acc, false))
				}(addr)
			}

		}
	}

	wg.Wait()
	return nil
}

func (n *OpenWeatherMap) createHttpClient() (*http.Client, error) {

	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = defaultResponseTimeout
	}

	client := &http.Client{
		Transport: &http.Transport{},
		Timeout:   n.ResponseTimeout.Duration,
	}

	return client, nil
}

func (n *OpenWeatherMap) gatherUrl(addr *url.URL, acc telegraf.Accumulator, forecast bool) error {
	resp, err := n.client.Get(addr.String())

	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", addr.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", addr.String(), resp.Status)
	}
	contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
	switch contentType {
	case "application/json":
		err = gatherWeatherUrl(bufio.NewReader(resp.Body), forecast, acc)
		return err
	default:
		return fmt.Errorf("%s returned unexpected content type %s", addr.String(), contentType)
	}
}

type WeatherEntry struct {
	Dt     int64  `json:"dt"`
	Dttxt  string `json:"dt_txt"` // empty for weather/
	Clouds struct {
		All int64 `json:"all"`
	} `json:"clouds"`
	Main struct {
		GrndLevel float64 `json:"grnd_level"` // empty for weather/
		Humidity  int64   `json:"humidity"`
		SeaLevel  float64 `json:"sea_level"` // empty for weather/
		Pressure  float64 `json:"pressure"`
		Temp      float64 `json:"temp"`
		TempMax   float64 `json:"temp_max"`
		TempMin   float64 `json:"temp_min"`
	} `json:"main"`
	Rain struct {
		Rain3 float64 `json:"3h"`
	} `json:"rain"`
	Sys struct {
		Pod     string  `json:"pod"`
		Country string  `json:"country"`
		Message float64 `json:"message"`
		Id      int64   `json:"id"`
		Type    int64   `json:"type"`
		Sunrise int64   `json:"sunrise"`
		Sunset  int64   `json:"sunset"`
	} `json:"sys"`
	Wind struct {
		Deg   float64 `json:"deg"`
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Id          int64  `json:"id"`
		Main        string `json:"main"`
	} `json:"weather"`

	// Additional entries for weather/
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Coord struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"coord"`
	Visibility int64 `json:"visibility"`
}

type Status struct {
	City struct {
		Coord struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"coord"`
		Country string `json:"country"`
		Id      int64  `json:"id"`
		Name    string `json:"name"`
	} `json:"city"`
	List []WeatherEntry `json:"list"`
}

func gatherWeatherUrl(r *bufio.Reader, forecast bool, acc telegraf.Accumulator) error {
	dec := json.NewDecoder(r)
	status := &Status{}
	if err := dec.Decode(status); err != nil {
		return fmt.Errorf("Error while decoding JSON response: %s", err)
	}
	status.Gather(forecast, acc)
	return nil
}

func (s *Status) Gather(forecast bool, acc telegraf.Accumulator) {
	tags := map[string]string{
		"city_id":  strconv.FormatInt(s.City.Id, 10),
		"forecast": "*",
	}

	for i, e := range s.List {
		tm := time.Unix(e.Dt, 0)
		if e.Id > 0 {
			tags["city_id"] = strconv.FormatInt(e.Id, 10)
		}
		if forecast {
			tags["forecast"] = fmt.Sprintf("%dh", (i+1)*3)
		}
		acc.AddFields(
			"weather",
			map[string]interface{}{
				"rain":         e.Rain.Rain3,
				"wind_degrees": e.Wind.Deg,
				"wind_speed":   e.Wind.Speed,
				"humidity":     e.Main.Humidity,
				"pressure":     e.Main.Pressure,
				"temperature":  e.Main.Temp,
			},
			tags,
			tm)
	}
	if forecast {
		// intentional: overwrite future data points
		// under the * tag
		tags := map[string]string{
			"city_id":  strconv.FormatInt(s.City.Id, 10),
			"forecast": "*",
		}
		for _, e := range s.List {
			tm := time.Unix(e.Dt, 0)
			if e.Id > 0 {
				tags["city_id"] = strconv.FormatInt(e.Id, 10)
			}
			acc.AddFields(
				"weather",
				map[string]interface{}{
					"rain":         e.Rain.Rain3,
					"wind_degrees": e.Wind.Deg,
					"wind_speed":   e.Wind.Speed,
					"humidity":     e.Main.Humidity,
					"pressure":     e.Main.Pressure,
					"temperature":  e.Main.Temp,
				},
				tags,
				tm)
		}
	}
}

func init() {
	inputs.Add("openweathermap", func() telegraf.Input {
		tmout := internal.Duration{
			Duration: defaultResponseTimeout,
		}
		return &OpenWeatherMap{
			ResponseTimeout: tmout,
			Units:           defaultUnits,
		}
	})
}
