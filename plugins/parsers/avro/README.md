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

  ## Avro message format
  ## Supported values are "binary" (default) and "json"
  # avro_format = "binary"

  ## Url of the schema registry; exactly one of schema registry and
  ## schema must be set
  avro_schema_registry = "http://localhost:8081"

  ## Schema string; exactly one of schema registry and schema must be set
  #avro_schema = '''
  #        {
  #          "type":"record",
  #          "name":"Value",
  #          "namespace":"com.example",
  #          "fields":[
  #              {
  #                  "name":"tag",
  #                  "type":"string"
  #              },
  #              {
  #                  "name":"field",
  #                  "type":"long"
  #              },
  #              {
  #                  "name":"timestamp",
  #                  "type":"long"
  #              }
  #          ]
  #      }
  #'''

  ## Measurement string; if not set, determine measurement name from
  ## schema (as "<namespace>.<name>")
  # avro_measurement = "ratings"

  ## Avro fields to be used as tags; optional.
  # avro_tags = ["CHANNEL", "CLUB_STATUS"]

  ## Avro fields to be used as fields; if empty, any Avro fields
  ## detected from the schema, not used as tags, will be used as
  ## measurement fields.
  # avro_fields = ["STARS"]

  ## Avro fields to be used as timestamp; if empty, current time will
  ## be used for the measurement timestamp.
  # avro_timestamp = ""
  ## If avro_timestamp is specified, avro_timestamp_format must be set
  ## to one of 'unix', 'unix_ms', 'unix_us', or 'unix_ns'
  # avro_timestamp_format = "unix"

  ## Used to separate parts of array structures.  As above, the default
  ## is the empty string, so a=["a", "b"] becomes a0="a", a1="b".
  ## If this were set to "_", then it would be a_0="a", a_1="b".
  # avro_field_separator = "_"

  ## Default values for given tags: optional
  # tags = { "application": "hermes", "region": "central" }

```

### `avro_format`

This optional setting specifies the format of the Avro messages. Currently, the
parser supports the `binary` and `json` formats with `binary` being the default.

### `avro_timestamp` and `avro_timestamp_format`

By default the current time at ingestion will be used for all created
metrics; to set the time using the Avro message you can use the
`avro_timestamp` and `avro_timestamp_format` options together to set the
time to a value in the parsed document.

The `avro_timestamp` option specifies the field containing the time
value.  If it is not set, the time of record ingestion is used.  If it
is set, the field may be any numerical type: notably, it is *not*
constrained to an Avro `long` (int64) (which Avro uses for timestamps in
millisecond or microsecond resolution).  However, it must represent the
number of time increments since the Unix epoch (00:00 UTC 1 Jan 1970).

The `avro_timestamp_format` option specifies the precision of the timestamp
field, and, if set, must be one of `unix`, `unix_ms`, `unix_us`, or
`unix_ns`.  If `avro_timestamp` is set, `avro_timestamp_format` must be
as well.

## Metrics

One metric is created for each message.  The type of each field is
automatically determined based on the schema.
