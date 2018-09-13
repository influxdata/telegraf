# Dropwizard

The `dropwizard` data format can parse the [JSON Dropwizard][dropwizard] representation of a single dropwizard metric registry. By default, tags are parsed from metric names as if they were actual influxdb line protocol keys (`measurement<,tag_set>`) which can be overriden by defining a custom [template pattern][templates]. All field value types are supported, `string`, `number` and `boolean`.

[templates]: /docs/TEMPLATE_PATTERN.md
[dropwizard]: http://metrics.dropwizard.io/3.1.0/manual/json/

### Configuration

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "dropwizard"

  ## Used by the templating engine to join matched values when cardinality is > 1
  separator = "_"

  ## Each template line requires a template pattern. It can have an optional
  ## filter before the template and separated by spaces. It can also have optional extra
  ## tags following the template. Multiple tags should be separated by commas and no spaces
  ## similar to the line protocol format. There can be only one default template.
  ## Templates support below format:
  ## 1. filter + template
  ## 2. filter + template + extra tag(s)
  ## 3. filter + template with field key
  ## 4. default template
  ## By providing an empty template array, templating is disabled and measurements are parsed as influxdb line protocol keys (measurement<,tag_set>)
  templates = []

  ## You may use an appropriate [gjson path](https://github.com/tidwall/gjson#path-syntax)
  ## to locate the metric registry within the JSON document
  # dropwizard_metric_registry_path = "metrics"

  ## You may use an appropriate [gjson path](https://github.com/tidwall/gjson#path-syntax)
  ## to locate the default time of the measurements within the JSON document
  # dropwizard_time_path = "time"
  # dropwizard_time_format = "2006-01-02T15:04:05Z07:00"

  ## You may use an appropriate [gjson path](https://github.com/tidwall/gjson#path-syntax)
  ## to locate the tags map within the JSON document
  # dropwizard_tags_path = "tags"

  ## You may even use tag paths per tag
  # [inputs.exec.dropwizard_tag_paths]
  #   tag1 = "tags.tag1"
  #   tag2 = "tags.tag2"
```


### Examples

A typical JSON of a dropwizard metric registry:

```json
{
	"version": "3.0.0",
	"counters" : {
		"measurement,tag1=green" : {
			"count" : 1
		}
	},
	"meters" : {
		"measurement" : {
			"count" : 1,
			"m15_rate" : 1.0,
			"m1_rate" : 1.0,
			"m5_rate" : 1.0,
			"mean_rate" : 1.0,
			"units" : "events/second"
		}
	},
	"gauges" : {
		"measurement" : {
			"value" : 1
		}
	},
	"histograms" : {
		"measurement" : {
			"count" : 1,
			"max" : 1.0,
			"mean" : 1.0,
			"min" : 1.0,
			"p50" : 1.0,
			"p75" : 1.0,
			"p95" : 1.0,
			"p98" : 1.0,
			"p99" : 1.0,
			"p999" : 1.0,
			"stddev" : 1.0
		}
	},
	"timers" : {
		"measurement" : {
			"count" : 1,
			"max" : 1.0,
			"mean" : 1.0,
			"min" : 1.0,
			"p50" : 1.0,
			"p75" : 1.0,
			"p95" : 1.0,
			"p98" : 1.0,
			"p99" : 1.0,
			"p999" : 1.0,
			"stddev" : 1.0,
			"m15_rate" : 1.0,
			"m1_rate" : 1.0,
			"m5_rate" : 1.0,
			"mean_rate" : 1.0,
			"duration_units" : "seconds",
			"rate_units" : "calls/second"
		}
	}
}
```

Would get translated into 4 different measurements:

```
measurement,metric_type=counter,tag1=green count=1
measurement,metric_type=meter count=1,m15_rate=1.0,m1_rate=1.0,m5_rate=1.0,mean_rate=1.0
measurement,metric_type=gauge value=1
measurement,metric_type=histogram count=1,max=1.0,mean=1.0,min=1.0,p50=1.0,p75=1.0,p95=1.0,p98=1.0,p99=1.0,p999=1.0
measurement,metric_type=timer count=1,max=1.0,mean=1.0,min=1.0,p50=1.0,p75=1.0,p95=1.0,p98=1.0,p99=1.0,p999=1.0,stddev=1.0,m15_rate=1.0,m1_rate=1.0,m5_rate=1.0,mean_rate=1.0
```

You may also parse a dropwizard registry from any JSON document which contains a dropwizard registry in some inner field.
Eg. to parse the following JSON document:

```json
{
	"time" : "2017-02-22T14:33:03.662+02:00",
	"tags" : {
		"tag1" : "green",
		"tag2" : "yellow"
	},
	"metrics" : {
		"counters" : 	{
			"measurement" : {
				"count" : 1
			}
		},
		"meters" : {},
		"gauges" : {},
		"histograms" : {},
		"timers" : {}
	}
}
```
and translate it into:

```
measurement,metric_type=counter,tag1=green,tag2=yellow count=1 1487766783662000000
```

you simply need to use the following additional configuration properties:

```toml
dropwizard_metric_registry_path = "metrics"
dropwizard_time_path = "time"
dropwizard_time_format = "2006-01-02T15:04:05Z07:00"
dropwizard_tags_path = "tags"
## tag paths per tag are supported too, eg.
#[inputs.yourinput.dropwizard_tag_paths]
#  tag1 = "tags.tag1"
#  tag2 = "tags.tag2"
```
