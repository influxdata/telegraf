# Icinga2 Input Plugin


The Icinga2 input plugin collects various information about a running Icinga2 process. It uses the icinga2 api  `/v1/stats` endpoint to gather metrics.

### Configuration:

```toml
[[inputs.icinga2]]
  ## URL for the kubelet
  url = "https://hostname:5665"

  ## Use bearer token for authorization
  # bearer_token = /path/to/bearer/token

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional SSL Config
  # ssl_ca = /path/to/cafile
  # ssl_cert = /path/to/certfile
  # ssl_key = /path/to/keyfile
  ## Use SSL but skip chain & host verification
  insecure_skip_verify = true

  ## Credentials for basic HTTP authentication.
  username = "root"
  password = "root"

```

### Summary Data

```json
        {
            "name": "CIB", 
            "perfdata": [], 
            "status": {
                "active_host_checks": 0.016666666666666666, 
                "active_host_checks_15min": 15.0, 
                "active_host_checks_1min": 1.0, 
                "active_host_checks_5min": 5.0, 
                "active_service_checks": 0.18333333333333332, 
                "active_service_checks_15min": 163.0, 
                "active_service_checks_1min": 11.0, 
                "active_service_checks_5min": 55.0, 
                "avg_execution_time": 1.3657611500133167, 
                "avg_latency": 0.0006881843913685192, 
                "max_execution_time": 10.007308959960938, 
                "max_latency": 0.0018439292907714844, 
                "min_execution_time": 0.0004010200500488281, 
                "min_latency": 0.00016307830810546875, 
                "num_hosts_acknowledged": 0.0, 
                "num_hosts_down": 0.0, 
                "num_hosts_flapping": 0.0, 
                "num_hosts_in_downtime": 0.0, 
                "num_hosts_pending": 0.0, 
                "num_hosts_unreachable": 0.0, 
                "num_hosts_up": 1.0, 
                "num_services_acknowledged": 0.0, 
                "num_services_critical": 2.0, 
                "num_services_flapping": 0.0, 
                "num_services_in_downtime": 0.0, 
                "num_services_ok": 7.0, 
                "num_services_pending": 0.0, 
                "num_services_unknown": 0.0, 
                "num_services_unreachable": 0.0, 
                "num_services_warning": 2.0, 
                "passive_host_checks": 0.0, 
                "passive_host_checks_15min": 0.0, 
                "passive_host_checks_1min": 0.0, 
                "passive_host_checks_5min": 0.0, 
                "passive_service_checks": 0.0, 
                "passive_service_checks_15min": 0.0, 
                "passive_service_checks_1min": 0.0, 
                "passive_service_checks_5min": 0.0, 
                "uptime": 5090429.723046064
            }

```

