# Pusher Output Plugin

This output plugin writes to the [Pusher REST API](https://pusher.com/docs/rest_api).

## Configuration

This plugin's configuration specifies a Pusher channel that incoming events should be published to.

The plugin will read Telegraf metric names and use those as the corresponding Pusher event names.

```
# Configuration for Pusher output.
[[outputs.pusher]]
  ## Pusher Credentials
  #app_id = ""
  #app_key = ""
  #app_secret = ""
  #channel_name = ""
  secure = true
  host = "app.pusher.com"

  data_format = "json"
```
