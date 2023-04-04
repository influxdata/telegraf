# HTTP JSON Input Plugin

**DEPRECATED in Telegraf v1.6: Use [HTTP input plugin][] as replacement**

The httpjson plugin collects data from HTTP URLs which respond with JSON.  It
flattens the JSON and finds all numeric values, treating them as floats.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read flattened metrics from one or more JSON HTTP endpoints
[[inputs.httpjson]]
  ## NOTE This plugin only reads numerical measurements, strings and booleans
  ## will be ignored.

  ## Name for the service being polled.  Will be appended to the name of the
  ## measurement e.g. "httpjson_webserver_stats".
  ##
  ## Deprecated (1.3.0): Use name_override, name_suffix, name_prefix instead.
  name = "webserver_stats"

  ## URL of each server in the service's cluster
  servers = [
    "http://localhost:9999/stats/",
    "http://localhost:9998/stats/",
  ]
  ## Set response_timeout (default 5 seconds)
  response_timeout = "5s"

  ## HTTP method to use: GET or POST (case-sensitive)
  method = "GET"

  ## Tags to extract from top-level of JSON server response.
  # tag_keys = [
  #   "my_tag_1",
  #   "my_tag_2"
  # ]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Request Parameters (all values must be strings).  For "GET" requests, data
  ## will be included in the query.  For "POST" requests, data will be included
  ## in the request body as "x-www-form-urlencoded".
  # [inputs.httpjson.parameters]
  #   event_type = "cpu_spike"
  #   threshold = "0.75"

  ## HTTP Request Headers (all values must be strings).
  # [inputs.httpjson.headers]
  #   X-Auth-Token = "my-xauth-token"
  #   apiVersion = "v1"
```

## Metrics

### Measurements & Fields

- httpjson
  - response_time (float): Response time in seconds

Additional fields are dependant on the response of the remote service being
polled.

### Tags

- All measurements have the following tags:
  - server: HTTP origin as defined in configuration as `servers`.

Any top level keys listed under `tag_keys` in the configuration are added as
tags.  Top level keys are defined as keys in the root level of the object in a
single object response, or in the root level of each object within an array of
objects.

## Example Output

This plugin understands responses containing a single JSON object, or a JSON
Array of Objects.

**Object Output:**

Given the following response body:

```json
{
    "a": 0.5,
    "b": {
        "c": "some text",
        "d": 0.1,
        "e": 5
    },
    "service": "service01"
}
```

The following metric is produced:

```text
httpjson,server=http://localhost:9999/stats/ b_d=0.1,a=0.5,b_e=5,response_time=0.001
```

Note that only numerical values are extracted and the type is float.

If `tag_keys` is included in the configuration:

```toml
[[inputs.httpjson]]
  tag_keys = ["service"]
```

Then the `service` tag will also be added:

```text
httpjson,server=http://localhost:9999/stats/,service=service01 b_d=0.1,a=0.5,b_e=5,response_time=0.001
```

**Array Output:**

If the service returns an array of objects, one metric is be created for each
object:

```json
[
    {
        "service": "service01",
        "a": 0.5,
        "b": {
            "c": "some text",
            "d": 0.1,
            "e": 5
        }
    },
    {
        "service": "service02",
        "a": 0.6,
        "b": {
            "c": "some text",
            "d": 0.2,
            "e": 6
        }
    }
]
```

```text
httpjson,server=http://localhost:9999/stats/,service=service01 a=0.5,b_d=0.1,b_e=5,response_time=0.003
httpjson,server=http://localhost:9999/stats/,service=service02 a=0.6,b_d=0.2,b_e=6,response_time=0.003
```

[HTTP input plugin]: ../http/README.md
