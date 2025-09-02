# Avro Parser Plugin

The `Avro` parser creates metrics from a message serialized with Avro.

The message is supposed to be encoded as follows:

| Bytes | Area       | Description                                      |
| ----- | ---------- | ------------------------------------------------ |
| 0     | Magic Byte | Confluent serialization format version number.   |
| 1-4   | Schema ID  | 4-byte schema ID as returned by Schema Registry. |
| 5-    | Data       | Serialized data.                                 |

The metric name will be set according the following priority:

  1. Try to get metric name from the message field if it is set in the
     `avro_measurement_field` option.
  2. If the name is not determined, then try to get it from
     `avro_measurement` option as the static value.
  3. If the name is still not determined, then try to get it from the
     schema definition in the following format `[schema_namespace.]schema_name`,
     where schema namespace is optional and will be added only if it is specified
     in the schema definition.

In case if the metric name could not be determined according to these steps
the error will be raised and the message will not be parsed.

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

  ## URL of the schema registry which may contain username and password in the
  ## form http[s]://[username[:password]@]<host>[:port]
  ## NOTE: Exactly one of schema registry and schema must be set
  avro_schema_registry = "http://localhost:8081"

  ## Path to the schema registry certificate. Should be specified only if
  ## required for connection to the schema registry.
  # avro_schema_registry_cert = "/etc/telegraf/ca_cert.crt"

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

  ## Measurement field name; The meauserment name will be taken 
  ## from this field. If not set, determine measurement name
  ## from the following 'avro_measurement' option
  # avro_measurement_field = "field_name"

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
  ## to one of 'unix', 'unix_ms', 'unix_us', or 'unix_ns'.  It will
  ## default to 'unix'.
  # avro_timestamp_format = "unix"

  ## Used to separate parts of array structures.  As above, the default
  ## is the empty string, so a=["a", "b"] becomes a0="a", a1="b".
  ## If this were set to "_", then it would be a_0="a", a_1="b".
  # avro_field_separator = "_"

  ## Define handling of union types. Possible values are:
  ##   flatten  -- add type suffix to field name (default)
  ##   nullable -- do not modify field name but discard "null" field values
  ##   any      -- do not modify field name and set field value to the received type
  # avro_union_mode = "flatten"

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

If the root of the schema is a record, one metric is created for each message.
If the root of the schema is an array containing records as elements,
the array elements are parsed as separate metrics.
Nested data structures within records are flattened.
