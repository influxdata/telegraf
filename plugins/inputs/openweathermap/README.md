# OpenWeatherMap Input Plugin

Collect current weather and forecast data from OpenWeatherMap.

To use this plugin you will need an [api key][] (app_id).

City identifiers can be found in the [city list][]. Alternately you can
[search][] by name; the `city_id` can be found as the last digits of the URL:
https://openweathermap.org/city/2643743

### Configuration

```toml
[[inputs.openweathermap]]
  ## OpenWeatherMap API key.
  app_id = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

  ## City ID's to collect weather data from.
  city_id = ["5391959"]

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

### Metrics

- weather
  - tags:
    - city_id
    - forecast
  - fields:
    - cloudiness (int, percent)
    - humidity (int, percent)
    - pressure (float, atmospheric pressure hPa)
    - rain (float, rain volume for the last 3 hours in mm)
    - sunrise (int, nanoseconds since unix epoch)
    - sunset (int, nanoseconds since unix epoch)
    - temperature (float, degrees)
    - visibility (int, meters, not available on forecast data)
    - wind_degrees (float, wind direction in degrees)
    - wind_speed (float, wind speed in meters/sec or miles/sec)


### Example Output

```
> weather,city=San\ Francisco,city_id=5391959,country=US,forecast=* cloudiness=40i,humidity=72i,pressure=1013,rain=0,sunrise=1559220629000000000i,sunset=1559273058000000000i,temperature=13.31,visibility=16093i,wind_degrees=280,wind_speed=4.6 1559268695000000000
> weather,city=San\ Francisco,city_id=5391959,country=US,forecast=3h cloudiness=0i,humidity=86i,pressure=1012.03,rain=0,temperature=10.69,wind_degrees=222.855,wind_speed=2.76 1559271600000000000
> weather,city=San\ Francisco,city_id=5391959,country=US,forecast=6h cloudiness=11i,humidity=93i,pressure=1012.79,rain=0,temperature=9.34,wind_degrees=212.685,wind_speed=1.85 1559282400000000000
```

[api key]: https://openweathermap.org/appid
[city list]: http://bulk.openweathermap.org/sample/city.list.json.gz
[search]: https://openweathermap.org/find
