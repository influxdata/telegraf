# Dynatrace Output Plugin

This plugin writes telegraf metrics to a Dynatrace environment. 

A API token is necessary, which can be obtained in your Dynatrace environment. Navigate to **Dynatrace > Settings > Integration > Dynatrace API** and create a new token with 
'Data ingest' access scope enabled. 

Telegraf measurements which cant be converted to a float64 are skipped.

Metrics fields are added to the measurement name by using '.' in the metric name. 

### Configuration

```toml
[[outputs.dynatrace]]
  ## Dynatrace environment URL.
  environmentURL = ""
  environmentApiToken = ""
  skipCertificateCheck = false

```