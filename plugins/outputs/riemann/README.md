# Riemann Output Plugin

This plugin writes to [Riemann](http://riemann.io/) via TCP or UDP.

### Configuration:

```toml
# Configuration for Riemann to send metrics to
[[outputs.riemann]]
  ## Address of the Riemann server
  address = "localhost:5555"

  ## Transport protocol to use, either tcp or udp
  transport = "tcp"

  ## Riemann TTL, floating-point time in seconds.
  ## Defines how long that an event is considered valid for in Riemann
  # ttl = 30.0

  ## Separator to use between measurement and field name in Riemann service name
  separator = "/"

  ## Set measurement name as a Riemann attribute,
  ## instead of prepending it to the Riemann service name
  # measurement_as_attribute = false

  ## Send string metrics as Riemann event states.
  ## Unless enabled all string metrics will be ignored
  # string_as_state = false

  ## A list of tag keys whose values get sent as Riemann tags.
  ## If empty, all Telegraf tag values will be sent as tags
  # tag_keys = ["telegraf","custom_tag"]

  ## Additional Riemann tags to send.
  # tags = ["telegraf-output"]

  ## Description for Riemann event
  # description_text = "metrics collected from telegraf"
```

### Required parameters:

* `address`: Address of the Riemann server to send Riemann events to.
* `transport`: Transport protocol to use, must be either tcp or udp.

### Optional parameters:

* `ttl`: Riemann event TTL, floating-point time in seconds. Defines how long that an event is considered valid for in Riemann.
* `separator`: Separator to use between measurement and field name in Riemann service name.
* `measurement_as_attribute`: Set measurement name as a Riemann attribute, instead of prepending it to the Riemann service name.
* `string_as_state`: Send string metrics as Riemann event states. If this is not enabled then all string metrics will be ignored.
* `tag_keys`: A list of tag keys whose values get sent as Riemann tags. If empty, all Telegraf tag values will be sent as tags.
* `tags`: Additional Riemann tags that will be sent.
* `description_text`: Description text for Riemann event.
