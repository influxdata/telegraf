# Enhanced JSON Parser

THIS PARSER IS STILL A WORK IN PROGRESS

This new JSON Parser is parses JSON into metric fields using [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md). This parser is designed to be more flexible then the previous implementation. The reason this is a separate parser is ensure backwards compatibility and allow users to migrate overtime to this new parser.

## Configuration

By setting the data_format to `json_v2` this parser will be used. You can then define what fields and tags you want by defining sub-sections such as so:

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

The configuration options for this parser has been separated into two different types, `uniform_collection` and `object_selection`. These types are described in more detail below.

### uniform_collection

To explicitly gather fields from basic types (string, int bool) and non-object array's (supports nested non-object array's), you need to use the configuration `uniform_collection`. If the query you provides returns any objects (nested or directly) they will be ignored and not show up in the resulting metrics.
The following keys can be set for `uniform_collection`:

* **query (REQUIRED)**: You must define the path query that gathers the object with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md)
* **name (OPTIONAL)**: You can define a string value to set the field name. If not defined it will use the trailing word from the provided query.
* **value_type (OPTIONAL)**: You can define a string value to set the desired type (int, bool, string, float). If not defined it won't enforce a type and default to using the original type defined in the JSON (bool, float64, or string).
* **set_type (OPTIONAL)**: Can be the string "field" or "tag"

### object_selection

To explicitly gather fields from objects (supports nested arrays/objects and basic types), you need to use the configuration `object_selection`.
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

This will ensure that the queried data isn't combined and will be outputted separately. Otherwise, all `uniform_collection` and `object_selection` under a `json_v2` subsection will be merged into a single metric.

## Arrays and Objects

The following describes the high-level approach when parsing arrays and objects:

**Array**: Every element in an array is treated as a *separate* metric

**Object**: Every key/value in a object is treated as a *single* metric

When handling nested arrays and objects, these above rules continue to apply as the parser creates metrics. Below you can see an example of this behavior, with an input json containing an array of book objects that has a nested array of characters.

Example JSON:

```json
{
    "book": [
        {
            "title": "Sword of Honour",
            "author": "Evelyn Waugh"
        },
        {
            "title": "The Lord of the Rings",
            "author": "J. R. R. Tolkien",
            "characters": [
                "Bilbo",
                "Frodo"
            ]
        }
    ]
}
```

Example configuration:

```toml
[[inputs.file]]
files = ["input.json"]
data_format = "json_v2"
        [[inputs.file.json_v2]]
            [[inputs.file.json_v2.object_selection]]
                query = "books"
```

Expected metrics:

```
file,host=test title="Sword of Honour",author="Evelyn Waugh" 1596294243000000000
file,host=test title="The Lord of the Rings",author="J. R. R. Tolkien",characters="Bilbo",chapter=1 1596294243000000000
file,host=test title="The Lord of the Rings",author="J. R. R. Tolkien",characters="Bilbo",chapter=2 1596294243000000000
file,host=test title="The Lord of the Rings",author="J. R. R. Tolkien",characters="Frodo",chapter=1 1596294243000000000
file,host=test title="The Lord of the Rings",author="J. R. R. Tolkien",characters="Frodo",chapter=2 1596294243000000000
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
