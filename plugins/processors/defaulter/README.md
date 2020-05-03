# Defaulter Processor

The *Defaulter* processor allows you to ensure certain fields will always exist with a specified default value on your metric(s).

There are two cases where this processor will insert a configured default field.

1. The field is nil on the incoming metric
1. The field is present, but its value is empty.
    
    This processor considers the field empty if its value is the empty string or a single space character.

### Configuration
```toml
## Set default fields on your metric(s) when they are nil or empty
[[processors.defaulter]]

## This table determines what fields will be inserted in your metric(s)
  [processors.defaulter.fields]
    field_1 = "bar"
    time_idle = 0
    is_error = true
```

### Example
Ensure a _status\_code_ field with _N/A_ is inserted in the metric when one it's not set in the metric be default:

```toml
[[processors.defaulter]]
  [processors.defaulter.fields]
    status_code = "N/A"
```

```diff
- lb,http_method=GET cache_status=HIT latency=230
+ lb,http_method=GET cache_status latency=230 status_code="N/A"
```

Ensure an empty string gets replaced by a default:

```diff
- lb,http_method=GET cache_status=HIT latency=230 status_code=""
+ lb,http_method=GET cache_status latency=230 status_code="N/A"
```