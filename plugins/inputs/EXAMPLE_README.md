# Example Input Plugin

The example plugin gathers metrics about example things.  This description
explains at a high level what the plugin does and provides links to where
additional information can be found.

### Configuration:

```toml
# Description
[[inputs.example]]
  # SampleConfig
```

### Measurements & Fields:

Here you should add an optional description and links to where the user can
get more information about the measurements.

- measurement1
    - field1 (type, unit)
    - field2 (float, percent)
- measurement2
    - field3 (integer, bytes)

### Tags:

- All measurements have the following tags:
    - tag1 (optional description)
    - tag2
- measurement2 has the following tags:
    - tag3

### Sample Queries:

This section should contain some useful InfluxDB queries that can be used to
get started with the plugin or to generate dashboards.  For each query listed,
describe at a high level what data is returned.

Get the max, mean, and min for the measurement in the last hour:
```
SELECT max(field1), mean(field1), min(field1) FROM measurement1 WHERE tag1=bar AND time > now() - 1h GROUP BY tag
```

### Example Output:

```
measurement1,tag1=foo,tag2=bar field1=1i,field2=2.1 1453831884664956455
measurement2,tag1=foo,tag2=bar,tag3=baz field3=1i 1453831884664956455
```
