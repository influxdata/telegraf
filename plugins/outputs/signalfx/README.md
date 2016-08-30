# SignalFx Output Plugin

This plugin writes to [SignalFx](https://signalfx.com/) via HTTP.<br/>
For each Telegraf metric a SignalFx gauge datapoint is written per field.<br/>
The datapoint's metric name is a concatention of the Telegraf metric name and the field name.<br/>
Tags are written as datapoint dimensions.

### Configuration:

```toml
# Send Telegraf metrics to SignalFx
[[outputs.signalfx]]
  ## Your organization's SignalFx API access token.
  auth_token = "SuperSecretToken"

  ## Optional HTTP User Agent value; Overrides the default.
  # user_agent = "Telegraf collector"

  ## Optional SignalFX API endpoint value; Overrides the default.
  # endpoint = "https://ingest.signalfx.com/v2/datapoint"
```

### Required parameters:

* `auth_token`: Your organization's SignalFx API access token.


### Optional parameters:

* `user_agent`: HTTP User Agent.
* `endpoint`: SignalFX API endpoint.
