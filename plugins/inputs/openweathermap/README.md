# OpenWeatherMap Input Plugin

Collect current weather and forecast data from OpenWeatherMap.

To use this plugin you will need an [api key][] (app_id).

City identifiers can be found in the [city list][]. Alternately you
can [search][] by name; the `city_id` can be found as the last digits
of the URL: <https://openweathermap.org/city/2643743>. Language
identifiers can be found in the [lang list][]. Documentation for
condition ID, icon, and main is at [weather conditions][].

## Configuration

```toml
# Read current weather and forecasts data from openweathermap.org
[[inputs.openweathermap]]
  ## OpenWeatherMap API key.
  app_id = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

  ## City ID's to collect weather data from.
  city_id = ["5391959"]

  ## Language of the description field. Can be one of "ar", "bg",
  ## "ca", "cz", "de", "el", "en", "fa", "fi", "fr", "gl", "hr", "hu",
  ## "it", "ja", "kr", "la", "lt", "mk", "nl", "pl", "pt", "ro", "ru",
  ## "se", "sk", "sl", "es", "tr", "ua", "vi", "zh_cn", "zh_tw"
  # lang = "en"

  ## APIs to fetch; can contain "weather" or "forecast".
  fetch = ["weather", "forecast"]

  ## OpenWeatherMap base URL
  # base_url = "https://api.openweathermap.org/"

  ## Timeout for HTTP response.
  # response_timeout = "5s"

  ## Preferred unit system for temperature and wind speed. Can be one of
  ## "metric", "imperial", or "standard".
  # units = "metric"

  ## Query interval; OpenWeatherMap weather data is updated every 10
  ## minutes.
  interval = "10m"
```

## Metrics

- weather
  - tags:
    - city_id
    - forecast
    - condition_id
    - condition_main
  - fields:
    - cloudiness (int, percent)
    - humidity (int, percent)
    - pressure (float, atmospheric pressure hPa)
    - rain (float, rain volume for the last 1-3 hours (depending on API response) in mm)
    - sunrise (int, nanoseconds since unix epoch)
    - sunset (int, nanoseconds since unix epoch)
    - temperature (float, degrees)
    - feels_like (float, degrees)
    - visibility (int, meters, not available on forecast data)
    - wind_degrees (float, wind direction in degrees)
    - wind_speed (float, wind speed in meters/sec or miles/sec)
    - condition_description (string, localized long description)
    - condition_icon

## Example Output

```shell
> weather,city=San\ Francisco,city_id=5391959,condition_id=803,condition_main=Clouds,country=US,forecast=114h,host=robot pressure=1027,temperature=10.09,wind_degrees=34,wind_speed=1.24,condition_description="broken clouds",cloudiness=80i,humidity=67i,rain=0,feels_like=8.9,condition_icon="04n" 1645952400000000000
> weather,city=San\ Francisco,city_id=5391959,condition_id=804,condition_main=Clouds,country=US,forecast=117h,host=robot humidity=65i,rain=0,temperature=10.12,wind_degrees=31,cloudiness=90i,pressure=1026,feels_like=8.88,wind_speed=1.31,condition_description="overcast clouds",condition_icon="04n" 1645963200000000000
> weather,city=San\ Francisco,city_id=5391959,condition_id=804,condition_main=Clouds,country=US,forecast=120h,host=robot cloudiness=100i,humidity=61i,rain=0,temperature=10.28,wind_speed=1.94,condition_icon="04d",pressure=1027,feels_like=8.96,wind_degrees=16,condition_description="overcast clouds" 1645974000000000000

```

[api key]: https://openweathermap.org/appid
[city list]: http://bulk.openweathermap.org/sample/city.list.json.gz
[search]: https://openweathermap.org/find
[lang list]: https://openweathermap.org/current#multi
[weather conditions]: https://openweathermap.org/weather-conditions
