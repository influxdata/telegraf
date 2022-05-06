# Defaults Processor

The *Defaults* processor allows you to ensure certain fields will always exist with a specified default value on your metric(s).

There are three cases where this processor will insert a configured default field.

1. The field is nil on the incoming metric
1. The field is not nil, but its value is an empty string.
1. The field is not nil, but its value is a string of one or more empty spaces.

Telegraf minimum version: Telegraf 1.15.0

## Configuration

```toml
## Set default fields on your metric(s) when they are nil or empty
[[processors.defaults]]
  ## Ensures a set of fields always exists on your metric(s) with their
  ## respective default value.
  ## For any given field pair (key = default), if it's not set, a field
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
```

## Example

Ensure a _status\_code_ field with _N/A_ is inserted in the metric when one is not set in the metric by default:

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
