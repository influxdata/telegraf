# Mosquitto Input Plugin

The [Mosquitto](http://mosquitto.org/) plugin reads from statistics topics 
and send measurements to InfluxDB.

Measurements are provided by mosquitto (see https://mosquitto.org/man/mosquitto-8.html for more)

### Configuration:

```toml
# Read metrics from MQTT topic(s)
[[inputs.mosquitto]]
  ## MQTT broker URLs to be used. The format should be scheme://host:port,
  ## schema can be tcp, ssl, or ws.
  servers = ["tcp://localhost:1883"]
  ## Connection timeout for initial connection in seconds
  connection_timeout = "30s"

  # If empty, a random client ID will be generated.
  client_id = ""

  ## username and password to connect MQTT server.
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional static tags to add
  # tags = [
  #  "mosquitto_instance_name: lx6109",
  # ]
```

### Tags:

- All measurements are tagged with the incoming topic and telegraf will add a host tag.
`topic=$SYS/broker/client/connected`
