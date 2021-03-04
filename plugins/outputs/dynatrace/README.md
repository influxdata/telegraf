# Dynatrace Output Plugin

This plugin is sending telegraf metrics to [Dynatrace](https://www.dynatrace.com). It has two operational modes.

Telegraf minimum version: Telegraf 1.16 
Plugin minimum tested version: 1.16

## Running alongside Dynatrace OneAgent

if you run the Telegraf agent on a host or VM that is monitored by the Dynatrace OneAgent then you only need to enable the plugin but need no further configuration. The Dynatrace telegraf output plugin will send all metrics to the OneAgent which will use its secure and load balanced connection to send the metrics to your Dynatrace SaaS or Managed environment.

## Running standalone

If you run the Telegraf agent on a host or VM without a OneAgent you will need to configure the environment API endpoint to send the metrics to and an API token for security.

The endpoint for the Dynatrace Metrics API is 

* Managed https://{your-domain}/e/{your-environment-id}/api/v2/metrics/ingest
* SaaS https://{your-environment-id}.live.dynatrace.com/api/v2/metrics/ingest

You can learn more about how to use the Dynatrace API [here](https://www.dynatrace.com/support/help/dynatrace-api/)

You will also need to configure an API token for secure access. Find out how to create a token [here](https://www.dynatrace.com/support/help/dynatrace-api/environment-api/tokens/) or simply navigate to **Settings > Integration > Dynatrace API** in your Dynatrace environment and create a token with Dynatrace API and create a new token with 
'Ingest metrics data points' access scope enabled. 

## Configuration

```toml
[[outputs.dynatrace]]
  ## Leave empty or use the local ingest endpoint of your OneAgent monitored host (e.g.: http://127.0.0.1:14499/metrics/ingest).
  ## Set Dynatrace environment URL (e.g.: https://YOUR_DOMAIN/api/v2/metrics/ingest) if you do not use a OneAgent
  url = ""
  api_token = ""
  ## Optional prefix for metric names (e.g.: "telegraf.")
  prefix = "telegraf."
  ## Flag for skipping the tls certificate check, just for testing purposes, should be false by default
  insecure_skip_verify = false
  ## If you want to convert values represented as gauges to counters, add the metric names here
  additional_counters = [ ]

```

## Requirements

You will either need a Dynatrace OneAgent (version 1.201 or higher) installed on the same host as Telegraf; or a Dynatrace environment with version 1.202 or higher. Monotonic counters (e.g. diskio.reads, system.uptime) require release 208 or later.
You will either need a Dynatrace OneAgent (version 1.201 or higher) installed on the same host as Telegraf; or a Dynatrace environment with version 1.202 or higher  

## Limitations
Telegraf measurements which can't be converted to a float64 are skipped.
