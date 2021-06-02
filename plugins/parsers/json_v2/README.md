# JSON Parser - Version 2

This parser takes valid JSON input and turns it into metrics. The query syntax supported is [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md). You can find multiple examples under the `testdata` folder.

## Configuration

You configure this parser by describing the metric you want by defining the fields and tags from the input. The configuration is divided into config sub-tables labeled `metric`, within which you can define multiple `field`, `tag`, and `object` config tables. In the example below you can see all the possible configuration keys you can define for each config table. In the sections that follow these configuration keys are defined in more detail.

**Example configuration:**

```toml
data_format = "json_v2"
    [[xxx.xxx.json_v2]]
        [[xxx.xxx.json_v2.metric]]
            measurement_name =
            measurement_name_path =
            timestamp_path =
            timestamp_format =
            timestamp_timezone =
            [[field]]
                path =
                rename =
                type =
            [[tag]]
                path =
                rename =
            [[object]]
                path = # (REQUIRED) A valid GJSON path
                disable_nesting = # (OPTIONAL) Set to true to prevent going into sub-objects/arrays
                disable_flattened_names = # (OPTIONAL) Set to true to prevent resulting nested names to be flattened in the result **NOTE**: duplicate names can overwrite each other when this is enabled
                excluded_keys = [] # (OPTIONAL) You can define json keys to be excluded in the metric, use flattened names for nested results
                tags = [] # (OPTIONAL) You can define json keys to be set as tags instead of fields
                [renames] # (OPTIONAL) A table matching the json key with the desired name (oppossed to defaulting to using the key), use flattened names for nested results
                    key = new_name
                [fields] # (OPTIONAL) A table matching the json key with the desired type (int,string,bool,float)
                    key = type
```
---
### `metric` config options

A `metric` config table can be used to describe how to parse metrics from JSON. This configuration can return multiple metrics when parsing an array, but eac There are a list of root level config options that you can set that will be

* **measurement_name (OPTIONAL)**:  Will set the measurement name to the provided string.
* **measurement_name_path (OPTIONAL)**: You can define a query with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md) to set a measurement name from the JSON input. The query must return a single data value or it will use the default measurement name. This takes precedence over `measurement_name`.
* **timestamp_path (OPTIONAL)**: You can define a query with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md) to set a timestamp from the JSON input. The query must return a single data value or it will default to the current time.
* **timestamp_format (OPTIONAL, but REQUIRED when timestamp_query is defined**: ust be set to `unix`, `unix_ms`, `unix_us`, `unix_ns`, or
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

With the configuration section `object_selection`, you can gather metrics from objects. The data doesn't have to be "uniform" and can contain multiple types.

The following keys can be set for `object_selection`:

* **path (REQUIRED)**: You must define the path query that gathers the object with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md)
* **disable_prepend_keys (OPTIONAL)**: Set to true to prevent resulting nested data to contain the parent key prepended to its key **NOTE**: duplicate names can overwrite each other when this is enabled
* **excluded_keys (OPTIONAL)**: You can define json keys to be excluded in the metric, use flattened names for nested results
* **tags (OPTIONAL)**: You can define json keys to be set as tags instead of fields, if you define a key that is an array or object then all nested values will become a tag
* **renames (OPTIONAL)**: A table matching the json key with the desired name (oppossed to defaulting to using the key), use names that include the prepended keys of its parent keys for nested results
* **fields (OPTIONAL)**: A table matching the json key with the desired type (int,string,bool,float), if you define a key that is an array or object then all nested values will become that type

### Separate metrics

You can define multiple "json_v2" under a plugin such as so:

```toml
[[inputs.file]]
files = ["input.json"]
data_format = "json_v2"
        [[inputs.file.json_v2]]
            [[inputs.file.json_v2.uniform_collection]]
                query = "books.#.characters"
        [[inputs.file.json_v2]]
            [[inputs.file.json_v2.object_selection]]
                query = "books"
```

This will ensure that the queried data isn't combined and will be outputted separately. Otherwise, all `uniform_collection` and `object_selection` under a `json_v2` subsections will be merged.

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
