# Arista LANZ Consumer Input Plugin

This plugin provides a consumer for use with Arista Networks’ Latency Analyzer (LANZ)

Metrics are read from a stream of data  via TCP through port 50001 on the
switches management IP. The data is in Protobuffers format. You will need to
enable and configure the LANZ datastream.

- https://www.arista.com/en/um-eos/eos-section-43-3-configuring-lanz#ww1149292

For more information on Arista LANZ

- https://www.arista.com/en/um-eos/eos-latency-analyzer-lanz

This plugin uses Arista's sdk.

- https://github.com/aristanetworks/goarista

Example config

```
[global_tags]
[agent]
  interval = "1s"
  round_interval = true
  metric_batch_size = 1000
  metric_buffer_limit = 10000
  collection_jitter = "0s"
  flush_interval = "10s"
  flush_jitter = "0s"
  precision = ""
  debug = false
  quiet = false
  logfile = ""
  hostname = ""
  omit_hostname = false
[[outputs.elasticsearch]]
  urls = ["http://localhost:9200"]
  index_name = "lanz-%Y.%m.%d"
  manage_template = true
  template_name = "telegraf"
[[outputs.influxdb]]
  urls = ["http://localhost:8086"]
  database = "lanz"
  password = "lanz"
  precision = "ns"
  retention_policy = ""
  username = "lanz"
[[inputs.lanz_consumer]]
  servers = [
    "tcp://switch1.example.com:50001",
    "tcp://switch2.example.com:50001",
  ]
```
