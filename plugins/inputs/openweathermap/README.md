# Telegraf Plugin: openweathermap

OpenWeatherMap provides the current weather and forecasts for more than 200,000 cities. To use this plugin you will need a token. For more information [click here](https://openweathermap.org/appid).

Find city identifiers in this [list](http://bulk.openweathermap.org/sample/city.list.json.gz)

### Configuration:

```
[[inputs.openweathermap]]
  ## Root url of API to pull stats
  base_url = "http://api.openweathermap.org/data/2.5/"
  # Your personal user token from openweathermap.org
  app_id = "xxxxxxxxxxxxxxxxxxxxxxx"
  # List of city identifiers
  cities = ["2988507", "519188"]
  # HTTP response timeout (default: 5s)
  response_timeout = "5s"
```

### Measurements & Fields:

- weather
  - humidity
  - temperature
  - pressure
  - rain
  - wind.speed
  - wind.deg

### Tags:

- weather
  - server
  - port
  - base_url
  - city_id

### Example Output:

Using this configuration:
```
[[inputs.openweathermap]]
  base_url = "http://api.openweathermap.org/data/2.5/"
  app_id = "change_this_with_your_appid"
  cities = ["2988507", "519188"]
  response_timeout = "5s"
```

When run with:
```
./telegraf -config telegraf.conf -input-filter openweathermap -test
```

It produces data similar to:
```
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=85i,pressure=1015,rain=0,temperature=-4.399999999999977,wind.deg=130,wind.speed=2 1544214600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=81i,pressure=1013,rain=0,temperature=6.6299999999999955,wind.deg=250,wind.speed=6.2 1544212800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=85i,pressure=1009.95,rain=0,temperature=-5.310000000000002,wind.deg=1.50009,wind.speed=1.3 1544216400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=81i,pressure=1009.6,rain=0,temperature=-7.139999999999986,wind.deg=327.5,wind.speed=0.42 1544227200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=86i,pressure=1009.32,rain=0,temperature=-7.8700000000000045,wind.deg=264.5,wind.speed=0.86 1544238000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=89i,pressure=1009.25,rain=0,temperature=-8.21999999999997,wind.deg=71.5003,wind.speed=0.17 1544248800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=91i,pressure=1008.57,rain=0,temperature=-6.658999999999992,wind.deg=110.501,wind.speed=1.27 1544259600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=91i,pressure=1008.05,rain=0,temperature=-6.55499999999995,wind.deg=126.5,wind.speed=1.41 1544270400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=84i,pressure=1007.72,rain=0,temperature=-7.617999999999995,wind.deg=149.504,wind.speed=1.66 1544281200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=87i,pressure=1006.51,rain=0,temperature=-7.324999999999989,wind.deg=151.501,wind.speed=2.41 1544292000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=90i,pressure=1005.78,rain=0,temperature=-7.103999999999985,wind.deg=170.506,wind.speed=3.21 1544302800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=89i,pressure=1004.9,rain=0,temperature=-8.407999999999959,wind.deg=167,wind.speed=3.47 1544313600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=89i,pressure=1003.49,rain=0,temperature=-9.12299999999999,wind.deg=157.504,wind.speed=3.86 1544324400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=93i,pressure=1002.54,rain=0,temperature=-7.638999999999953,wind.deg=156.501,wind.speed=3.99 1544335200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=96i,pressure=1001.59,rain=0,temperature=-5.065999999999974,wind.deg=158.006,wind.speed=4.09 1544346000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=94i,pressure=1000.56,rain=0,temperature=-3.4350000000000023,wind.deg=154,wind.speed=3.81 1544356800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=93i,pressure=999.59,rain=0,temperature=-2.70799999999997,wind.deg=151.501,wind.speed=3.71 1544367600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=998.53,rain=0,temperature=-1.6970000000000027,wind.deg=159,wind.speed=3.11 1544378400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=94i,pressure=998.21,rain=0,temperature=-0.7939999999999827,wind.deg=182.504,wind.speed=2.56 1544389200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=97i,pressure=998.73,rain=0,temperature=-0.39299999999997226,wind.deg=210.503,wind.speed=2.45 1544400000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=998.96,rain=0,temperature=-0.1379999999999768,wind.deg=211.503,wind.speed=2.83 1544410800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=98i,pressure=999.72,rain=0,temperature=-0.19299999999998363,wind.deg=197.001,wind.speed=3.12 1544421600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=99i,pressure=1000.21,rain=0,temperature=0.382000000000005,wind.deg=184,wind.speed=3.41 1544432400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=96i,pressure=1000.77,rain=0,temperature=0.492999999999995,wind.deg=179.502,wind.speed=3.67 1544443200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=97i,pressure=1001.33,rain=0,temperature=-0.4590000000000032,wind.deg=178.504,wind.speed=3.62 1544454000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=97i,pressure=1001.75,rain=0,temperature=-0.8369999999999891,wind.deg=177.504,wind.speed=3.3 1544464800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=97i,pressure=1002.17,rain=0,temperature=-0.7069999999999936,wind.deg=167.001,wind.speed=2.86 1544475600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=97i,pressure=1002.29,rain=0,temperature=-0.39400000000000546,wind.deg=152.003,wind.speed=2.41 1544486400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=98i,pressure=1002.04,rain=0,temperature=-0.18099999999998317,wind.deg=134.502,wind.speed=2.31 1544497200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=97i,pressure=1002.09,rain=0,temperature=0.05600000000004002,wind.deg=136.501,wind.speed=2.72 1544508000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=98i,pressure=1002.56,rain=0,temperature=0.5660000000000309,wind.deg=146.001,wind.speed=2.82 1544518800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=98i,pressure=1002.84,rain=0,temperature=0.6129999999999995,wind.deg=136.002,wind.speed=2.32 1544529600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=100i,pressure=1003.12,rain=0,temperature=0.36299999999999955,wind.deg=129,wind.speed=2.66 1544540400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=99i,pressure=1002.9,rain=0,temperature=0.4159999999999968,wind.deg=120.502,wind.speed=2.76 1544551200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=98i,pressure=1002.32,rain=0,temperature=0.19400000000001683,wind.deg=116.003,wind.speed=3.01 1544562000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=97i,pressure=1001.47,rain=0,temperature=-0.7029999999999745,wind.deg=102.003,wind.speed=3.3 1544572800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=1000.08,rain=0,temperature=-0.9569999999999936,wind.deg=90.0075,wind.speed=3.77 1544583600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=94i,pressure=999.32,rain=0,temperature=-1.238999999999976,wind.deg=94.002,wind.speed=3.71 1544594400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=97i,pressure=998.63,rain=0,temperature=-0.7459999999999809,wind.deg=99.5016,wind.speed=3.76 1544605200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=96i,pressure=997.62,rain=0,temperature=0.08100000000001728,wind.deg=100.508,wind.speed=3.67 1544616000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=96i,pressure=997.71,rain=0,temperature=0.4159999999999968,wind.deg=111.502,wind.speed=3.87 1544626800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=519188,host=localhost,port=80,server=api.openweathermap.org humidity=98i,pressure=997.89,rain=0,temperature=0.5520000000000209,wind.deg=123.008,wind.speed=3.55 1544637600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=92i,pressure=1015.86,rain=0,temperature=6.480000000000018,wind.deg=263.5,wind.speed=7.4 1544216400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=92i,pressure=1016.49,rain=0.035,temperature=6.3799999999999955,wind.deg=250,wind.speed=8.82 1544227200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=92i,pressure=1015.72,rain=0.015000000000001,temperature=7.189999999999998,wind.deg=252,wind.speed=9.51 1544238000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=93i,pressure=1015.65,rain=0.15,temperature=8.07000000000005,wind.deg=255.5,wind.speed=9.77 1544248800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=91i,pressure=1016.68,rain=0.14,temperature=9.01600000000002,wind.deg=257.501,wind.speed=9.27 1544259600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=1016.89,rain=0.47,temperature=9.69500000000005,wind.deg=255,wind.speed=8.51 1544270400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=92i,pressure=1015.58,rain=0.295,temperature=10.182000000000016,wind.deg=247.504,wind.speed=8.81 1544281200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=94i,pressure=1012.59,rain=0.98,temperature=10.425000000000011,wind.deg=235.001,wind.speed=10.26 1544292000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=1009.11,rain=0.9,temperature=11.271000000000015,wind.deg=241.006,wind.speed=11.21 1544302800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=93i,pressure=1008.06,rain=0.51,temperature=10.342000000000041,wind.deg=252,wind.speed=11.02 1544313600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=96i,pressure=1009.08,rain=1.35,temperature=9.527000000000044,wind.deg=267.504,wind.speed=10.04 1544324400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=92i,pressure=1009.11,rain=0.29,temperature=9.236000000000047,wind.deg=260.001,wind.speed=11.04 1544335200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=1009.86,rain=0.67,temperature=9.384000000000015,wind.deg=270.006,wind.speed=11.26 1544346000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=97i,pressure=1011.42,rain=1.585,temperature=9.140000000000043,wind.deg=279,wind.speed=10.56 1544356800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=1014.67,rain=1.03,temperature=9.04200000000003,wind.deg=293.001,wind.speed=9.71 1544367600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=93i,pressure=1019.05,rain=0.040000000000001,temperature=9.27800000000002,wind.deg=304,wind.speed=7.01 1544378400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=93i,pressure=1022.36,rain=0.014999999999999,temperature=8.706000000000017,wind.deg=303.504,wind.speed=5.56 1544389200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=94i,pressure=1025.07,rain=0.055000000000001,temperature=8.307000000000016,wind.deg=297.003,wind.speed=5.05 1544400000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=94i,pressure=1026.43,rain=0.039999999999999,temperature=8.087000000000046,wind.deg=287.003,wind.speed=4.28 1544410800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=96i,pressure=1027.27,rain=0.43,temperature=7.732000000000028,wind.deg=272.001,wind.speed=3.57 1544421600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=98i,pressure=1028.66,rain=0.91,temperature=8.357000000000028,wind.deg=282.5,wind.speed=4.36 1544432400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=96i,pressure=1029.3,rain=0.14,temperature=9.19300000000004,wind.deg=281.002,wind.speed=4.92 1544443200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=93i,pressure=1028.81,rain=0.030000000000001,temperature=8.991000000000042,wind.deg=273.004,wind.speed=4.12 1544454000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=91i,pressure=1028.34,rain=0.010000000000002,temperature=8.113,wind.deg=248.004,wind.speed=4.05 1544464800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=91i,pressure=1027.62,rain=0,temperature=7.617999999999995,wind.deg=241.501,wind.speed=4.81 1544475600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=90i,pressure=1026.53,rain=0,temperature=7.506000000000029,wind.deg=245.003,wind.speed=5.21 1544486400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=1025.7,rain=0,temperature=6.944000000000017,wind.deg=249.502,wind.speed=4.66 1544497200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=1024.87,rain=0.02,temperature=7.156000000000006,wind.deg=246.001,wind.speed=4.12 1544508000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=95i,pressure=1024.69,rain=0,temperature=6.716000000000008,wind.deg=231.501,wind.speed=3.97 1544518800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=94i,pressure=1023.76,rain=0,temperature=9.438000000000045,wind.deg=222.002,wind.speed=3.77 1544529600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=87i,pressure=1021.84,rain=0,temperature=9.863,wind.deg=197,wind.speed=3.11 1544540400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=92i,pressure=1019.92,rain=0,temperature=7.040999999999997,wind.deg=173.502,wind.speed=3.66 1544551200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=91i,pressure=1018.12,rain=0,temperature=6.16900000000004,wind.deg=167.503,wind.speed=3.96 1544562000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=89i,pressure=1016.55,rain=0,temperature=5.9220000000000255,wind.deg=187.003,wind.speed=4.3 1544572800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=92i,pressure=1015.4,rain=0,temperature=6.543000000000006,wind.deg=212.507,wind.speed=4.72 1544583600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=96i,pressure=1015.69,rain=0.23,temperature=8.336000000000013,wind.deg=241.002,wind.speed=6.11 1544594400000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=89i,pressure=1017.36,rain=0.059999999999999,temperature=9.204000000000008,wind.deg=250.502,wind.speed=5.56 1544605200000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=92i,pressure=1017.15,rain=0.24,temperature=10.05600000000004,wind.deg=247.008,wind.speed=6.12 1544616000000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=94i,pressure=1018.14,rain=0.51,temperature=9.866000000000042,wind.deg=275.502,wind.speed=4.17 1544626800000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=2988507,host=localhost,port=80,server=api.openweathermap.org humidity=92i,pressure=1019.21,rain=0.030000000000001,temperature=8.701999999999998,wind.deg=269.008,wind.speed=3.2 1544637600000000000
> weather,base_url=http://api.openweathermap.org/data/2.5/,city_id=0,host=localhost,port=80,server=api.openweathermap.org humidity=0i,pressure=0,rain=0,temperature=-273.15,wind.deg=0,wind.speed=0 0
```





