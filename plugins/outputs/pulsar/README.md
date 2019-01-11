# Apache Pulsar Output Plugin

This plugin writes to Apache Pulsar. The sample configuration is:

```
  ## URL to Pulsar cluster
  ## If you use SSL, then the protocol should be "pulsar+ssl"
  url = "pulsar://localhost:6650"

  ## Timeout while trying to connect
  dial_timeout = "15s"

  ## Timeout while trying to send message
  send_timeout = "5s"

  ## Topic of message
  topic = ""

  ## Name of the producer
  name = ""

  ## Path to certificates and key for TLS
  # tls_ca = ""
  # tls_cert = ""
  # tls_key = ""

  ## Other optionals
  # ping_frequency = "1s"
  # ping_timeout = "1s"
  # initial_reconnect_delay = "3s"
  # max_reconnect_delay = "10s"
  # new_producer_timeout = "10s"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```
