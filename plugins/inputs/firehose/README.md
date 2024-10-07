# AWS Data Firehose HTTP Listener Input Plugin

Firehose HTTP Listener is a service input plugin that listens for metrics sent
via HTTP from AWS Data Firehose. It strictly follows the request response
schema as describe in the official
[documentation](https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html).

## Service Input <!-- @/docs/includes/service_input.md -->

This plugin is a service input. Normal plugins gather metrics determined by the
interval setting. Service plugins start a service to listens and waits for
metrics or events to occur. Service plugins have two key differences from
normal plugins:

1. The global or plugin specific `interval` setting may not apply
2. The CLI options of `--test`, `--test-wait`, and `--once` may not produce
   output for this plugin

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# AWS Data Firehose HTTP listener
[[inputs.firehose]]
  ## Address and port to host HTTP listener on
  service_address = ":8080"

  ## Paths to listen to.
  # paths = ["/telegraf"]

  ## Save path as firehose_http_path tag if set to true
  # path_tag = false

  ## maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## maximum duration before timing out write of the response
  # write_timeout = "10s"

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Minimal TLS version accepted by the server
  # tls_min_version = "TLS12"

  ## Optional access key to accept for authentication.
  ## AWS Data Firehose uses "x-amz-firehose-access-key" header to set the access key
  # access_key = "foobar"
  
  ## Optional setting to add parameters as tags
  ## If the http header "x-amz-firehose-common-attributes" is not present on the request, no corresponding tag will be added
  ## The header value should be a json and should follow the schema as describe in the official documentation: https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html#requestformat
  # parameter_tags = ["env"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

## Metrics

Metrics are collected from the `records.[*].data` field in the request body.
The data must be base64 encoded and may be sent in any supported
[data format][data_format]. Metrics are parsed depending on the value of
`data_format`.

## Example Output

When run with this configuration:

```toml
[[inputs.firehose]]
  service_address = ":8080"
  paths = ["/telegraf"]
  path_tag = true
  data_format = "json_v2"
  [[inputs.firehose.json_v2]]
    timestamp_path = "timestamp"
    timestamp_format = "unix_ms"
    [[inputs.firehose.json_v2.tag]]
      path = "type"
    [[inputs.firehose.json_v2.field]]
      path = "message"
```

the following curl command:

```sh
curl -i -XPOST 'localhost:8080/telegraf' \
--header 'x-amz-firehose-request-id: ed4acda5-034f-9f42-bba1-f29aea6d7d8f' \
--header 'Content-Type: application/json' \
--data '{
    "requestId": "ed4acda5-034f-9f42-bba1-f29aea6d7d8f",
    "timestamp": 1578090901599,
    "records": [
        {
            "data": "eyJ0aW1lc3RhbXAiOjE3MjUwMDE4NTEwMDAsInR5cGUiOiJleGFtcGxlIiwibWVzc2FnZSI6ImhlbGxvIn0K"
        }
    ]
}'
```

produces:

```text
firehose,firehose_http_path=/telegraf,type=example message="hello" 1725001851000000000
```

[data_format]: /docs/DATA_FORMATS_INPUT.md
