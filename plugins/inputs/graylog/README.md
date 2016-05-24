# GrayLog plugin

The Graylog plugin can collect data from remote Graylog service URLs which respond. 


Sample configration
```
[[inputs.graylog]]
  ## NOTE This plugin only reads numerical measurements, strings and booleans
  ## will be ignored.

  ## a name for the service being polled
  name = "graylog_jvm"

  ## URL of each server in the service's cluster
  servers = [
    "http://10.224.162.16:12900/system/metrics/multiple"
  ]


  ## List of tag names to extract from top-level of JSON server response
  tag_keys = [
  ]

  metrics = [
    "jvm.cl.loaded",
    "jvm.memory.pools.Metaspace.committed"
  ]
  ## HTTP Header parameters (all values must be strings)
  [inputs.graylog.headers]
    Authorization = "Basic YWRtaW46YWRtaW4"
    Content-Type = "application/json"
    Accept = "application/json"


  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
```

List of metrics can be found on Graylog webservice documentation or by hitting the the web service api `http://[graylog-host]:12900/system/metrics`  
