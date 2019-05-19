# Telegraf Plugin: openweathermap

OpenWeatherMap provides the current weather and forecasts for more than 200,000 cities. To use this plugin you will need a token. For more information [click here](https://openweathermap.org/appid).

Find city identifiers in this [list](http://bulk.openweathermap.org/sample/city.list.json.gz). You can also use this [url](https://openweathermap.org/find) as an alternative to downloading a file. The ID is in the url of the city: `https://openweathermap.org/city/2643743`

### Configuration:

```toml
[[inputs.openweathermap]]
  ## Root url of API to pull stats
  # base_url = "https://api.openweathermap.org/data/2.5/"
  ## Your personal user token from openweathermap.org
  # app_id = "xxxxxxxxxxxxxxxxxxxxxxx"
  ## List of city identifiers
  # city_id = ["2988507", "519188"]
  ## HTTP response timeout (default: 5s)
  # response_timeout = "5s"
  ## Query the current weather and future forecast
  # fetch = ["weather", "forecast"]
  ## For temperature in Fahrenheit use units=imperial
  ## For temperature in Celsius use units=metric (default)
  # units = "metric"
```

### Metrics:

+ weather
  - fields:
    - humidity (int, Humidity percentage)
    - temperature (float, Unit: Celcius)
    - pressure (float, Atmospheric pressure in hPa)
    - rain (float, Rain volume for the last 3 hours, mm)
    - wind_speed (float, Wind speed. Unit Default: meter/sec)
    - wind_degrees (float,  Wind direction, degrees)
  - tags:
    - city_id
    - forecast

### Example Output:

Using this configuration:
```toml
[[inputs.openweathermap]]
  base_url = "https://api.openweathermap.org/data/2.5/"
  app_id = "change_this_with_your_appid"
  city_id = ["2988507", "519188"]
  response_timeout = "5s"
  fetch = ["weather", "forecast"]
  units = "metric"
```

When run with:
```
./telegraf -config telegraf.conf -input-filter openweathermap -test
```

It produces data similar to:
```
> weather,city_id=4303602,forecast=* humidity=51i,pressure=1012,rain=0,temperature=16.410000000000025,wind_degrees=170,wind_speed=2.6 1556393944000000000
> weather,city_id=2988507,forecast=* humidity=87i,pressure=1020,rain=0,temperature=7.110000000000014,wind_degrees=260,wind_speed=5.1 1556393841000000000
> weather,city_id=2988507,forecast=3h humidity=69i,pressure=1020.38,rain=0,temperature=5.650000000000034,wind_degrees=268.456,wind_speed=5.83 1556398800000000000
> weather,city_id=2988507,forecast=* humidity=69i,pressure=1020.38,rain=0,temperature=5.650000000000034,wind_degrees=268.456,wind_speed=5.83 1556398800000000000
> weather,city_id=2988507,forecast=6h humidity=74i,pressure=1020.87,rain=0,temperature=5.810000000000002,wind_degrees=261.296,wind_speed=5.43 1556409600000000000
> weather,city_id=2988507,forecast=* humidity=74i,pressure=1020.87,rain=0,temperature=5.810000000000002,wind_degrees=261.296,wind_speed=5.43 1556409600000000000
> weather,city_id=4303602,forecast=9h humidity=66i,pressure=1010.63,rain=0,temperature=14.740000000000009,wind_degrees=196.264,wind_speed=4.3 1556398800000000000
> weather,city_id=4303602,forecast=* humidity=66i,pressure=1010.63,rain=0,temperature=14.740000000000009,wind_degrees=196.264,wind_speed=4.3 1556398800000000000
```





