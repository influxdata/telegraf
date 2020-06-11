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
