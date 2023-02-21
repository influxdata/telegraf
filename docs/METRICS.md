# Metrics

Telegraf metrics are the internal representation used to model data during
processing.  Metrics are closely based on InfluxDB's data model and contain
four main components:

- **Measurement Name**: Description and namespace for the metric.
- **Tags**: Key/Value string pairs and usually used to identify the
  metric.
- **Fields**: Key/Value pairs that are typed and usually contain the
  metric data.
- **Timestamp**: Date and time associated with the fields.

This metric type exists only in memory and must be converted to a concrete
representation in order to be transmitted or viewed.  To achieve this we
provide several [output data formats][] sometimes referred to as
*serializers*.  Our default serializer converts to [InfluxDB Line
Protocol][line protocol] which provides a high performance and one-to-one
direct mapping from Telegraf metrics.

[output data formats]: /docs/DATA_FORMATS_OUTPUT.md
[line protocol]: /plugins/serializers/influx

## Tracking Metrics

Tracking metrics are metrics that ensure that data is passed from the input and
handed to an output before acknowledging the message back to the input. The
use case for these types of metrics is to ensure that the message makes it to
the destination before removing the metric from the input.

For example, if a configuration is reading from MQTT, Kafka, or an AMQP source
Telegraf will read the message and wait till the metric is handed to the output
before telling the metric source that the message was read. If Telegraf were to
stop or the system running Telegraf to crash, this allows the messages that
were not completely delivered to an output to get re-read at a later date.

### Undelivered Messages

When an input uses tracking metrics, an additional setting,
`max_undelivered_messages`, is available in that plugin. This setting
determines how many metrics should be read in before reading additional
messages. In practice, this means that Telegraf may not read new messages from
an input at every collection interval.

Users need to use caution with this setting. Setting the value too high may
mean that Telegraf pushes constant batches to an output, ignoring the flush
interval.
