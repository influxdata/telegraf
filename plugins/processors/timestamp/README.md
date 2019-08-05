# Timestamp Processor Plugin

Use the `timestamp` processor to add a unix nano timestamp to the metric.

This can be used to mimic logs from syslog that you'd want to display in Chronograf.

### Configuration

```toml
[[processors.timestamp]]
  ## New tag to create
  field_key = "timestamp"
```

### Example

```diff
- syslog,appname=myapp,facility=user,hostname=test,severity=notice message="notice msg",severity_code=5i,version=1i,facility_code=1i 1564997347582799644
+ syslog,appname=myapp,facility=user,hostname=test,severity=notice message="notice msg",severity_code=5i,version=1i,facility_code=1i,timestamp=1564997347582799644i 1564997347582799644
```
