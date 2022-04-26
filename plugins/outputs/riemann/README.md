# Riemann Output Plugin

This plugin writes to [Riemann](http://riemann.io/) via TCP or UDP.

## Configuration

```toml
# Configuration for Riemann to send metrics to
[[outputs.riemann]]
  ## The full TCP or UDP URL of the Riemann server
  url = "tcp://localhost:5555"

  ## Riemann event TTL, floating-point time in seconds.
  ## Defines how long that an event is considered valid for in Riemann
  # ttl = 30.0

  ## Separator to use between measurement and field name in Riemann service name
  ## This does not have any effect if 'measurement_as_attribute' is set to 'true'
  separator = "/"

  ## Set measurement name as Riemann attribute 'measurement', instead of prepending it to the Riemann service name
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

  ## Riemann client write timeout, defaults to "5s" if not set.
  # timeout = "5s"
```

### Required parameters

* `url`: The full TCP or UDP URL of the Riemann server to send events to.

### Optional parameters

* `ttl`: Riemann event TTL, floating-point time in seconds. Defines how long
  that an event is considered valid for in Riemann.
* `separator`: Separator to use between measurement and field name in Riemann
  service name.
* `measurement_as_attribute`: Set measurement name as a Riemann attribute,
  instead of prepending it to the Riemann service name.
* `string_as_state`: Send string metrics as Riemann event states. If this is not
  enabled then all string metrics will be ignored.
* `tag_keys`: A list of tag keys whose values get sent as Riemann tags. If
  empty, all Telegraf tag values will be sent as tags.
* `tags`: Additional Riemann tags that will be sent.
* `description_text`: Description text for Riemann event.

## Example Events

Riemann event emitted by Telegraf with default configuration:

```text
#riemann.codec.Event{
:host "postgresql-1e612b44-e92f-4d27-9f30-5e2f53947870", :state nil, :description nil, :ttl 30.0,
:service "disk/used_percent", :metric 73.16736001949994, :path "/boot", :fstype "ext4", :time 1475605021}
```

Telegraf emitting the same Riemann event with `measurement_as_attribute` set to
`true`:

```text
#riemann.codec.Event{ ...
:measurement "disk", :service "used_percent", :metric 73.16736001949994,
... :time 1475605021}
```

Telegraf emitting the same Riemann event with additional Riemann tags defined:

```text
#riemann.codec.Event{
:host "postgresql-1e612b44-e92f-4d27-9f30-5e2f53947870", :state nil, :description nil, :ttl 30.0,
:service "disk/used_percent", :metric 73.16736001949994, :path "/boot", :fstype "ext4", :time 1475605021,
:tags ["telegraf" "postgres_cluster"]}
```

Telegraf emitting a Riemann event with a status text and `string_as_state` set
to `true`, and a `description_text` defined:

```text
#riemann.codec.Event{
:host "postgresql-1e612b44-e92f-4d27-9f30-5e2f53947870", :state "Running", :ttl 30.0,
:description "PostgreSQL master node is up and running",
:service "status", :time 1475605021}
```
