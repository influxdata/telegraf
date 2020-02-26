# Vespa Input Plugin

Collects metrics reported by the [Vespa metrics API](https://docs.vespa.ai/documentation/reference/metrics.html).

### Configuration
```toml
# Collects metrics reported by the Vespa metrics API.
[[inputs.vespa]]
  ## URL to Vespa metrics API.
  url = "http://localhost:19092/metrics/v2/values"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true
```

### Metrics
The complete list of default metrics can be found [here](https://docs.datadoghq.com/integrations/vespa/#data-collected).

### Example Output

```
vespa,host=vespa-container,httpMethod=GET,serviceId=container http.status.2xx.rate=0.0333756091049 1582710361000000000
vespa,host=vespa-container,serviceId=container mem.heap.free.average=1397378050.6666667,serverActiveThreads.average=0 1582710361000000000
vespa,host=vespa-container,serviceId=container memory_virt=3973877760,memory_rss=1674264576,cpu=11.1090154658619 1582710361000000000
```
