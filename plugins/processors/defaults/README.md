# Defaults Processor Plugin

This plugin allows to specify default values for fields and tags for cases
where the tag or field does not exist or has an empty value.

⭐ Telegraf v1.15.0
🏷️ transformation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
## Set default fields and tags on your metric(s) when they are nil or empty
[[processors.defaults]]
  ## Ensures a set of fields or tags always exists on your metric(s) with their
  ## respective default value.
  ## For any given field/tag pair (key = default), if it's not set, a field/tag
  ## is set on the metric with the specified default.
  ##
  ## A field is considered not set if it is nil on the incoming metric;
  ## or it is not nil but its value is an empty string or is a string
  ## of one or more spaces.
  ##   <target-field> = <value>
  [processors.defaults.fields]
    field_1 = "bar"
    time_idle = 0
    is_error = true
  ## A tag is considered not set if it is nil on the incoming metric;
  ## or it is not nil but it is empty string or a string of one or
  ## more spaces.
  ## <target-tag> = <value>
  [processors.defaults.tags]
    tag_1 = "foo"
```

## Example

Ensure a _status\_code_ field with _N/A_ is inserted in the metric when one is
not set in the metric by default:

```toml
[[processors.defaults]]
  [processors.defaults.fields]
    status_code = "N/A"
```

```diff
- lb,http_method=GET cache_status=HIT,latency=230
+ lb,http_method=GET cache_status=HIT,latency=230,status_code="N/A"
```

Ensure an empty string gets replaced by a default:

```diff
- lb,http_method=GET cache_status=HIT,latency=230,status_code=""
+ lb,http_method=GET cache_status=HIT,latency=230,status_code="N/A"
```
