# JSON Parser - Version 2

This parser takes valid JSON input and turns it into metrics. The goals for this parser is to gracefully handle arrays/objects and provide more flexbility in gathering tags/fields. The query syntax supported is still [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md), but with the new configuration options you will have an easier time gathering metrics.

## Configuration

By setting the data_format to `json_v2` this parser will be used. You can then define what fields and tags you want by defining sub-tables such as so:

```toml
[[inputs.file]]
files = ["input.json"]
data_format = "json_v2"
        [[inputs.file.json_v2]]
            [[inputs.file.json_v2.uniform_collection]]
                query = "books.#.characters"
            [[inputs.file.json_v2.object_selection]]
                query = "books"
```

The following keys can be set for the root configuration section:

* **measurement_name_query (OPTIONAL)**: You can define a query with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md) to set a measurement name from the JSON input. The query must return a single data value or it will use the default measurement name. This query is completely independent from the queries in `uniform_collection` or `object_Selection`.

The query configuration options for this parser has been separated into two different sections, `uniform_collection` and `object_selection`.

### uniform_collection

With the configuration section `uniform_collection`, you can gather a collection of metrics from "uniform" data. The definition of "uniform" data is data which all share the same type and name. This can either be a single value or an array of values, but can't be an object (any objects found will be ignored, a debug log will be created if this happens). Any valid JSON type is supported (string,int,bool) and if possible can be converted to another type (string,int,float,bool), read the section [Types](#Types) to see what type conversion is supported.

The following keys can be set for `uniform_collection`:

* **query (REQUIRED)**: You must define the path query that gathers the object with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md).
* **name (OPTIONAL)**: You can define a string value to set the field name. If not defined it will use the trailing word from the provided query.
* **value_type (OPTIONAL)**: You can define a string value to set the desired type (int, bool, string, float). If not defined it won't enforce a type and default to using the original type defined in the JSON (bool, float64, or string).
* **set_type (OPTIONAL)**: Can be the string "field" or "tag"

### object_selection

With the configuration section `object_selection`, you can gather metrics from objects. The data doesn't have to be "uniform" and can contain multiple types.

The following keys can be set for `object_selection`:

* **query (REQUIRED)**: You must define the path query that gathers the object with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md)
* **included_keys (OPTIONAL)**: You can define a list of key's that should be the only data included in the metric, by default it will include everything.
* **ignored_keys (OPTIONAL)**: You can define a list of key's that should be ignored, by default it won't ignore anything.
* **names (OPTIONAL)**: You can define a key-value map, to associate object keys with the desired field name. If not defined it will use the JSON key as the field name by default.
* **value_types (OPTIONAL)**: You can define a key-value map, to associate object keys with the desired type (int, bool, string, float). If not defined it won't enforce a type and default to using the original type defined in the JSON (bool, float64, or string).
* **tag_list (OPTIONAL)**: Can be the string "field" or "tag"

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
            [[inputs.file.json_v2.object_selection]]
            query = "book"
            tag_list = ["title"]
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

* `int`, floats or strings (with valid numbers) can be converted to a int.
* `string`, any data can be formatted as a string.
* `float`, string values (with valid numbers) or integers can be converted to a float.
* `bool`, the string values "true" or "false" (regardless of capitalization) or the integer values `0` or `1`  can be turned to a bool.
