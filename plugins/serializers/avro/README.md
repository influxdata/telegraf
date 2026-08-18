# Avro Serializer Plugin

The `avro` output data format encodes metrics into Avro using a schema you
provide. Each metric becomes one Avro record. Tag values, field values and,
optionally, the measurement name and timestamp are matched to the schema fields
by name, so the schema decides what ends up in the output and in what order.

This is the counterpart to the [Avro parser][parser]. It emits bare Avro binary
(or Avro JSON), not Confluent wire format, so there's no schema-registry
interaction and no schema ID prefix on the output.

[parser]: /plugins/parsers/avro

## Configuration

```toml
[[outputs.file]]
  files = ["stdout"]

  ## Data format to output.
  data_format = "avro"

  ## Avro schema used to encode the metrics. Required.
  ## Must be a record schema. Fields the schema doesn't name are dropped;
  ## schema fields with no matching tag/field must be nullable or have a
  ## default, otherwise encoding fails.
  avro_schema = '''
    {
      "type": "record",
      "name": "metric",
      "fields": [
        {"name": "host", "type": "string"},
        {"name": "value", "type": "double"},
        {"name": "timestamp", "type": "long"}
      ]
    }
  '''

  ## Output encoding: "binary" (default) or "json".
  # avro_format = "binary"

  ## Schema field to receive the measurement name. Leave empty to skip.
  # avro_measurement_field = ""

  ## Schema field to receive the timestamp. Leave empty to skip.
  # avro_timestamp = "timestamp"

  ## How to encode the timestamp when avro_timestamp is set.
  ## One of "unix" (seconds, default), "unix_ms", "unix_us", "unix_ns".
  # avro_timestamp_format = "unix"
```

### Type coercion

Telegraf field values are coerced to the Avro type declared for the matching
schema field, so an `int64` metric field maps cleanly onto a `double` or
`float` schema field and a numeric field maps onto `string`. Supported target
types are `boolean`, `int`, `long`, `float`, `double`, `string` and `bytes`.

A union of the form `["null", T]` is handled by emitting `null` for a missing
value and the `T` branch otherwise. More complex types (records, arrays, maps,
logical types) are passed through to the encoder unchanged.

## Example

With the schema above and the timestamp field configured, the line protocol
metric

```text
metric,host=server01 value=42.5 1600000000000000000
```

serializes to an Avro record equivalent to

```json
{"host": "server01", "value": 42.5, "timestamp": 1600000000}
```
