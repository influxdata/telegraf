package openweathermap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
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
                "temp": 6.71,
                "temp_kf": -2.14
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
                "temp": 6.38,
                "temp_kf": 0
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

const groupWeatherResponse = `
{
    "cnt": 1,
    "list": [{
		"clouds": {
			"all": 0
		},
        "coord": {
            "lat": 48.85,
            "lon": 2.35
        },
        "dt": 1544194800,
        "id": 2988507,
        "main": {
            "humidity": 87,
            "pressure": 1007,
            "temp": 9.25
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
    }]
}
`

const batchWeatherResponse = `
{
	"cnt": 3,
	"list": [{
		"coord": {
			"lon": 37.62,
			"lat": 55.75
		},
		"sys": {
			"type": 1,
			"id": 9029,
			"message": 0.0061,
			"country": "RU",
			"sunrise": 1556416455,
			"sunset": 1556470779
		},
		"weather": [{
			"id": 802,
			"main": "Clouds",
			"description": "scattered clouds",
			"icon": "03d"
		}],
		"main": {
			"temp": 9.57,
			"pressure": 1014,
			"humidity": 46
		},
		"visibility": 10000,
		"wind": {
			"speed": 5,
			"deg": 60
		},
		"clouds": {
			"all": 40
		},
		"dt": 1556444155,
		"id": 524901,
		"name": "Moscow"
	}, {
		"coord": {
			"lon": 30.52,
			"lat": 50.43
		},
		"sys": {
			"type": 1,
			"id": 8903,
			"message": 0.0076,
			"country": "UA",
			"sunrise": 1556419155,
			"sunset": 1556471486
		},
		"weather": [{
			"id": 520,
			"main": "Rain",
			"description": "light intensity shower rain",
			"icon": "09d"
		}],
		"main": {
			"temp": 19.29,
			"pressure": 1009,
			"humidity": 63
		},
		"visibility": 10000,
		"wind": {
			"speed": 1
		},
		"clouds": {
			"all": 0
		},
		"dt": 1556444155,
		"id": 703448,
		"name": "Kiev"
	}, {
		"coord": {
			"lon": -0.13,
			"lat": 51.51
		},
		"sys": {
			"type": 1,
			"id": 1414,
			"message": 0.0088,
			"country": "GB",
			"sunrise": 1556426319,
			"sunset": 1556479032
		},
		"weather": [{
			"id": 803,
			"main": "Clouds",
			"description": "broken clouds",
			"icon": "04d"
		}],
		"main": {
			"temp": 10.62,
			"pressure": 1019,
			"humidity": 66
		},
		"visibility": 10000,
		"wind": {
			"speed": 6.2,
			"deg": 290
		},
		"rain": {
			"3h": 0.072
		},
		"clouds": {
			"all": 75
		},
		"dt": 1556444155,
		"id": 2643743,
		"name": "London"
	}]
}
`

func TestForecastGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/data/2.5/forecast" {
			rsp = sampleStatusResponse
			w.Header()["Content-Type"] = []string{"application/json"}
		} else if r.URL.Path == "/data/2.5/group" {
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
		CityId:  []string{"2988507"},
		Fetch:   []string{"weather", "forecast"},
		Units:   "metric",
	}

	var acc testutil.Accumulator

	err := n.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"weather",
			map[string]string{
				"city_id":  "2988507",
				"forecast": "3h",
				"city":     "Paris",
				"country":  "FR",
			},
			map[string]interface{}{
				"cloudiness":   int64(88),
				"humidity":     int64(91),
				"pressure":     1018.65,
				"temperature":  6.71,
				"rain":         0.035,
				"wind_degrees": 228.501,
				"wind_speed":   3.76,
			},
			time.Unix(1543622400, 0),
		),
		testutil.MustMetric(
			"weather",
			map[string]string{
				"city_id":  "2988507",
				"forecast": "6h",
				"city":     "Paris",
				"country":  "FR",
			},
			map[string]interface{}{
				"cloudiness":   int64(92),
				"humidity":     int64(98),
				"pressure":     1032.18,
				"temperature":  6.38,
				"rain":         0.049999999999997,
				"wind_degrees": 335.005,
				"wind_speed":   2.66,
			},
			time.Unix(1544043600, 0),
		),
	}

	testutil.RequireMetricsEqual(t,
		expected, acc.GetTelegrafMetrics(),
		testutil.SortMetrics())
}

func TestWeatherGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/data/2.5/group" {
			rsp = groupWeatherResponse
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
		CityId:  []string{"2988507"},
		Fetch:   []string{"weather"},
		Units:   "metric",
	}

	var acc testutil.Accumulator

	err := n.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"weather",
			map[string]string{
				"city_id":  "2988507",
				"forecast": "*",
				"city":     "Paris",
				"country":  "FR",
			},
			map[string]interface{}{
				"cloudiness":   int64(0),
				"humidity":     int64(87),
				"pressure":     1007.0,
				"temperature":  9.25,
				"rain":         0.0,
				"sunrise":      int64(1544167818000000000),
				"sunset":       int64(1544198047000000000),
				"wind_degrees": 290.0,
				"wind_speed":   8.7,
				"visibility":   10000,
			},
			time.Unix(1544194800, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestBatchWeatherGeneratesMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string
		if r.URL.Path == "/data/2.5/group" {
			rsp = batchWeatherResponse
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
		CityId:  []string{"524901", "703448", "2643743"},
		Fetch:   []string{"weather"},
		Units:   "metric",
	}

	var acc testutil.Accumulator

	err := n.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"weather",
			map[string]string{
				"city_id":  "524901",
				"forecast": "*",
				"city":     "Moscow",
				"country":  "RU",
			},
			map[string]interface{}{
				"cloudiness":   40,
				"humidity":     int64(46),
				"pressure":     1014.0,
				"temperature":  9.57,
				"wind_degrees": 60.0,
				"wind_speed":   5.0,
				"rain":         0.0,
				"sunrise":      int64(1556416455000000000),
				"sunset":       int64(1556470779000000000),
				"visibility":   10000,
			},
			time.Unix(1556444155, 0),
		),
		testutil.MustMetric(
			"weather",
			map[string]string{
				"city_id":  "703448",
				"forecast": "*",
				"city":     "Kiev",
				"country":  "UA",
			},
			map[string]interface{}{
				"cloudiness":   0,
				"humidity":     int64(63),
				"pressure":     1009.0,
				"temperature":  19.29,
				"wind_degrees": 0.0,
				"wind_speed":   1.0,
				"rain":         0.0,
				"sunrise":      int64(1556419155000000000),
				"sunset":       int64(1556471486000000000),
				"visibility":   10000,
			},
			time.Unix(1556444155, 0),
		),
		testutil.MustMetric(
			"weather",
			map[string]string{
				"city_id":  "2643743",
				"forecast": "*",
				"city":     "London",
				"country":  "GB",
			},
			map[string]interface{}{
				"cloudiness":   75,
				"humidity":     int64(66),
				"pressure":     1019.0,
				"temperature":  10.62,
				"wind_degrees": 290.0,
				"wind_speed":   6.2,
				"rain":         0.072,
				"sunrise":      int64(1556426319000000000),
				"sunset":       int64(1556479032000000000),
				"visibility":   10000,
			},
			time.Unix(1556444155, 0),
		),
	}
	testutil.RequireMetricsEqual(t,
		expected, acc.GetTelegrafMetrics(),
		testutil.SortMetrics())
}
