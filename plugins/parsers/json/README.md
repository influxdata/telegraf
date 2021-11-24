# JSON

The JSON data format parses a [JSON][json] object or an array of objects into
metric fields.

**NOTE:** All JSON numbers are converted to float fields.  JSON strings and booleans are
ignored unless specified in the `tag_key` or `json_string_fields` options.

## Configuration

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "json"

  ## When strict is true and a JSON array is being parsed, all objects within the
  ## array must be valid
  json_strict = true

  ## Query is a GJSON path that specifies a specific chunk of JSON to be
  ## parsed, if not specified the whole document will be parsed.
  ##
  ## GJSON query paths are described here:
  ##   https://github.com/tidwall/gjson/tree/v1.3.0#path-syntax
  json_query = ""

  ## Tag keys is an array of keys that should be added as tags.  Matching keys
  ## are no longer saved as fields. Supports wildcard glob matching.
  tag_keys = [
    "my_tag_1",
    "my_tag_2",
    "tags_*",
    "tag*"
  ]

  ## Array of glob pattern strings or booleans keys that should be added as string fields.
  json_string_fields = []

  ## Name key is the key to use as the measurement name.
  json_name_key = ""

  ## Time key is the key containing the time that should be used to create the
  ## metric.
  json_time_key = ""

  ## Time format is the time layout that should be used to interpret the json_time_key.
  ## The time must be `unix`, `unix_ms`, `unix_us`, `unix_ns`, or a time in the
  ## "reference time".  To define a different format, arrange the values from
  ## the "reference time" in the example to match the format you will be
  ## using.  For more information on the "reference time", visit
  ## https://golang.org/pkg/time/#Time.Format
  ##   ex: json_time_format = "Mon Jan 2 15:04:05 -0700 MST 2006"
  ##       json_time_format = "2006-01-02T15:04:05Z07:00"
  ##       json_time_format = "01/02/2006 15:04:05"
  ##       json_time_format = "unix"
  ##       json_time_format = "unix_ms"
  json_time_format = ""

  ## Timezone allows you to provide an override for timestamps that
  ## don't already include an offset
  ## e.g. 04/06/2016 12:41:45
  ##
  ## Default: "" which renders UTC
  ## Options are as follows:
  ##   1. Local               -- interpret based on machine localtime
  ##   2. "America/New_York"  -- Unix TZ values like those found in https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
  ##   3. UTC                 -- or blank/unspecified, will return timestamp in UTC
  json_timezone = ""
```

### json_query

The `json_query` is a [GJSON][gjson] path that can be used to transform the
JSON document before being parsed.  The query is performed before any other
options are applied and the new document produced will be parsed instead of the
original document, as such, the result of the query should be a JSON object or
an array of objects.

Consult the GJSON [path syntax][gjson syntax] for details and examples, and
consider using the [GJSON playground][gjson playground] for developing and
debugging your query.

### json_time_key, json_time_format, json_timezone

By default the current time will be used for all created metrics, to set the
time using the JSON document you can use the `json_time_key` and
`json_time_format` options together to set the time to a value in the parsed
document.

The `json_time_key` option specifies the key containing the time value and
`json_time_format` must be set to `unix`, `unix_ms`, `unix_us`, `unix_ns`, or
the Go "reference time" which is defined to be the specific time:
`Mon Jan 2 15:04:05 MST 2006`.

Consult the Go [time][time parse] package for details and additional examples
on how to set the time format.

When parsing times that don't include a timezone specifier, times are assumed
to be UTC. To default to another timezone, or to local time, specify the
`json_timezone` option.  This option should be set to a
[Unix TZ value](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones),
such as `America/New_York`, to `Local` to utilize the system timezone, or to `UTC`.

## Examples

### Basic Parsing

Config:

```toml
[[inputs.file]]
  files = ["example"]
  name_override = "myjsonmetric"
  data_format = "json"
```

Input:

```json
{
    "a": 5,
    "b": {
        "c": 6
    },
    "ignored": "I'm a string"
}
```

Output:

```text
myjsonmetric a=5,b_c=6
```

### Name, Tags, and String Fields

Config:

```toml
[[inputs.file]]
  files = ["example"]
  json_name_key = "name"
  tag_keys = ["my_tag_1"]
  json_string_fields = ["b_my_field"]
  data_format = "json"
```

Input:

```json
{
    "a": 5,
    "b": {
        "c": 6,
        "my_field": "description"
    },
    "my_tag_1": "foo",
    "name": "my_json"
}
```

Output:

```text
my_json,my_tag_1=foo a=5,b_c=6,b_my_field="description"
```

### Arrays

If the JSON data is an array, then each object within the array is parsed with
the configured settings.

Config:

```toml
[[inputs.file]]
  files = ["example"]
  data_format = "json"
  json_time_key = "b_time"
  json_time_format = "02 Jan 06 15:04 MST"
```

Input:

```json
[
    {
        "a": 5,
        "b": {
            "c": 6,
            "time":"04 Jan 06 15:04 MST"
        }
    },
    {
        "a": 7,
        "b": {
            "c": 8,
            "time":"11 Jan 07 15:04 MST"
        }
    }
]
```

Output:

```text
file a=5,b_c=6 1136387040000000000
file a=7,b_c=8 1168527840000000000
```

### Query

The `json_query` option can be used to parse a subset of the document.

Config:

```toml
[[inputs.file]]
  files = ["example"]
  data_format = "json"
  tag_keys = ["first"]
  json_string_fields = ["last"]
  json_query = "obj.friends"
```

Input:

```json
{
    "obj": {
        "name": {"first": "Tom", "last": "Anderson"},
        "age":37,
        "children": ["Sara","Alex","Jack"],
        "fav.movie": "Deer Hunter",
        "friends": [
            {"first": "Dale", "last": "Murphy", "age": 44},
            {"first": "Roger", "last": "Craig", "age": 68},
            {"first": "Jane", "last": "Murphy", "age": 47}
        ]
    }
}
```

Output:

```text
file,first=Dale last="Murphy",age=44
file,first=Roger last="Craig",age=68
file,first=Jane last="Murphy",age=47
```

[gjson]:        https://github.com/tidwall/gjson
[gjson syntax]: https://github.com/tidwall/gjson#path-syntax
[gjson playground]: https://gjson.dev/
[json]:         https://www.json.org/
[time parse]:   https://golang.org/pkg/time/#Parse
