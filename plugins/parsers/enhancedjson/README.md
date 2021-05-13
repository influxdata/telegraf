# Enhanced JSON Parser

THIS PARSER IS STILL A WORK IN PROGRESS

This new JSON Parser is parses JSON into metric fields using [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md). This parser is designed to be more flexible then the previous implementation. The reason this is a separate parser is ensure backwards compatibility and allow users to migrate overtime to this new parser.

## Configuration

By setting the data_format to `enhancedjson` this parser will be used. You can then define what fields and tags you want by defining sub-sections such as so:

```toml
[[inputs.file]]
files = ["input.json"]
data_format = "enhancedjson"
        [[inputs.file.enhancedjson]]
            [[inputs.file.enhancedjson.basic_fields]]
                query = "books.#.characters"
            [[inputs.file.enhancedjson.object_fields]]
                query = "books"
```

The configuration options for this parser has been separated into two different types, `basic_fields` and `object_fields`. These types are described in more detail below.

### basic_fields

To explicitly gather fields from basic types (string, int bool) and non-object array's (supports nested non-object array's), you need to use the configuration `basic_fields`.
The following keys can be set for `basic_fields`:

* **query (REQUIRED)**: You must define the path query that gathers the object with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md)
* **name (OPTIONAL)**: You can define a string value to set the field name. If not defined it will use the trailing word from the provided query.
* **type (OPTIONAL)**: You can define a string value to set the desired type (int, bool, string, float). If not defined it won't enforce a type and default to using the original type defined in the JSON (bool, float64, or string).
* **ignore_objects (OPTIONAL)**: By default an error will be thrown if a object is found. You can set this key to `true` or `false`, if `true` then if an object is encountered it will ignore it and not throw an error.

### object_fields

To explicitly gather fields from objects (supports nested arrays/objects and basic types), you need to use the configuration `object_fields`.
The following keys can be set for `object_fields`:

* **query (REQUIRED)**: You must define the path query that gathers the object with [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/v1.7.5/SYNTAX.md)
* **include_list (OPTIONAL)**: You can define a list of key's that should be the only data included in the metric, by default it will include everything.
* **ignore_list (OPTIONAL)**: You can define a list of key's that should be ignored, by default it won't ignore anything.
* **name_map (OPTIONAL)**: You can define a key-value map, to associate object keys with the desired field name. If not defined it will use the JSON key as the field name by default.
* **type_map (OPTIONAL)**: You can define a key-value map, to associate object keys with the desired type (int, bool, string, float). If not defined it won't enforce a type and default to using the original type defined in the JSON (bool, float64, or string).

### Separate metrics

You can define multiple "enhancedjson" under a plugin such as so:

```toml
[[inputs.file]]
files = ["input.json"]
data_format = "enhancedjson"
        [[inputs.file.enhancedjson]]
            [[inputs.file.enhancedjson.basic_fields]]
                query = "books.#.characters"
        [[inputs.file.enhancedjson]]
            [[inputs.file.enhancedjson.object_fields]]
                query = "books"
```

This will ensure that the queried data isn't combined and will be outputted separately. Otherwise, all `basic_fields` and `object_fields` under a `enhancedjson` subsection will be merged into a single metric.

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
data_format = "enhancedjson"
        [[inputs.file.enhancedjson]]
            [[inputs.file.enhancedjson.object_fields]]
                query = "books"
```

Expected metrics:

```
file,host=test title="Sword of Honour",author="Evelyn Waugh" 1596294243000000000
file,host=test title="The Lord of the Rings",author="J. R. R. Tolkien",characters="Bilbo" 1596294243000000000
file,host=test title="The Lord of the Rings",author="J. R. R. Tolkien",characters="Frodo" 1596294243000000000
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
