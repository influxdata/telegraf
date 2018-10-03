# Telegraf Input Plugin: Traefik

This plugin gather health check status from services registered in Traefik. Traefik must be started with the `--api` option (formerly `--web`). See the [configuration docs](https://docs.traefik.io/configuration/api/) for more information.

### Configuration:

Simple configuration, with one measurement per request capture:

```toml
# Description
[[inputs.traefik]]
  ## Required Traefik server address, host and port (default: "127.0.0.1")
  # address = "http://127.0.0.1:8080"
  #
  # [inputs.traefik.tags]
  #  instance = "prod"
```

Configuration with additional measurements for each status code encountered:

```toml
# Description
[[inputs.traefik]]
  ## Required Traefik server address, host and port (default: "127.0.0.1")
  # address = "http://127.0.0.1:8080"
  # include_status_code_measurement = true
  #
  # [inputs.traefik.tags]
  #  instance = "prod"
```

### Measurements & Fields:

- `traefik` measurement have the following fields:
    - **total_count** (int) - total number of responses since the application was last restarted
    - **average_response_time_sec** (float64) - average time (in seconds) of all responses
    - **total_response_time_sec** (float64) - total time (in seconds) of all responses
    - **unixtime** (int) - unix timestamp of when this service was last restarted
    - **uptime_sec** (int) - elapsed time (in seconds) since this service was restarted
    - **health_response_time_sec** (float64) - round trip time to gather this set of metrics
    - Http Codes - each status code count since this service was last restarted
	- **status_code_200** (int)
	- **status_code_400** (int)
	- **status_code_500** (int)
	- *status_code_*...

- `traefik_status_codes` measurement have the following fields:
    - **total_count** (int) - total number of responses since the application was last restarted
    - **unixtime** (int) - unix timestamp of when this service was last restarted
    - **uptime_sec** (int) - elapsed time (in seconds) since this service was restarted
    - **health_response_time_sec** (float64) - round trip time to gather this set of metrics
    - **count** (int) - count for this status code since the last time this service was restarted


### Tags:

- All measurements have the following tags:
    - **source**

- `traefik_status_codes` measurement have the following additional tags:
    - **status_code**

### Sample Queries:

```
// Count number of requests with response code 200
SELECT COUNT("status_code_200") FROM "traefik" WHERE time > now() - 24h 

// Select Status Code Counts PER Hour for the past 24 hours
SELECT difference(last("count")) FROM "traefik_status_codes" WHERE time > now() - 24h GROUP BY time(1h), "status_code" fill(null)

```

### Example Output:

```
./telegraf config > telegraf.conf

# edit telegraf.conf

./telegraf -config telegraf.conf -input-filter traefik -test
	
> traefik,source=http://localhost:8080,instance=prod,host=My-MacBook-Pro.local status_code_302=27117i,status_code_307=4i,status_code_400=1011i,status_code_404=812i,status_code_401=1i,total_response_time_sec=83381.956659941,total_count=1222527i,uptime_sec=1387245.424820805,unixtime=1518039391i,status_code_500=34i,status_code_504=1206i,status_code_503=1i,status_code_301=3i,average_response_time_sec=0.068204593,status_code_502=715i,status_code_304=559007i,status_code_200=632616i,health_response_time_sec=0.061906656 1518039391000000000

> traefik_status_codes,source=http://localhost:8080,status_code=304,instance=prod total_count=1588167i,uptime_sec=1629993.771782303,unixtime=1518624416i,count=752109i,health_response_time_sec=1.0587943 1518624424000000000

```
