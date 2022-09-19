# Avro Parser Plugin

The `Avro` parser creates metrics from a message serialized with Avro.

The message is supposed to be encoded as follows:

| Bytes | Area       | Description                                      |
| ----- | ---------- | ------------------------------------------------ |
| 0     | Magic Byte | Confluent serialization format version number.   |
| 1-4   | Schema ID  | 4-byte schema ID as returned by Schema Registry. |
| 5-    | Data       | Serialized data.                                 |

## Configuration

```toml
[[inputs.kafka_consumer]]
  ## Kafka brokers.
  brokers = ["localhost:9092"]

  ## Topics to consume.
  topics = ["telegraf"]

  ## Maximum length of a message to consume, in bytes (default 0/unlimited);
  ## larger messages are dropped
  max_message_len = 1000000

  ## Avro data format settings
  data_format = "avro"

  ## Url of the schema registry
  avro_schema_registry = "http://schema-registry:8081"

  ## Schema string; only used if schema registry is not set
  avro_schema = """
          {
            "type":"record",
            "name":"Value",
            "namespace":"com.example",
            "fields":[
                {
                    "name":"tag",
                    "type":"string"
                },
                {
                    "name":"field",
                    "type":"long"
                },
                {
                    "name":"timestamp",
                    "type":"long"
                }
            ]
        }
  """

  ## Measurement string; if not set, determine measurement name from
  ## schema (as "<namespace>.<name>")
  # avro_measurement = "ratings"

  ## Avro fields to be used as tags
  avro_tags = ["CHANNEL", "CLUB_STATUS"]

  ## Avro fields to be used as fields; if empty, any Avro fields
  ## detected from the schema, not used as tags, will be used as
  ## measurement fields.
  # avro_fields = ["STARS"]

  ## Avro fields to be used as timestamp; if empty, current time will
  ## be used for the measurement timestamp.
  avro_timestamp = "TIMESTAMP"

  ## Timestamp format
  avro_timestamp_format = "unix_ms"

  ## If true, any array values received by Avro will be silently
  ## discarded; otherwise they will be converted into a series of
  ## scalars, e.g. a=["a", "b", "c"] would become a0="a", a1="b",
  ## a2="c".
  # avro_discard_arrays = true

  ## Used to separate parts of array structures.  As above, the default
  ## is the empty string, so a=["a", "b"] becomes a0="a", a1="b".
  ## If this were set to "_", then it would be a_0="a", a_1="b".
  # avro_field_separator = "_"
```

### avro_timestamp, avro_timestamp_format

By default the current time will be used for all created metrics, to set
the time using the Avro message you can use the `avro_timestamp` and
`avro_timestamp_format` options together to set the time to a value in
the parsed document.

The `avro_timestamp` option specifies the field containing the time
value and `avro_timestamp_format` must be set to `unix`, `unix_ms`,
`unix_us`, `unix_ns`, `unix_float_ms`, `unix_float_us`, or
`unix_float_ns`.  The `unix` and `unix_float_ns` formats are identical
and can parse a timestamp in a floating-point type to nanoseconds.  The
other two `unix_float` timestamps round the time to the nearest
millisecond or microsecond.

## Metrics

One metric is created for each message.  The type of each field is
automatically determined based on the schema.
