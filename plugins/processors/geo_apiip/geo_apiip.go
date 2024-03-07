//go:generate ../../../tools/readme_config_includer/generator
package geo_apiip

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/jellydator/ttlcache/v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

// TODO: switch to HTTPS with paid subscription
const apiURL = "http://apiip.net/api/check"

type Response struct {
	IP                  string `json:"ip"`
	ContinentCode       string `json:"continentCode"`
	ContinentName       string `json:"continentName"`
	CountryCode         string `json:"countryCode"`
	CountryName         string `json:"countryName"`
	CountryNameNative   string `json:"countryNameNative"`
	OfficialCountryName string `json:"officialCountryName"`
	RegionCode          string `json:"regionCode"`
	RegionName          string `json:"regionName"`
	City                string `json:"city"`
	Capital             string `json:"capital"`
	IsEu                bool   `json:"isEu"`
}

type ErrorResponse struct {
	Success bool `json:"success"`
	Message struct {
		Code int    `json:"code"`
		Type string `json:"type"`
		Info string `json:"info"`
	} `json:"message"`
}

type GeoAPI struct {
	APIKey         string          `toml:"api_key"`
	IPTag          string          `toml:"ip_field"`
	IP             string          `toml:"ip"`
	RegionTag      string          `toml:"region_tag"`
	CountryTag     string          `toml:"country_tag"`
	CityTag        string          `toml:"city_tag"`
	UpdateInterval config.Duration `toml:"update_interval"`
	DefaultRegion  string          `toml:"default_region"`
	DefaultCountry string          `toml:"default_country"`
	DefaultCity    string          `toml:"default_city"`
	Log            telegraf.Logger `toml:"-"`
	client         *http.Client
	cache          *ttlcache.Cache[string, location]
}

type location struct {
	Region  string
	Country string
	City    string
}

func (*GeoAPI) SampleConfig() string {
	return sampleConfig
}

func (c *GeoAPI) Apply(in ...telegraf.Metric) []telegraf.Metric {
	var ip string
	for _, m := range in {
		if c.IP != "" {
			ip = c.IP
		} else if c.IPTag != "" {
			if v, ok := m.GetTag(c.IPTag); ok {
				ip = v
			}
		}
		c.Log.Debugf("IP: %v", ip)
		l, err := c.getLocation(ip)
		if err != nil {
			c.Log.Errorf("Error getting location: %v", err)
			if c.DefaultRegion != "" {
				m.AddTag(c.RegionTag, c.DefaultRegion)
			}
			if c.DefaultCountry != "" {
				m.AddTag(c.CountryTag, c.DefaultCountry)
			}
			if c.DefaultCity != "" {
				m.AddTag(c.CityTag, c.DefaultCity)
			}
			continue
		}
		m.AddTag(c.RegionTag, l.Region)
		m.AddTag(c.CountryTag, l.Country)
		m.AddTag(c.CityTag, l.City)
	}
	return in
}

func (c *GeoAPI) Init() error {
	c.Log.Debug("Initializing geo_apiip Processor")

	if c.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	if c.IPTag != "" && c.IP != "" {
		return fmt.Errorf("ip_field and ip are mutually exclusive")
	}

	go c.cache.Start()

	return nil
}

// New returns a new GeoAPI processor with defaults.
func New() *GeoAPI {
	return &GeoAPI{
		RegionTag:      "region",
		CountryTag:     "country",
		CityTag:        "city",
		UpdateInterval: config.Duration(time.Minute * time.Duration(5)),
		client: &http.Client{
			Timeout: time.Second * 5,
		},
		cache: ttlcache.New[string, location](),
	}
}

func validateIP(ip string) string {
	check := net.ParseIP(ip)
	if check == nil {
		return ""
	}
	return ip
}

func (c *GeoAPI) getLocation(ip string) (l location, err error) {
	vip := validateIP(ip)
	cl := c.cache.Get(vip)
	if cl != nil {
		if vip == "" {
			c.Log.Debug("Cache hit for origin")
		} else {
			c.Log.Debugf("Cache hit for %s", vip)
		}
		l = cl.Value()
		return l, nil
	}
	l, err = c.getLocationCall(vip)
	if err == nil {
		c.cache.Set(ip, l, time.Duration(c.UpdateInterval))
	}
	return l, err
}

func (c *GeoAPI) getLocationCall(ip string) (l location, err error) {
	if ip == "" {
		c.Log.Debug("Cache miss for origin")
	} else {
		c.Log.Debugf("Cache miss for %v", ip)
	}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return l, err
	}

	param := req.URL.Query()
	param.Add("format", "json")
	param.Add("language", "en")
	if c.APIKey != "" {
		param.Add("accessKey", c.APIKey)
	}
	if ip != "" {
		param.Add("ip", ip)
	}
	req.URL.RawQuery = param.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return l, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var e ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&e)
		return l, errors.New(e.Message.Info)
	}

	var ar Response
	err = json.NewDecoder(resp.Body).Decode(&ar)
	if err != nil {
		return l, err
	}
	l.Region = ar.ContinentName
	l.Country = ar.CountryCode
	if ar.City != "" {
		l.City = ar.City
	} else {
		l.City = ar.Capital
	}

	return l, err
}

func init() {
	processors.Add("geo_apiip", func() telegraf.Processor {
		return New()
	})
}
