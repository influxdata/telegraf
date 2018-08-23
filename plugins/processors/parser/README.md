# Parser Processor Plugin

This plugin parses defined fields containing the specified data format and
creates new metrics based on the contents of the field.

## Configuration
```toml
[[processors.parser]]
  ## The name of the fields whose value will be parsed.
  parse_fields = ["message"]

  ## If true, incoming metrics are not emitted.
  drop_original = false

  ## If set to override, emitted metrics will be merged by overriding the
  ## original metric using the newly parsed metrics.
  merge = "override"

  ## The dataformat to be read from files
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

### Example:

```toml
[[processors.parser]]
  parse_fields = ["message"]
  merge = "override"
  data_format = "logfmt"
```

**Input**:
```
syslog,appname=influxd,facility=daemon,hostname=http://influxdb.example.org\ (influxdb.example.org),severity=info facility_code=3i,message=" ts=2018-08-09T21:01:48.137963Z lvl=info msg=\"Executing query\" log_id=09p7QbOG000 service=query query=\"SHOW DATABASES\"",procid="6629",severity_code=6i,timestamp=1533848508138040000i,version=1i
```

**Output**:
```
syslog,appname=influxd,facility=daemon,hostname=http://influxdb.example.org\ (influxdb.example.org),severity=info facility_code=3i,log_id="09p7QbOG000",lvl="info",message=" ts=2018-08-09T21:01:48.137963Z lvl=info msg=\"Executing query\" log_id=09p7QbOG000 service=query query=\"SHOW DATABASES\"",msg="Executing query",procid="6629",query="SHOW DATABASES",service="query",severity_code=6i,timestamp=1533848508138040000i,ts="2018-08-09T21:01:48.137963Z",version=1i
```


