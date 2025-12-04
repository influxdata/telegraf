# Enum Processor Plugin

This plugin allows the mapping of field or tag values according to the
configured enumeration. The main use-case is to rewrite numerical values into
human-readable values or vice versa. Default mappings can be configured to be
used for all remaining values.

⭐ Telegraf v1.8.0
🏷️ transformation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Map enum values according to given table.
[[processors.enum]]
  [[processors.enum.mapping]]
    ## Names of the fields to map. Globs accepted.
    fields = ["status"]

    ## Name of the tags to map. Globs accepted.
    # tags = ["status"]

    ## Destination tag or field to be used for the mapped value.  By default the
    ## source tag or field is used, overwriting the original value.
    dest = "status_code"

    ## Default value to be used for all values not contained in the mapping
    ## table.  When unset and no match is found, the original field will remain
    ## unmodified and the destination tag or field will not be created.
    # default = 0

    ## Table of mappings
    [processors.enum.mapping.value_mappings]
      green = 1
      amber = 2
      red = 3
```

## Example

```diff
- xyzzy status="green" 1502489900000000000
+ xyzzy status="green",status_code=1i 1502489900000000000
```

With unknown value and no default set:

```diff
- xyzzy status="black" 1502489900000000000
+ xyzzy status="black" 1502489900000000000
```
