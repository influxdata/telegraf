# Influx protocol-buffer serializer

This serializer outputs metrics in the binary
[InfluxDB Protobuf Data Protocol][1]. This protocol is optimized for efficient
ingestion into IOx, and is also a good choice for transporting timeseries data
in general. The protocol-buffer definition can be found [here][2].

Please also read the [optimization notes][3] if you intent to parse the
serialized data yourself.

## Example

This is the most simple usage example for this serializer.

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx_protobuf"

  ## Database name to use in DatabaseBatch.
  # influx_protobuf_database = ""

  ## Serialize to line-protocol-like representation if `false`
  ## and to IOx representation otherwise.
  ## Please note: IOx representation will loose information on tag-field separation.
  # influx_protobuf_iox = false
```

The shown configuration will serialize metrics to the line-protocol-like binary
format preserving tag and field separation with an empty _database name_.
For a more sophisticated use-case check the configuration options below.

## Options

### `influx_protobuf_database` (string)

With this option you can set the `DatabaseName` in `DatabaseBatch`. This can
be handy in case you want to directly push the serialized data to an InfluxDB
or IOx instance. By default, the database-name will be empty.

### `influx_protobuf_iox` (bool)

If set to true, the `IOx` [semantic type][4] will be used for all columns.
Please note, that this type will drop information on whether a column is a tag
or a field and is most suited to be ingested by IOx itself. By default,
`Tag` and `Field` [semantic types][4] will be used corresponding to the
traditional line-protocol elements.

[1]: https://github.com/influxdata/influxdb-pb-data-protocol
[2]: https://github.com/influxdata/influxdb-pb-data-protocol/blob/main/influxdb-pb-data-protocol.proto
[3]: https://github.com/influxdata/influxdb-pb-data-protocol#optimizations
[4]: https://github.com/influxdata/influxdb-pb-data-protocol#semantic-type
