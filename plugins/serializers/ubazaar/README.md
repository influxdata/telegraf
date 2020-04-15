# uBazaar

The `ubazaar` output data format converts metrics into JSON documents valid for
sending to the Unity uBazaar metered billing platform.

### Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "ubazaar"
```

### Examples:

```json
{
	"eventID": "c3d2d292-e216-491c-9ee4-518057b0e8f6",
	"serviceCustomerID": "mp-user-1",
	"service": "MP",
	"unitOfMeasure": "MP-network-gb",
	"quantity": 1.2345,
	"startTime": "2020-04-14-10T15:55:00Z",
	"endTime": "2020-04-14-10T15:55:00Z"
}
```
