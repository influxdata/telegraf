# HTTP listener service input plugin

The HTTP listener is a service input plugin that listens for messages sent via HTTP POST.
The plugin expects messages in the InfluxDB line-protocol ONLY, other Telegraf input data formats are not supported.
The intent of the plugin is to allow Telegraf to serve as a proxy/router for the `/write` endpoint of the InfluxDB HTTP API.

The `/write` endpoint supports the `precision` query parameter and can be set to one of `ns`, `u`, `ms`, `s`, `m`, `h`.  All other parameters are ignored and defer to the output plugins configuration.

When chaining Telegraf instances using this plugin, CREATE DATABASE requests receive a 200 OK response with message body `{"results":[]}` but they are not relayed. The output configuration of the Telegraf instance which ultimately submits data to InfluxDB determines the destination database.

See: [Telegraf Input Data Formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md#influx).

**Example:**
```
curl -i -XPOST 'http://localhost:8186/write' --data-binary 'cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000'
```

### Configuration:

This is a sample configuration for the plugin.

```toml
# # Influx HTTP write listener
[[inputs.http_listener]]
  ## Address and port to host HTTP listener on
  service_address = ":8186"

  ## timeouts
  read_timeout = "10s"
  write_timeout = "10s"
```
