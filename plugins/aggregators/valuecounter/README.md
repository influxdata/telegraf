# Value Counter Aggregator Plugin

This plugin counts the occurrence of unique values in fields and emits the
counter once every `period` with the field-names being suffixed by the unique
value converted to `string`.

> [!NOTE]
> The fields to be counted must be configured using the `fields` setting,
> otherwise no field will be counted and no metric is emitted.

This plugin is useful to e.g. count the occurrances of HTTP status codes or
other categorical values in the defined `period`.

> [!IMPORTANT]
> Counting fields with a high number of potential values may produce a
> significant amounts of new fields and results in an increased memory usage.
> Take care to only count fields with a limited set of values.

⭐ Telegraf v1.8.0
🏷️ statistics
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Count the occurrence of values in fields.
[[aggregators.valuecounter]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  # period = "30s"

  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  # drop_original = false

  ## The fields for which the values will be counted
  fields = ["status"]
```

### Measurements & Fields

- measurement1
  - field_value1
  - field_value2

### Tags

No tags are applied by this aggregator.

## Example Output

Example for parsing a HTTP access log.

telegraf.conf:

```toml
[[inputs.logparser]]
  files = ["/tmp/tst.log"]
  [inputs.logparser.grok]
    patterns = ['%{DATA:url:tag} %{NUMBER:response:string}']
    measurement = "access"

[[aggregators.valuecounter]]
  namepass = ["access"]
  fields = ["response"]
```

/tmp/tst.log

```text
/some/path 200
/some/path 401
/some/path 200
```

```text
access,url=/some/path,path=/tmp/tst.log,host=localhost.localdomain response="200" 1511948755991487011
access,url=/some/path,path=/tmp/tst.log,host=localhost.localdomain response="401" 1511948755991522282
access,url=/some/path,path=/tmp/tst.log,host=localhost.localdomain response="200" 1511948755991531697
access,path=/tmp/tst.log,host=localhost.localdomain,url=/some/path response_200=2i,response_401=1i 1511948761000000000
```
