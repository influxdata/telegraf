# CloudEvents Serializer

The `cloudevents` data format outputs metrics as [CloudEvents][CloudEvents] in
[JSON format][JSON Spec]. Currently, versions v1.0 and v0.3 of the specification
are supported with the former being the default.

[CloudEvents]: https://cloudevents.io
[JSON Spec]: https://github.com/cloudevents/spec/blob/v1.0/json-format.md

## Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "cloudevents"

  ## Specification version to use for events
  ## Currently versions "0.3" and "1.0" are supported.
  # cloudevents_version = "1.0"

  ## Event source specifier
  ## This allows to overwrite the source header-field with the given value.
  # cloudevents_source = "telegraf"

  ## Tag to use as event source specifier
  ## This allows to overwrite the source header-field with the value of the
  ## specified tag. If both 'cloudevents_source' and 'cloudevents_source_tag'
  ## are set, the this setting will take precedence. In case the specified tag
  ## value does not exist for a metric, the serializer will fallback to
  ## 'cloudevents_source'.
  # cloudevents_source_tag = ""

  ## Event-type specifier to overwrite the default value
  ## By default, events (and event batches) containing a single metric will
  ## set the event-type to 'com.influxdata.telegraf.metric' while events
  ## containing a batch of metrics will set the event-type to
  ## 'com.influxdata.telegraf.metric' (plural).
  # cloudevents_event_type = ""

  ## Set time header of the event
  ## Supported values are:
  ##   none     -- do not set event time
  ##   earliest -- use timestamp of the earliest metric
  ##   latest   -- use timestamp of the latest metric
  ##   creation -- use timestamp of event creation
  ## For events containing only a single metric, earliest and latest are
  ## equivalent.
  # cloudevents_event_time = "latest"

  ## Batch format of the output when running in batch mode
  ## If set to 'events' the resulting output will contain a list of events,
  ## each with a single metric according to the JSON Batch Format of the
  ## specification. Use 'application/cloudevents-batch+json' for this format.
  ##
  ## When set to 'metrics', a single event will be generated containing a list
  ## of metrics as payload. Use 'application/cloudevents+json' for this format.
  # cloudevents_batch_format = "events"
```
