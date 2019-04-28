package openweathermap

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Cities  []string

	client *http.Client

	ResponseTimeout internal.Duration
	ForecastEnable  bool
}

// https://openweathermap.org/current#severalid
// Call for several city IDs
// The limit of locations is 20.
const OWM_REQUEST_SEVERAL_CITY_IDS int = 20
const DEFAULT_RESPONSE_TIMEOUT time.Duration = time.Second * 5

var sampleConfig = `
  ## Root url of weather map REST API
  base_url = "http://api.openweathermap.org/"
  # Your personal user token from openweathermap.org
  app_id = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  cities = ["2988507", "2988588"]

  # HTTP response timeout (default: 5s)
  response_timeout = "5s"
  forecast_enable = true
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

	if n.ForecastEnable {
		var u *url.URL
		var addr *url.URL

		tags := map[string]string{
			"forecast": "true",
		}
		for _, city := range n.Cities {
			u, err = url.Parse(fmt.Sprintf("/data/2.5/forecast?id=%s&APPID=%s", city, n.AppId))
			if err != nil {
				acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
				continue
			}
			addr = base.ResolveReference(u)
			wg.Add(1)
			go func(addr *url.URL) {
				defer wg.Done()
				acc.AddError(n.gatherUrl(addr, acc, tags))
			}(addr)
		}
	}
	j := 0
	for j < len(n.Cities) {
		var u *url.URL
		var addr *url.URL
		strs = make([]string, 0)
		for i := 0; j < len(n.Cities) && i < OWM_REQUEST_SEVERAL_CITY_IDS; i++ {
			strs = append(strs, n.Cities[j])
			j++
		}
		cities := strings.Join(strs, ",")

		tags := map[string]string{
			"forecast": "false",
		}
		u, err = url.Parse(fmt.Sprintf("/data/2.5/group?id=%s&APPID=%s", cities, n.AppId))
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse address '%s': %s", u, err))
			continue
		}

		addr = base.ResolveReference(u)
		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			acc.AddError(n.gatherUrl(addr, acc, tags))
		}(addr)
	}

	wg.Wait()
	return nil
}

func (n *OpenWeatherMap) createHttpClient() (*http.Client, error) {

	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = DEFAULT_RESPONSE_TIMEOUT
	}

	client := &http.Client{
		Transport: &http.Transport{},
		Timeout:   n.ResponseTimeout.Duration,
	}

	return client, nil
}

func (n *OpenWeatherMap) gatherUrl(addr *url.URL, acc telegraf.Accumulator, tags map[string]string) error {
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
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		err = gatherStatusUrl(bufio.NewReader(resp.Body), tags, acc)
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		err = gatherWeatherUrl(bufio.NewReader(resp.Body), tags, acc)
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

func gatherStatusUrl(r *bufio.Reader, tags map[string]string, acc telegraf.Accumulator) error {
	dec := json.NewDecoder(r)
	status := &Status{}
	if err := dec.Decode(status); err != nil {
		return fmt.Errorf("Error while decoding JSON response: %s", err)
	}
	status.Gather(tags, acc)
	return nil
}

func gatherWeatherUrl(r *bufio.Reader, tags map[string]string, acc telegraf.Accumulator) error {
	dec := json.NewDecoder(r)
	status := &Status{}
	e := WeatherEntry{}
	if err := dec.Decode(&e); err != nil {
		return fmt.Errorf("Error while decoding JSON response: %s", err)
	}
	status.List = make([]WeatherEntry, 0)
	if len(e.Name) > 0 {
		status.List = append(status.List, e)
		status.City.Coord.Lat = e.Coord.Lat
		status.City.Coord.Lon = e.Coord.Lon
		status.City.Id = e.Id
		status.City.Name = e.Name
		status.Gather(tags, acc)
	}
	return nil
}

func (s *Status) Gather(tags map[string]string, acc telegraf.Accumulator) {
	tags["city_id"] = strconv.FormatInt(s.City.Id, 10)
	for _, e := range s.List {
		tm := time.Unix(e.Dt, 0)
		if e.Id > 0 {
			tags["city_id"] = strconv.FormatInt(e.Id, 10)
		}
		acc.AddFields(
			"weather",
			map[string]interface{}{
				"rain":        e.Rain.Rain3,
				"wind_deg":    e.Wind.Deg,
				"wind_speed":  e.Wind.Speed,
				"humidity":    e.Main.Humidity,
				"pressure":    e.Main.Pressure,
				"temperature": e.Main.Temp - 273.15, // Kelvin to Celsius
			},
			tags,
			tm)
	}
}

func init() {
	inputs.Add("openweathermap", func() telegraf.Input {
		tmout := internal.Duration{
			Duration: DEFAULT_RESPONSE_TIMEOUT,
		}
		return &OpenWeatherMap{
			ResponseTimeout: tmout,
			ForecastEnable:  true,
		}
	})
}
