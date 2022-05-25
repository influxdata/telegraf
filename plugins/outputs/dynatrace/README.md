# Dynatrace Output Plugin

This plugin sends Telegraf metrics to [Dynatrace](https://www.dynatrace.com) via
the [Dynatrace Metrics API V2][api-v2]. It may be run alongside the Dynatrace
OneAgent for automatic authentication or it may be run standalone on a host
without a OneAgent by specifying a URL and API Token.  More information on the
plugin can be found in the [Dynatrace documentation][docs].  All metrics are
reported as gauges, unless they are specified to be delta counters using the
`additional_counters` config option (see below).  See the [Dynatrace Metrics
ingestion protocol documentation][proto-docs] for details on the types defined
there.

[api-v2]: https://www.dynatrace.com/support/help/dynatrace-api/environment-api/metric-v2/

[docs]: https://www.dynatrace.com/support/help/how-to-use-dynatrace/metrics/metric-ingestion/ingestion-methods/telegraf/

[proto-docs]: https://www.dynatrace.com/support/help/how-to-use-dynatrace/metrics/metric-ingestion/metric-ingestion-protocol

## Requirements

You will either need a Dynatrace OneAgent (version 1.201 or higher) installed on
the same host as Telegraf; or a Dynatrace environment with version 1.202 or
higher.

- Telegraf minimum version: Telegraf 1.16

## Getting Started

Setting up Telegraf is explained in the [Telegraf
Documentation][getting-started].
The Dynatrace exporter may be enabled by adding an `[[outputs.dynatrace]]`
section to your `telegraf.conf` config file.  All configurations are optional,
but if a `url` other than the OneAgent metric ingestion endpoint is specified
then an `api_token` is required.  To see all available options, see
[Configuration](#configuration) below.

[getting-started]: https://docs.influxdata.com/telegraf/latest/introduction/getting-started/

### Running alongside Dynatrace OneAgent (preferred)

If you run the Telegraf agent on a host or VM that is monitored by the Dynatrace
OneAgent then you only need to enable the plugin, but need no further
configuration. The Dynatrace Telegraf output plugin will send all metrics to the
OneAgent which will use its secure and load balanced connection to send the
metrics to your Dynatrace SaaS or Managed environment.  Depending on your
environment, you might have to enable metrics ingestion on the OneAgent first as
described in the [Dynatrace documentation][docs].

Note: The name and identifier of the host running Telegraf will be added as a
dimension to every metric. If this is undesirable, then the output plugin may be
used in standalone mode using the directions below.

```toml
[[outputs.dynatrace]]
  ## No options are required. By default, metrics will be exported via the OneAgent on the local host.
```

### Running standalone

If you run the Telegraf agent on a host or VM without a OneAgent you will need
to configure the environment API endpoint to send the metrics to and an API
token for security.

You will also need to configure an API token for secure access. Find out how to
create a token in the [Dynatrace documentation][api-auth] or simply navigate to
**Settings > Integration > Dynatrace API** in your Dynatrace environment and
create a token with Dynatrace API and create a new token with 'Ingest metrics'
(`metrics.ingest`) scope enabled. It is recommended to limit Token scope to only
this permission.

The endpoint for the Dynatrace Metrics API v2 is

- on Dynatrace Managed:
  `https://{your-domain}/e/{your-environment-id}/api/v2/metrics/ingest`
- on Dynatrace SaaS:
  `https://{your-environment-id}.live.dynatrace.com/api/v2/metrics/ingest`

```toml
[[outputs.dynatrace]]
  ## If no OneAgent is running on the host, url and api_token need to be set

  ## Dynatrace Metrics Ingest v2 endpoint to receive metrics
  url = "https://{your-environment-id}.live.dynatrace.com/api/v2/metrics/ingest"

  ## API token is required if a URL is specified and should be restricted to the 'Ingest metrics' scope
  api_token = "your API token here" // hard-coded for illustration only, should be read from environment
```

You can learn more about how to use the Dynatrace API
[here](https://www.dynatrace.com/support/help/dynatrace-api/).

[api-auth]: https://www.dynatrace.com/support/help/dynatrace-api/basics/dynatrace-api-authentication/

## Configuration

```toml
# Send telegraf metrics to a Dynatrace environment
[[outputs.dynatrace]]
  ## For usage with the Dynatrace OneAgent you can omit any configuration,
  ## the only requirement is that the OneAgent is running on the same host.
  ## Only setup environment url and token if you want to monitor a Host without the OneAgent present.
  ##
  ## Your Dynatrace environment URL.
  ## For Dynatrace OneAgent you can leave this empty or set it to "http://127.0.0.1:14499/metrics/ingest" (default)
  ## For Dynatrace SaaS environments the URL scheme is "https://{your-environment-id}.live.dynatrace.com/api/v2/metrics/ingest"
  ## For Dynatrace Managed environments the URL scheme is "https://{your-domain}/e/{your-environment-id}/api/v2/metrics/ingest"
  url = ""

  ## Your Dynatrace API token.
  ## Create an API token within your Dynatrace environment, by navigating to Settings > Integration > Dynatrace API
  ## The API token needs data ingest scope permission. When using OneAgent, no API token is required.
  api_token = ""

  ## Optional prefix for metric names (e.g.: "telegraf")
  prefix = "telegraf"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Optional flag for ignoring tls certificate check
  # insecure_skip_verify = false

  ## Connection timeout, defaults to "5s" if not set.
  timeout = "5s"

  ## If you want metrics to be treated and reported as delta counters, add the metric names here
  additional_counters = [ ]

  ## Optional dimensions to be added to every metric
  # [outputs.dynatrace.default_dimensions]
  # default_key = "default value"
```

### `url`

*required*: `false`

*default*: Local OneAgent endpoint

Set your Dynatrace environment URL (e.g.:
`https://{your-environment-id}.live.dynatrace.com/api/v2/metrics/ingest`, see
the [Dynatrace documentation][post-ingest] for details) if you do not use a
OneAgent or wish to export metrics directly to a Dynatrace metrics v2
endpoint. If a URL is set to anything other than the local OneAgent endpoint,
then an API token is required.

```toml
url = "https://{your-environment-id}.live.dynatrace.com/api/v2/metrics/ingest"
```

[post-ingest]: https://www.dynatrace.com/support/help/dynatrace-api/environment-api/metric-v2/post-ingest-metrics/

### `api_token`

*required*: `false` unless `url` is specified

API token is required if a URL other than the OneAgent endpoint is specified and
it should be restricted to the 'Ingest metrics' scope.

```toml
api_token = "your API token here"
```

### `prefix`

*required*: `false`

Optional prefix to be prepended to all metric names (will be separated with a
`.`).

```toml
prefix = "telegraf"
```

### `insecure_skip_verify`

*required*: `false`

Setting this option to true skips TLS verification for testing or when using
self-signed certificates.

```toml
insecure_skip_verify = false
```

### `additional_counters`

*required*: `false`

If you want a metric to be treated and reported as a delta counter, add its name
to this list.

```toml
additional_counters = [ ]
```

### `default_dimensions`

*required*: `false`

Default dimensions that will be added to every exported metric.

```toml
[outputs.dynatrace.default_dimensions]
default_key = "default value"
```

## Limitations

Telegraf measurements which can't be converted to a number are skipped.
