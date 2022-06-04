# EXTR

The `extr` output data format converts metrics into JSON documents, performing the following operatins on batched metrics:
   - Combines sequential metrics matching name, tags, and timestamps into a single JSON metric, combining the fields of each metric into an array of fields.
   - Groups metric fields appended with _min, _max, _avg
        usage_min=1,usage_max=100,usage_avg-50
        --> "usage":{"avg":50,"max":100,"min":1}
   - Groups metric fields appended with _key.
       ifIndex_key=1, name_key="1:2"
       --> "key":{ifIndex:1, name:"1:2:}
   - Groups like metric names into a toplevel map. Name of group is same as name, but with first char lowercase
       "fanStats" :[{grouped_FanStats_Metric1}, {grouped_FanStats_Metric2} ]

*extr serializer batches metrics by default.
   
### Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "extr"
  use_batch_format = true

  ## The resolution to use for the metric timestamp.  Must be a duration string
  ## such as "1ns", "1us", "1ms", "10ms", "1s".  Durations are truncated to
  ## the power of 10 less than the specified units.
  json_timestamp_units = "1s"


[[outputs.http]]
  url = "http://10.139.101.72:9443/telegraf/rest/v1"
  method = "POST"

  data_format = "extr"
  flush_interval = "2s"

  [outputs.http.headers]
  Content-Type = "application/json; charset=utf-8"

```

### Examples:

The following Telegraf batched metrics
   
```text
CpuStats,serialnum=XYZ-1234 core_key=0i,usage_min=35.1,usage_max=99.1,usage_avg=35.1
CpuStats,serialnum=XYZ-1234 core_key=1i,usage_min=50.1,usage_max=88.1,usage_avg=51.1
FanStats,serialnum=XYZ-1234 slot_key=1i,tray_key=2i,fan_key=10i,rpm_min=4101,rpm_max=5001,rpm_avg=4201,pwm_min=31,pwm_max=41,pwm_avg=31
FanStats,serialnum=XYZ-1234 slot_key=1i,tray_key=2i,fan_key=11i,rpm_min=4001,rpm_max=4991,rpm_avg=4001,pwm_min=41,pwm_max=51,pwm_avg=41
FanStats,serialnum=XYZ-1234 slot_key=2i,tray_key=3i,fan_key=9i,rpm_min=2101,rpm_max=3211,rpm_avg=2201,pwm_min=11,pwm_max=41,pwm_avg=11
CpuStats,serialnum=XYZ-1234 core_key=0i,usage_min=10.2,usage_max=91.2,usage_avg=44.2
CpuStats,serialnum=XYZ-1234 core_key=1i,usage_min=22.2,usage_max=89.2,usage_avg=41.2
CpuStats,serialnum=XYZ-1234 core_key=2i,usage_min=33.2,usage_max=79.2,usage_avg=47.2
FanStats,serialnum=XYZ-1234 slot_key=1i,tray_key=2i,fan_key=10i,rpm_min=4112,rpm_max=5012,rpm_avg=4212,pwm_min=32,pwm_max=52,pwm_avg=32
FanStats,serialnum=XYZ-1234 slot_key=1i,tray_key=2i,fan_key=11i,rpm_min=5002,rpm_max=5092,rpm_avg=4102,pwm_min=52,pwm_max=62,pwm_avg=52
```

will serialize into the following extr JSON ouput
   
```json
{
  "cpuStats": [
    {
      "device": {
        "serialnum": "XYZ-1234"
      },
      "items": [
        {
          "keys": {
            "core": 0
          },
          "usage": {
            "avg": 35.1,
            "max": 99.1,
            "min": 35.1
          }
        },
        {
          "keys": {
            "core": 1
          },
          "usage": {
            "avg": 51.1,
            "max": 88.1,
            "min": 50.1
          }
        }
      ],
      "name": "CpuStats",
      "ts": 1654306730
    },
    {
      "device": {
        "serialnum": "XYZ-1234"
      },
      "items": [
        {
          "keys": {
            "core": 0
          },
          "usage": {
            "avg": 44.2,
            "max": 91.2,
            "min": 10.2
          }
        },
        {
          "keys": {
            "core": 1
          },
          "usage": {
            "avg": 41.2,
            "max": 89.2,
            "min": 22.2
          }
        },
        {
          "keys": {
            "core": 2
          },
          "usage": {
            "avg": 47.2,
            "max": 79.2,
            "min": 33.2
          }
        }
      ],
      "name": "CpuStats",
      "ts": 1654306730
    }
  ],
  "fanStats": [
    {
      "device": {
        "serialnum": "XYZ-1234"
      },
      "items": [
        {
          "keys": {
            "fan": 10,
            "slot": 1,
            "tray": 2
          },
          "pwm": {
            "avg": 31,
            "max": 41,
            "min": 31
          },
          "rpm": {
            "avg": 4201,
            "max": 5001,
            "min": 4101
          }
        },
        {
          "keys": {
            "fan": 11,
            "slot": 1,
            "tray": 2
          },
          "pwm": {
            "avg": 41,
            "max": 51,
            "min": 41
          },
          "rpm": {
            "avg": 4001,
            "max": 4991,
            "min": 4001
          }
        },
        {
          "keys": {
            "fan": 9,
            "slot": 2,
            "tray": 3
          },
          "pwm": {
            "avg": 11,
            "max": 41,
            "min": 11
          },
          "rpm": {
            "avg": 2201,
            "max": 3211,
            "min": 2101
          }
        }
      ],
      "name": "FanStats",
      "ts": 1654306730
    },
    {
      "device": {
        "serialnum": "XYZ-1234"
      },
      "items": [
        {
          "keys": {
            "fan": 10,
            "slot": 1,
            "tray": 2
          },
          "pwm": {
            "avg": 32,
            "max": 52,
            "min": 32
          },
          "rpm": {
            "avg": 4212,
            "max": 5012,
            "min": 4112
          }
        },
        {
          "keys": {
            "fan": 11,
            "slot": 1,
            "tray": 2
          },
          "pwm": {
            "avg": 52,
            "max": 62,
            "min": 52
          },
          "rpm": {
            "avg": 4102,
            "max": 5092,
            "min": 5002
          }
        }
      ],
      "name": "FanStats",
      "ts": 1654306730
    }
  ]
}
```
