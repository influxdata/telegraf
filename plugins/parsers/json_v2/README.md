# JSON Parser - Version 2

This parser takes valid JSON input and turns it into metrics. The query syntax supported is [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md), you can go to this playground to test out your GJSON path here: https://gjson.dev/. You can find multiple examples under the `testdata` folder.

## Configuration

You configure this parser by describing the metric you want by defining the fields and tags from the input. The configuration is divided into config sub-tables called `field`, `tag`, and `object`. In the example below you can see all the possible configuration keys you can define for each config table. In the sections that follow these configuration keys are defined in more detail.

**Example configuration:**

```toml
 [[inputs.file]]
    urls = []
    data_format = "json_v2"
    [[inputs.file.json_v2]]
        measurement_name = "" # A string that will become the new measurement name
        measurement_name_path = "" # A string with valid GJSON path syntax, will override measurement_name
        timestamp_path = "" # A string with valid GJSON path syntax to a valid timestamp (single value)
        timestamp_format = "" # A string with a valid timestamp format (see below for possible values)
        timestamp_timezone = "" # A string with with a valid timezone (see below for possible values)
        [[inputs.file.json_v2.tag]]
            path = "" # A string with valid GJSON path syntax
            rename = "new name" # A string with a new name for the tag key
        [[inputs.file.json_v2.field]]
            path = "" # A string with valid GJSON path syntax
            rename = "new name" # A string with a new name for the tag key
            type = "int" # A string specifying the type (int,uint,float,string,bool)
        [[inputs.file.json_v2.object]]
            path = "" # A string with valid GJSON path syntax
            timestamp_key = "" # A JSON key (for a nested key, prepend the parent keys with underscores) to a valid timestamp
            timestamp_format = "" # A string with a valid timestamp format (see below for possible values)
            timestamp_timezone = "" # A string with with a valid timezone (see below for possible values)
            disable_prepend_keys = false (or true, just not both)
            included_keys = [] # List of JSON keys (for a nested key, prepend the parent keys with underscores) that should be only included in result
            excluded_keys = [] # List of JSON keys (for a nested key, prepend the parent keys with underscores) that shouldn't be included in result
            tags = [] # List of JSON keys (for a nested key, prepend the parent keys with underscores) to be a tag instead of a field
            [inputs.file.json_v2.object.renames] # A map of JSON keys (for a nested key, prepend the parent keys with underscores) with a new name for the tag key
                key = "new name"
            [inputs.file.json_v2.object.fields] # A map of JSON keys (for a nested key, prepend the parent keys with underscores) with a type (int,uint,float,string,bool)
                key = "int"
```
---
### root config options

* **measurement_name (OPTIONAL)**:  Will set the measurement name to the provided string.
* **measurement_name_path (OPTIONAL)**: You can define a query with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md) to set a measurement name from the JSON input. The query must return a single data value or it will use the default measurement name. This takes precedence over `measurement_name`.
* **timestamp_path (OPTIONAL)**: You can define a query with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md) to set a timestamp from the JSON input. The query must return a single data value or it will default to the current time.
* **timestamp_format (OPTIONAL, but REQUIRED when timestamp_query is defined**: Must be set to `unix`, `unix_ms`, `unix_us`, `unix_ns`, or
the Go "reference time" which is defined to be the specific time:
`Mon Jan 2 15:04:05 MST 2006`
* **timestamp_timezone (OPTIONAL, but REQUIRES timestamp_query**: This option should be set to a
[Unix TZ value](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones),
such as `America/New_York`, to `Local` to utilize the system timezone, or to `UTC`. Defaults to `UTC`

---

### `field` and `tag` config options

`field` and `tag` represent the elements of [line protocol](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/), which is used to define a `metric`. You can use the `field` and `tag` config tables to gather a single value or an array of values that all share the same type and name. With this you can add a field or tag to a metric from data stored anywhere in your JSON. If you define the GJSON path to return a single value then you will get a single resutling metric that contains the field/tag. If you define the GJSON path to return an array of values, then each field/tag will be put into a separate metric (you use the # character to retrieve JSON arrays, find examples [here](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md#arrays)).

Note that objects are handled separately, therefore if you provide a path that returns a object it will be ignored. You will need use the `object` config table to parse objects, because `field` and `tag` doesn't handle relationships between data. Each `field` and `tag` you define is handled as a separate data point.

The notable difference between `field` and `tag`, is that `tag` values will always be type string while `field` can be multiple types. You can define the type of `field` to be any [type that line protocol supports](https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/#data-types-and-format), which are:
* float
* int
* uint
* string
* bool


#### **field**

* **path (REQUIRED)**: You must define the path query that gathers the object with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md).
* **name (OPTIONAL)**: You can define a string value to set the field name. If not defined it will use the trailing word from the provided query.
* **type (OPTIONAL)**: You can define a string value to set the desired type (float, int, uint, string, bool). If not defined it won't enforce a type and default to using the original type defined in the JSON (bool, float, or string).

#### **tag**

* **path (REQUIRED)**: You must define the path query that gathers the object with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md).
* **name (OPTIONAL)**: You can define a string value to set the field name. If not defined it will use the trailing word from the provided query.

For good examples in using `field` and `tag` you can reference the following example configs:

* [fields_and_tags](testdata/fields_and_tags/telegraf.conf)
---
### object

With the configuration section `object`, you can gather metrics from [JSON objects](https://www.w3schools.com/js/js_json_objects.asp).

The following keys can be set for `object`:

* **path (REQUIRED)**: You must define the path query that gathers the object with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md)
* **timestamp_key(OPTIONAL)**: You can define a json key (for a nested key, prepend the parent keys with underscores) for the value to be set as the timestamp from the JSON input.
* **timestamp_format (OPTIONAL, but REQUIRED when timestamp_query is defined**: Must be set to `unix`, `unix_ms`, `unix_us`, `unix_ns`, or
the Go "reference time" which is defined to be the specific time:
`Mon Jan 2 15:04:05 MST 2006`
* **timestamp_timezone (OPTIONAL, but REQUIRES timestamp_query**: This option should be set to a
[Unix TZ value](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones),
such as `America/New_York`, to `Local` to utilize the system timezone, or to `UTC`. Defaults to `UTC`
* **disable_prepend_keys (OPTIONAL)**: Set to true to prevent resulting nested data to contain the parent key prepended to its key **NOTE**: duplicate names can overwrite each other when this is enabled
* **included_keys (OPTIONAL)**: You can define a list of key's that should be the only data included in the metric, by default it will include everything.
* **excluded_keys (OPTIONAL)**: You can define json keys to be excluded in the metric, for a nested key, prepend the parent keys with underscores
* **tags (OPTIONAL)**: You can define json keys to be set as tags instead of fields, if you define a key that is an array or object then all nested values will become a tag
* **renames (OPTIONAL)**: A table matching the json key with the desired name (oppossed to defaulting to using the key), use names that include the prepended keys of its parent keys for nested results
* **fields (OPTIONAL)**: A table matching the json key with the desired type (int,string,bool,float), if you define a key that is an array or object then all nested values will become that type

## Arrays and Objects

The following describes the high-level approach when parsing arrays and objects:

**Array**: Every element in an array is treated as a *separate* metric

**Object**: Every key/value in a object is treated as a *single* metric

When handling nested arrays and objects, these above rules continue to apply as the parser creates metrics. When an object has multiple array's as values, the array's will become separate metrics containing only non-array values from the obejct. Below you can see an example of this behavior, with an input json containing an array of book objects that has a nested array of characters.

Example JSON:

```json
{
    "book": {
        "title": "The Lord Of The Rings",
        "chapters": [
            "A Long-expected Party",
            "The Shadow of the Past"
        ],
        "author": "Tolkien",
        "characters": [
            {
                "name": "Bilbo",
                "species": "hobbit"
            },
            {
                "name": "Frodo",
                "species": "hobbit"
            }
        ],
        "random": [
            1,
            2
        ]
    }
}

```

Example configuration:

```toml
[[inputs.file]]
    files = ["./testdata/multiple_arrays_in_object/input.json"]
    data_format = "json_v2"
    [[inputs.file.json_v2]]
        [[inputs.file.json_v2.object]]
            path = "book"
            tags = ["title"]
            disable_prepend_keys = true
```

Expected metrics:

```
file,title=The\ Lord\ Of\ The\ Rings author="Tolkien",chapters="A Long-expected Party"
file,title=The\ Lord\ Of\ The\ Rings author="Tolkien",chapters="The Shadow of the Past"
file,title=The\ Lord\ Of\ The\ Rings author="Tolkien",name="Bilbo",species="hobbit"
file,title=The\ Lord\ Of\ The\ Rings author="Tolkien",name="Frodo",species="hobbit"
file,title=The\ Lord\ Of\ The\ Rings author="Tolkien",random=1
file,title=The\ Lord\ Of\ The\ Rings author="Tolkien",random=2

```

You can find more complicated examples under the folder `testdata`.

## Types

For each field you have the option to define the types for each metric. The following rules are in place for this configuration:

* If a type is explicitly defined, the parser will enforce this type and convert the data to the defined type if possible. If the type can't be converted then the parser will fail.
* If a type isn't defined, the parser will use the default type defined in the JSON (int, float, string)

The type values you can set:

* `int`, bool, floats or strings (with valid numbers) can be converted to a int.
* `uint`, bool, floats or strings (with valid numbers) can be converted to a uint.
* `string`, any data can be formatted as a string.
* `float`, string values (with valid numbers) or integers can be converted to a float.
* `bool`, the string values "true" or "false" (regardless of capitalization) or the integer values `0` or `1`  can be turned to a bool.

## Migrating from the original JSON parser to json_v2

The json_v2 parser offers an improved approach to gather metrics from JSON but still supports all the features of the original JSON parser (which will be referred to as json_v1 to avoid confusion). Therefore migration should be possible for all configurations, and in the process hopefully improve on the old configuration. This section will highlight the main differences between the two parsers and provide examples of converting from a json_v1 config to a json_v2 config.

### More verbose but clearer Telegraf configuration

The configuration no longer follows the format of needing a prepended `json_` for the config key, but instead uses TOML sub-tables. The goal with this change is to provide an easier to read configuration file that you can clearly see the distinct difference between the plugin config and the parser config. While this does lead to a more verbose config file, the improved ability to write/read the file should make up for this.

### More control over the field type

When using json_v1 you were limited to using `json_string_fields` to specify a field as a string. In json_v2 you can now specify any type that the influx line protocol supports for a field, even converting between types if possible. Read the section [#Types](#Types) for more details.

### Independent GJSON queries

A limitation with the json_v1 parser was that you could only provide a single GJSON query, and you could only use the returned data from that query to build your line protocol. The json_v2 parser lets you provide multiple independent GJSON queries so that you can gather the timestamp, measurement name, and field/tags from anywhere in the JSON.

### Control over the resulting tag/fields names

The json_v2 parser provides you the ability to rename all tags/fields without requiring a post-processing step like the json_v1 parser. You also have access to the setting `disable_prepend_keys` which you can enable to automatically provide cleaner field/tag names that don't include the parent keys prependend when gathering from nested arrays/objects.

### Example 1

This example was taken from the issue [#9376](https://github.com/influxdata/telegraf/issues/9376). Given the following JSON:

```json
{
  "timestamp": 1623786169180,
  "values": [
    {
      "id": "GET_S19_GATEHOST.VGF120B.VGF120_D435.Variables_HMI.Host_IdPart",
      "v": 630,
      "q": true,
      "t": 1623785875134
    },
    {
      "id": "GET_S19_GATEHOST.VGF120B.VGF120_D435.Variables_HMI.Host_MSN",
      "v": "MSN0573_CG10563",
      "q": true,
      "t": 1623784689919
    }
  ]
}
```

Parsing it with json_v1 would look like:

```toml
[[inputs.mqtt_consumer]]
  # ....
  data_format = "json"
  json_query = "values"

  json_string_fields= ["v", "q"]
  tag_keys = ["id"]
  json_time_key = "t"
  json_time_format = "unix_ms"
```

Parsing it with json_v2 would look like:

```toml
[[inputs.mqtt_consumer]]
    # ....
    data_format = "json_v2"
    [[inputs.file.json_v2]]
        timestamp_path = "timestamp"
        timestamp_format = "unix_ms"
        [[inputs.file.json_v2.object]]
            path = "values"
            tags = ["id"]
```

Both leading to the expected output:

```
file,id=GET_S19_GATEHOST.VGF120B.VGF120_D435.Variables_HMI.Host_IdPart v=630,q=true,t=1623785875134 162378616918000000
file,id=GET_S19_GATEHOST.VGF120B.VGF120_D435.Variables_HMI.Host_MSN v="MSN0573_CG10563",q=true,t=1623785875134,t=1623784689919 1623786169180000000
```
