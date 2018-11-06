# Newrelic

The `newrelic` output data format converts metrics into New Relic Insighs events.

### Configuration

```toml
[[outputs.newrelic]]
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "newrelic"

```

### Examples:

Standard form:
```newrelic
{
    "eventType":"disk",
    "free":63412539392,
    "inodes_free":9223372036853311651,
    "inodes_total":9223372036854775807,
    "inodes_used":1464156,
    "total":499963170816,
    "used":433700421632,
    "used_percent":87.24383704231391,
    "fstype":"nullfs",
    "host":"computer.local",
    "mode":"ro",
    "path":"/private/var/folders/hb/8dt_wwzj0hqbm0jstrn90_f41d1q9b/T/AppTranslocation/00EF667E-92ED-4B7F-825D-F207148CBD79",
    "timestamp":1541459740
}
```
