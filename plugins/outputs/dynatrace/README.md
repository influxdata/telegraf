# Dynatrace Output Plugin

This plugin writes telegraf metrics to a Dynatrace environment. 

An API token is necessary, which can be obtained in your Dynatrace environment. Navigate to **Dynatrace > Settings > Integration > Dynatrace API** and create a new token with 
'Data ingest' access scope enabled. 

Telegraf measurements which can't be converted to a float64 are skipped.

Metrics fields are added to the measurement name by using '.' in the metric name. 

### Configuration

```toml
[[outputs.dynatrace]]
  ## Dynatrace environment URL (e.g.: https://YOUR_DOMAIN/api/v2/metrics/ingest) or use the local ingest endpoint of your OneAgent monitored host (e.g.: http://127.0.0.1:14499/metrics/ingest).
  environmentURL = ""
  environmentApiToken = ""
  ## Optional prefix for metric names (e.g.: "telegraf.")
  prefix = "telegraf."
  ## Flag for skipping the tls certificate check, just for testing purposes, should be false by default
  skipCertificateCheck = false

```