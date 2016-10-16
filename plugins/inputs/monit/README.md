# Monit Plugin

The monit plugin gathers metrics and status information about local processes managed and watched over by Monit.

### Configuration:

```toml
# Description
[[inputs.example]]
  # SampleConfig
```

### Measurements & Fields:

<optional description>

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

These are some useful queries (to generate dashboards or other) to run against data from this plugin:

```
SELECT max(field1), mean(field1), min(field1) FROM measurement1 WHERE tag1=bar AND time > now() - 1h GROUP BY tag
```

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter example -test
measurement1,tag1=foo,tag2=bar field1=1i,field2=2.1 1453831884664956455
measurement2,tag1=foo,tag2=bar,tag3=baz field3=1i 1453831884664956455
```
