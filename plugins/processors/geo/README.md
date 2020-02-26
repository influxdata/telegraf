# Geo Processor Plugin

Use the `geo` processor to add tag with S2 cell ID token of specified [`cell level`][cell levels].

The tag is used in `experimental/geo` Flux package functions.

### Configuration

```toml
[[processors.geo]]
  ## The name of the lat and lon fields
  lat_field = "lat"
  lon_field = "lon"

  ## New tag to create
  tag_key = "_ci"

  ## Cell level (see https://s2geometry.io/resources/s2cell_statistics.html)
  cell_level = 11

  ## Log mismatches
  log_mismatches = false
```

### Example

```diff
- mta,area=llir,id=GO505_20_2704,status=1 lat=40.878738,lon=-72.517572 1560540094
+ mta,area=llir,id=GO505_20_2704,status=1,_ci=89e8ed4 lat=40.878738,lon=-72.517572 1560540094
```

[cell levels]: https://s2geometry.io/resources/s2cell_statistics.html
