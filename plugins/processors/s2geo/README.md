# S2 Geo Processor Plugin

Use the `s2geo` processor to add tag with S2 cell ID token of specified [cell
level][cell levels].  The tag is used in `experimental/geo` Flux package
functions.  The `lat` and `lon` fields values should contain WGS-84 coordinates
in decimal degrees.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md

## Configuration

```toml @sample.conf
# Add the S2 Cell ID as a tag based on latitude and longitude fields
[[processors.s2geo]]
  ## The name of the lat and lon fields containing WGS-84 latitude and
  ## longitude in decimal degrees.
  # lat_field = "lat"
  # lon_field = "lon"

  ## New tag to create
  # tag_key = "s2_cell_id"

  ## Cell level (see https://s2geometry.io/resources/s2cell_statistics.html)
  # cell_level = 9
```

## Example

```diff
- mta,area=llir,id=GO505_20_2704,status=1 lat=40.878738,lon=-72.517572 1560540094
+ mta,area=llir,id=GO505_20_2704,status=1,s2_cell_id=89e8ed4 lat=40.878738,lon=-72.517572 1560540094
```

[cell levels]: https://s2geometry.io/resources/s2cell_statistics.html
