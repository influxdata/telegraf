package openweathermap

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleNoContent = `
{
}
`

const sampleStatusResponse = `
{
    "city": {
        "coord": {
            "lat": 48.8534,
            "lon": 2.3488
        },
        "country": "FR",
        "id": 2988507,
        "name": "Paris"
    },
    "cnt": 40,
    "cod": "200",
    "list": [
        {
            "clouds": {
                "all": 88
            },
            "dt": 1543622400,
            "dt_txt": "2018-12-01 00:00:00",
            "main": {
                "grnd_level": 1018.65,
                "humidity": 91,
                "pressure": 1018.65,
                "sea_level": 1030.99,
                "temp": 279.86,
                "temp_kf": -2.14,
                "temp_max": 281.999,
                "temp_min": 279.86
            },
            "rain": {
                "3h": 0.035
            },
            "sys": {
                "pod": "n"
            },
            "weather": [
                {
                    "description": "light rain",
                    "icon": "10n",
                    "id": 500,
                    "main": "Rain"
                }
            ],
            "wind": {
                "deg": 228.501,
                "speed": 3.76
            }
        },
        {
            "clouds": {
                "all": 92
            },
            "dt": 1544043600,
            "dt_txt": "2018-12-05 21:00:00",
            "main": {
                "grnd_level": 1032.18,
                "humidity": 98,
                "pressure": 1032.18,
                "sea_level": 1044.78,
                "temp": 279.535,
                "temp_kf": 0,
                "temp_max": 279.535,
                "temp_min": 279.535
            },
            "rain": {
                "3h": 0.049999999999997
            },
            "sys": {
                "pod": "n"
            },
            "weather": [
                {
                    "description": "light rain",
                    "icon": "10n",
                    "id": 500,
                    "main": "Rain"
                }
            ],
            "wind": {
                "deg": 335.005,
                "speed": 2.66
            }
        }
    ],
    "message": 0.0025
}
`

const sampleWeatherResponse = `
{
    "base": "stations",
    "clouds": {
        "all": 75
    },
    "cod": 200,
    "coord": {
        "lat": 48.85,
        "lon": 2.35
    },
    "dt": 1544194800,
    "id": 2988507,
    "main": {
        "humidity": 87,
        "pressure": 1007,
        "temp": 282.4,
        "temp_max": 283.15,
        "temp_min": 281.15
    },
    "name": "Paris",
    "sys": {
        "country": "FR",
        "id": 6550,
        "message": 0.002,
        "sunrise": 1544167818,
        "sunset": 1544198047,
        "type": 1
    },
    "visibility": 10000,
    "weather": [
        {
            "description": "light intensity drizzle",
            "icon": "09d",
            "id": 300,
            "main": "Drizzle"
        }
    ],
    "wind": {
        "deg": 290,
        "speed": 8.7
    }
}
`

func TestForecastGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/data/2.5/forecast" {
			rsp = sampleStatusResponse
			w.Header()["Content-Type"] = []string{"application/json"}
		} else if r.URL.Path == "/data/2.5/weather" {
			rsp = sampleNoContent
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	n := &OpenWeatherMap{
		BaseUrl: ts.URL,
		AppId:   "noappid",
		Cities:  []string{"2988507"},
	}

	var acc testutil.Accumulator

	err_openweathermap := n.Gather(&acc)
	require.NoError(t, err_openweathermap)

	addr, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	host, port, err := net.SplitHostPort(addr.Host)
	if err != nil {
		host = addr.Host
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}

	acc.AssertContainsTaggedFields(
		t,
		"weather",
		map[string]interface{}{
			"humidity":    int64(91),
			"pressure":    1018.65,
			"temperature": 6.710000000000036,
			"rain":        0.035,
			"wind.deg":    228.501,
			"wind.speed":  3.76,
		},
		map[string]string{
			"server":   host,
			"port":     port,
			"base_url": addr.String(),
			"city_id":  "2988507",
		})
}

func TestWeatherGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/data/2.5/weather" {
			rsp = sampleWeatherResponse
			w.Header()["Content-Type"] = []string{"application/json"}
		} else if r.URL.Path == "/data/2.5/forecast" {
			rsp = sampleNoContent
		} else {
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
	defer ts.Close()

	n := &OpenWeatherMap{
		BaseUrl: ts.URL,
		AppId:   "noappid",
		Cities:  []string{"2988507"},
	}

	var acc testutil.Accumulator

	err_openweathermap := n.Gather(&acc)

	require.NoError(t, err_openweathermap)

	addr, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	host, port, err := net.SplitHostPort(addr.Host)
	if err != nil {
		host = addr.Host
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}

	acc.AssertContainsTaggedFields(
		t,
		"weather",
		map[string]interface{}{
			"humidity":    int64(87),
			"pressure":    1007.0,
			"temperature": 9.25,
			"wind.deg":    290.0,
			"wind.speed":  8.7,
			"rain":        0.0,
		},
		map[string]string{
			"server":   host,
			"port":     port,
			"base_url": addr.String(),
			"city_id":  "2988507",
		})
}
