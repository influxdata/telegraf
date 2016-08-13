# Pusher Output Plugin

This output plugin writes to the [Pusher REST API](https://pusher.com/docs/rest_api).

## Configuration

This plugin's configuration specifies a Pusher channel that incoming events should be published to.

The plugin will read Telegraf metric names and use those as the corresponding Pusher event names.

```
# Configuration for Pusher output.
[[outputs.pusher]]
  ## Pusher Credentials
  ## Pusher requires all three of app ID, key and secret for authentication.
  app_id = ""
  app_key = ""
  app_secret = ""
  ## Pusher requires a channel name to be specified
  channel_name = ""
  ## Whether to use https (true) or not (false)
  secure = true
  ## Modify if your Pusher Cluster is not USA (e.g. EU or Asia)
  host = "api.pusherapp.com"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options; read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```
