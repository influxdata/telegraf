# Encoding Processor

The `encoding` processor provides compression/decompression of string fields along with conversion to/from relevant encodings that function with Influx and line protocol.

## Configuration

```toml
[[processors.encoding]]
  ## (required) Field specifies which string field to operate on
  # field = ""

  ## (required) Encoding is the algorithm used to encode the compressed binary
  ## data into a string Influx can process. Only base64 is supported for now.
  ## Because compression deals with binary data (not supported by Influx),
  ## encoding is required. However, encoding may be used with no compression if
  ## desired.
  # encoding = "base64"

  ## Destination field is the field where the encoding result will be stored.
  ## If not specified, field is used.
  # dest_field = ""

  ## Whether the original field should be removed if it doesn't match dest_field.
  # remove_original = false

  ## Operation determines whether to "encode" or "decode"
  # operation = "decode"

  ## Compression describes the compression algorithm used for the field. If
  ## empty, compression is skipped. Only gzip is supported for now. Compression
  ## requires that an encoding be set as Influx does not support binary data.
  # compression = "gzip"

  ## Compression field and tag allow the compression algorithm to be retrieved
  ## from a field or tag on a metric. Tag takes precedence over field. If
  ## neither is found, "compression" (if any) is used for the metric.
  # compression_field = ""
  # compression_tag = ""
```
