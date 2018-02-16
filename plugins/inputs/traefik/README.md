# Telegraf Input Plugin: Traefik

This plugin gather health check status from services registered in Traefik.

### Configuration:

```toml
# Description
[[inputs.traefik]]
  ## Required Traefik server address (default: "127.0.0.1")
  server = "127.0.0.1"
  ## Required Traefik port (default "8080")
  port = 8080
  ## Required Traefik instance name (default: "default")
  instance = "default"
```

### Measurements & Fields:

- all measurements have the following fields:
    - total_count (int)
    - average_response_time_sec (float64)
    - total_response_time_sec (float64)
    - Http Codes
      - 200 (int)
      - 400 (int)
      - 500 (int)
      - ...

### Tags:

- All measurements have the following tags:
    - instance

### Sample Queries:

```
SELECT COUNT("200") FROM "traefik_healthchecks" WHERE time > now() - 24h // Count number of requests with response code 200
```

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter traefik -test
traefik_healthchecks,instance=prod-instance average_response_time_sec=0.001169439,total_count=13i,404=6i,200=7i,total_response_time_sec=0.015202713 1492169158000000000
```
