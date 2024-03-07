# geo_apiip Processor Plugin

The apiip processor plugin enhance each metric passing through it with the location tags received from external API [https://apiip.net/](https://apiip.net/documentation).

`api_key` is required to use the API.
By default, external IP uses origin IP address of the metric. You can
override it by taking the IP address from a tag using `ip_tag` option
or hardcode ip with `ip` option.

By default, the plugin will add the following tags to the metric:

- `region` - continent name (e.g. *Europe*)
- `country` - country code (e.g. *CY*)
- `city` - city name (e.g. *Nicosia*)

Tag names can be changed using `region_tag`, `country_tag` and `city_tag`
options.

In case of external API failure, the plugin will use non-empty
`default_region`, `default_country` and `default_city` options as fallback to
set the tags.

Select the metrics to modify using the standard [metric
filtering](../../../docs/CONFIGURATION.md#metric-filtering) options.

Values of  already present *tags* with conflicting keys will be overwritten. Absent *tags* will be created.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Enhance metrics with location tags received from apiip
[[processors.geo_apiip]]
  ## All modifications on inputs and aggregators can be overridden:
  # api_key = "XXX" # Required
  # ip_tag = "tag containing ip" # if empty - origin ip will be used. overridden by ip
  # ip = "" # hardcode ip
  # update_interval = "5m" # how often to update ip database
  # region_tag = "region"
  # country_tag = "country"
  # city_tag = "city"
  # default_region = "Europe"
  # default_country = "NL"
  # default_city = "Amsterdam"
```
