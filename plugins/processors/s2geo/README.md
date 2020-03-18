# S2 Geo Processor Plugin

Use the `s2geo` processor to add tag with S2 cell ID token of specified [cell level][cell levels].
The tag is used in `experimental/geo` Flux package functions.
The `lat` and `lon` fields values should contain WGS-84 coordinates in decimal degrees.

### Configuration

```toml
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

### Example

```diff
- mta,area=llir,id=GO505_20_2704,status=1 lat=40.878738,lon=-72.517572 1560540094
+ mta,area=llir,id=GO505_20_2704,status=1,s2_cell_id=89e8ed4 lat=40.878738,lon=-72.517572 1560540094
```

[cell levels]: https://s2geometry.io/resources/s2cell_statistics.html
