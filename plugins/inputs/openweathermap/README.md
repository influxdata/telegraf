# OpenWeatherMap Input Plugin

This plugin collects weather and forecast data from the
[OpenWeatherMap][openweathermap] service.

> [!IMPORTANT]
> To use this plugin you will need an [APP-ID][api_key] to work.

⭐ Telegraf v1.11.0
🏷️ applications, web
💻 all

[openweathermap]: https://openweathermap.org
[api_key]: https://openweathermap.org/appid

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
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
  # fetch = ["weather", "forecast"]

  ## OpenWeatherMap base URL
  # base_url = "https://api.openweathermap.org/"

  ## Timeout for HTTP response.
  # response_timeout = "5s"

  ## Preferred unit system for temperature and wind speed. Can be one of
  ## "metric", "imperial", or "standard".
  ## The default is "metric" if not specified.
  # units = "metric"

  ## Style to query the current weather; available options
  ##   batch      -- query multiple cities at once using the "group" endpoint
  ##   individual -- query each city individually using the "weather" endpoint
  ## You should use "individual" here as it is documented and provides more
  ## frequent updates. The default is "batch" for backward compatibility.
  # query_style = "batch"

  ## Query interval to fetch data.
  ## By default the global 'interval' setting is used. You should override the
  ## interval here if the global setting is shorter than 10 minutes as
  ## OpenWeatherMap weather data is only updated every 10 minutes.
  # interval = "10m"
```

City identifiers can be found in the [city list file][city_list] or you search
your city by name on the [OpenWeatherMap website][openweathermap] and use the
numeric last element of the resulting URL.
Language identifiers can be found in the [API documentation][languages].

[city_list]: http://bulk.openweathermap.org/sample/city.list.json.gz
[languages]: https://openweathermap.org/current#multi

## Metrics

- weather
  - tags:
    - city_id
    - forecast
    - condition_id
    - condition_main
  - fields:
    - cloudiness            (int, percent)
    - humidity              (int, percent)
    - pressure              (float)       - atmospheric pressure hPa
    - rain                  (float)       - rain volume in mm for the last 1-3h
                                            (depending on API response)
    - snow                  (float)       - snow volume in mm for the last 1-3h
                                            (depending on API response)
    - sunrise               (int)         - nanoseconds since unix epoch
    - sunset                (int)         - nanoseconds since unix epoch
    - temperature           (float, degrees)
    - feels_like            (float, degrees)
    - visibility            (int, meters) - not available on forecast data
    - wind_degrees          (float)       - wind direction in degrees
    - wind_speed            (float)       - wind speed in meters/sec or miles/sec
    - condition_description (string, localized long description)
    - condition_icon

Documentation for condition ID, icon, and main is can be found in the
[documentation][weather_conditions].

[weather_conditions]: https://openweathermap.org/weather-conditions

## Example Output

```text
weather,city=San\ Francisco,city_id=5391959,condition_id=803,condition_main=Clouds,country=US,forecast=114h,host=robot pressure=1027,temperature=10.09,wind_degrees=34,wind_speed=1.24,condition_description="broken clouds",cloudiness=80i,humidity=67i,rain=0,feels_like=8.9,condition_icon="04n" 1645952400000000000
weather,city=San\ Francisco,city_id=5391959,condition_id=804,condition_main=Clouds,country=US,forecast=117h,host=robot humidity=65i,rain=0,temperature=10.12,wind_degrees=31,cloudiness=90i,pressure=1026,feels_like=8.88,wind_speed=1.31,condition_description="overcast clouds",condition_icon="04n" 1645963200000000000
weather,city=San\ Francisco,city_id=5391959,condition_id=804,condition_main=Clouds,country=US,forecast=120h,host=robot cloudiness=100i,humidity=61i,rain=0,temperature=10.28,wind_speed=1.94,condition_icon="04d",pressure=1027,feels_like=8.96,wind_degrees=16,condition_description="overcast clouds" 1645974000000000000
```
