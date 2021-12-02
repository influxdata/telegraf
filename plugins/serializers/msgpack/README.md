# MessagePack

[MessagePack](https://msgpack.org) is an efficient binary serialization format. It lets you exchange data among multiple languages like JSON.

## Format Definitions

Output of this format is MessagePack binary representation of metrics that have identical structure of the below JSON.

```json
{
   "name":"cpu",
   "time": <TIMESTAMP>, // https://github.com/msgpack/msgpack/blob/master/spec.md#timestamp-extension-type
   "tags":{
      "tag_1":"host01",
      ...
   },
   "fields":{
      "field_1":30,
      "field_2":true,
      "field_3":"field_value"
      "field_4":30.1
      ...
   }
}
```

MessagePack has it's own timestamp representation. You can find additional informations from [MessagePack specification](https://github.com/msgpack/msgpack/blob/master/spec.md#timestamp-extension-type).

## MessagePack Configuration

There are no additional configuration options for MessagePack format.

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "msgpack"
```
