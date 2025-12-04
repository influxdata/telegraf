# Wavefront Output Plugin

This plugin writes metrics to a [Wavefront][wavefront] instance or a Wavefront
Proxy instance over HTTP or HTTPS.

⭐ Telegraf v1.5.0
🏷️ applications, cloud
💻 all

[wavefront]: https://www.wavefront.com

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `token` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
[[outputs.wavefront]]
  ## URL for Wavefront API or Wavefront proxy instance
  ## Direct Ingestion via Wavefront API requires authentication. See below.
  url = "https://metrics.wavefront.com"

  ## Maximum number of metrics to send per HTTP request. This value should be
  ## higher than the `metric_batch_size`. Values higher than 40,000 are not
  ## recommended.
  # http_maximum_batch_size = 10000

  ## Prefix for metrics keys
  # prefix = "my.specific.prefix."

  ## Use "value" for name of simple fields
  # simple_fields = false

  ## character to use between metric and field name
  # metric_separator = "."

  ## Convert metric name paths to use metricSeparator character
  ## When true will convert all _ (underscore) characters in final metric name.
  # convert_paths = true

  ## Use Strict rules to sanitize metric and tag names from invalid characters
  ## When enabled forward slash (/) and comma (,) will be accepted
  # use_strict = false

  ## Use Regex to sanitize metric and tag names from invalid characters
  ## Regex is more thorough, but significantly slower.
  # use_regex = false

  ## Tags to use as the source name for Wavefront ("host" if none is found)
  # source_override = ["hostname", "address", "agent_host", "node_host"]

  ## Convert boolean values to numeric values, with false -> 0.0 and true -> 1.0
  # convert_bool = true

  ## Truncate metric tags to a total of 254 characters for the tag name value
  ## Wavefront will reject any data point exceeding this limit if not truncated
  ## Defaults to 'false' to provide backwards compatibility.
  # truncate_tags = false

  ## Flush the internal buffers after each batch. This effectively bypasses the
  ## background sending of metrics normally done by the Wavefront SDK. This can
  ## be used if you are experiencing buffer overruns. The sending of metrics
  ## will block for a longer time, but this will be handled gracefully by
  ## internal buffering in Telegraf.
  # immediate_flush = true

  ## Send internal metrics (starting with `~sdk.go`) for valid, invalid, and
  ## dropped metrics
  # send_internal_metrics = true

  ## Optional TLS Config
  ## Set to true/false to enforce TLS being enabled/disabled. If not set,
  ## enable TLS only if any of the other options are specified.
  # tls_enable =
  ## Trusted root certificates for server
  # tls_ca = "/path/to/cafile"
  ## Used for TLS client certificate authentication
  # tls_cert = "/path/to/certfile"
  ## Used for TLS client certificate authentication
  # tls_key = "/path/to/keyfile"
  ## Send the specified TLS server name via SNI
  # tls_server_name = "kubernetes.example.com"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Timeout
  # timeout="10s"

  ## MaxIdleConns controls the maximum number of idle (keep-alive) connections
  ## across all hosts. Zero means unlimited.
  # max_idle_conn = 0

  ## MaxIdleConnsPerHost, if non-zero, controls the maximum idle (keep-alive)
  ## connections to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used.
  # max_idle_conn_per_host = 2

  ## Idle (keep-alive) connection timeout
  # idle_conn_timeout = 0

  ## Authentication for Direct Ingestion.
  ## Direct Ingestion requires one of: `token`,`auth_csp_api_token`, or
  ## `auth_csp_client_credentials` (see https://docs.wavefront.com/csp_getting_started.html)
  ## to learn more about using CSP credentials with Wavefront.
  ## Not required if using a Wavefront proxy.

  ## Wavefront API Token Authentication, ignored if using a Wavefront proxy
  ## 1. Click the gear icon at the top right in the Wavefront UI.
  ## 2. Click your account name (usually your email)
  ## 3. Click *API access*.
  # token = "YOUR_TOKEN"

  ## Base URL used for authentication, ignored if using a Wavefront proxy or a
  ## Wavefront API token.
  # auth_csp_base_url=https://console.cloud.vmware.com

  ## CSP API Token Authentication, ignored if using a Wavefront proxy
  # auth_csp_api_token=CSP_API_TOKEN_HERE

  ## CSP Client Credentials Authentication Information, ignored if using a
  ## Wavefront proxy.
  ## See also: https://docs.wavefront.com/csp_getting_started.html#whats-a-server-to-server-app
  # [outputs.wavefront.auth_csp_client_credentials]
  #  app_id=CSP_APP_ID_HERE
  #  app_secret=CSP_APP_SECRET_HERE
  #  org_id=CSP_ORG_ID_HERE
```

### Convert Path & Metric Separator

If the `convert_path` option is true any `_` in metric and field names will be
converted to the `metric_separator` value.  By default, to ease metrics browsing
in the Wavefront UI, the `convert_path` option is true, and `metric_separator`
is `.` (dot).  Default integrations within Wavefront expect these values to be
set to their defaults, however if converting from another platform it may be
desirable to change these defaults.

### Use Regex

Most illegal characters in the metric name are automatically converted to `-`.
The `use_regex` setting can be used to ensure all illegal characters are
properly handled, but can lead to performance degradation.

### Source Override

Often when collecting metrics from another system, you want to use the target
system as the source, not the one running Telegraf.  Many Telegraf plugins will
identify the target source with a tag. The tag name can vary for different
plugins. The `source_override` option will use the value specified in any of the
listed tags if found. The tag names are checked in the same order as listed, and
if found, the other tags will not be checked. If no tags specified are found,
the default host tag will be used to identify the source of the metric.

### Wavefront Data format

The expected input for Wavefront is specified in the following way:

```text
<metric> <value> [<timestamp>] <source|host>=<sourceTagValue> [tagk1=tagv1 ...tagkN=tagvN]
```

More information about the Wavefront data format is available in the
[documentation][docs].

[docs]: https://community.wavefront.com/docs/DOC-1031

### Allowed values for metrics

Wavefront allows `integers` and `floats` as input values.  By default it also
maps `bool` values to numeric, false -> 0.0, true -> 1.0.  To map `strings` use
the [enum](../../processors/enum) processor plugin.
