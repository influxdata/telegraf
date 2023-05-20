# EXTR

The `extr` output data format converts metrics into JSON documents, performing the following operatins on batched metrics:

- Combines sequential metrics matching name, tags, and timestamps into a single JSON metric, combining the fields of each metric into an array of fields.

- Groups metric fields appended with min, max, avg or old, new
        usage_min=1,usage_max=100,usage_avg=50
        --> "usage":{"avg":50,"max":100,"min":1}
        ifAdminStatus_old="Up",ifAdminStatus_new="Down"
        --> "ifAdminStatus":{"old":"Up","new":"Down"}
        *You could also use old_ifAdminStatus="Up",new_ifAdminStatus="Down" to achieve the same result.

- Groups metric fields appended with _key to a keys group
        ifIndex_key=1, name_key="1:2"
        --> "keys":{ifIndex:1, name:"1:2"}
        *You could also use ifIndex_keys=1,name_keys="1:2"

- Groups metric fields appended with _tag to a tags group
        partNumber_tag="1647G-00129 800751-00-01", revision_tag="01"
        --> "tags":{partNumber:"1647G-00129 800751-00-01", revision:"01"}
        *You could also use partNumber_tags="1647G-00129 800751-00-01", revision_tags="01"

- Groups like metric names into a toplevel map. Name of group is same as name, but with first char lowercase
        "fanStats" :[{grouped_FanStats_Metric1}, {grouped_FanStats_Metric2} ]

- Creates nested JSON schema by parsing underscore "_" seperated field keys
        cpu1_subcore_core_key=2
        cpu2_subcore_core_key=5
        --> {"keys":{"core":{"subcore":{"cpu1":2, "cpu2":5}}}}
        usage_subcore_cpu1_min=21
        usage_subcore_cpu1_max=100
        usage_subcore_cpu1_avg=54
        --> "cpu1":{"subcore":{"usage":{"avg":54,"max":100,"min":21}}}
        x_foo_bar=21
        y_foo_bar=37
        --> "bar":{"foo":{"x"=21,"y"=21}}

- To keep "_" in a field name add /_
        x/_foo/_bar=21
        y/_foo/_bar=33
        --> "x_foo_bar"=21
        --> "y_foo_bar"=33

- To specify an array, precede the fieldKey with @uniquevalue_.  The uniquevalue string can be anything, as long as it makes the metric unique.
        @1_sysCap_lldp=11,@xx_sysCap_lldp=43,@abc_sysCap=87
        --> {"lldp":{"sysCap":[11,43,87]}}

*extr serializer batches metrics by default.

## Configuration

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

## Examples

The following Telegraf batched metrics

```text
CpuStats,serialnumber=XYZ-1234 core_key=0i,usage_min=35.1,usage_max=99.1,usage_avg=35.1
CpuStats,serialnumber=XYZ-1234 core_key=1i,usage_min=50.1,usage_max=88.1,usage_avg=51.1
FanStateChanged,serialnumber=XYZ-1234 tray_key=4i,fan_key=1i,partNumber_tag="1647G-00129 800751-00-01",revision_tag="01",airFlowDirection_tag="FrontToBack",state_old="Ok",state_new="Failed"
FanStateChanged,serialnumber=XYZ-1234 tray_key=4i,fan_key=2i,partNumber_tag="1647G-00129 800751-00-01",revision_tag="01",airFlowDirection_tag="FrontToBack",state_old="Failed",state_new="Ok"
FanStats,serialnumber=XYZ-1234 tray_key=2i,fan_key=10i,rpm_min=4101,rpm_max=5001,rpm_avg=4201,pwm_min=31,pwm_max=41,pwm_avg=31
FanStats,serialnumber=XYZ-1234 tray_key=2i,fan_key=11i,rpm_min=4001,rpm_max=4991,rpm_avg=4001,pwm_min=41,pwm_max=51,pwm_avg=41
FanStats,serialnumber=XYZ-1234 tray_key=3i,fan_key=9i,rpm_min=2101,rpm_max=3211,rpm_avg=2201,pwm_min=11,pwm_max=41,pwm_avg=11
InterfaceStateChanged,serialnumber=XYZ-1234 ifIndex_key=1001,name_key="1:1",adminStatus_old="Down",adminStatus_new="Up",operStatus_old="Down",operStatus_new="Up"
InterfaceStateChanged,serialnumber=XYZ-1234 ifIndex_key=1002,name_key="1:2",adminStatus_old="Down",adminStatus_new="Up",operStatus_old="Down",operStatus_new="Up"
CpuStats,serialnumber=XYZ-1234 core_key=0i,usage_min=10.2,usage_max=91.2,usage_avg=44.2
CpuStats,serialnumber=XYZ-1234 core_key=1i,usage_min=22.2,usage_max=89.2,usage_avg=41.2
CpuStats,serialnumber=XYZ-1234 core_key=2i,usage_min=33.2,usage_max=79.2,usage_avg=47.2
FanStats,serialnumber=XYZ-1234 tray_key=2i,fan_key=10i,rpm_min=4112,rpm_max=5012,rpm_avg=4212,pwm_min=32,pwm_max=52,pwm_avg=32
FanStats,serialnumber=XYZ-1234 tray_key=2i,fan_key=11i,rpm_min=5002,rpm_max=5092,rpm_avg=4102,pwm_min=52,pwm_max=62,pwm_avg=52
OspfNeighborStateChange,serialnum="ABCD-1234",reporterSerialnum="XYZ-5678" routerId_key=10,neighborAddress_key="10.20.4.1",neighborAdressLessInterface_key=0,neighborRouterId=10,name_vrf_key="vrf-1",id_vrf_key=99,state_old="Init",state_new="2Way"
OspfNeighborStateChange,serialnum="ABCD-1234",reporterSerialnum="XYZ-5678" routerId_key=10,neighborAddress_key="10.20.66.1",neighborAdressLessInterface_key=0,neighborRouterId=20,name_vrf_key="vrf-4",id_vrf_key=33,state_old="Exchange",state_new="Full"
LldpStateChange,serialnum="ABCD-1234" name_key="1:10",ifIndex_key=233,timeMark_key=0,index_key=2,chassisId_tag="5420",chassisIdSubtype_tag="MAC_ADDRESS",macSrcAddress_tag="01:03:22:33:00:66",@1_medCapSupported_tag="EXTENDED_PD",@1_medCapCurrent_tag="EXTENDED_PD",portDesc_tag="Extreme Networks Virtual Services Platform 7432CQ - 100GbCR4 Port 1/10",portId_tag="port 1:10",portIdSubtype_tag="INTERFACE_NAME",sysDescription_tag="VSP-7432CQ (9.0.0.0_B024) (PRIVATE)",@1_sysCapSupported_tag="ROUTER",@2_sysCapSupported_tag="BRIDGE",@3_sysCapSupported_tag="REPEATER",@1_sysCapEnabled_tag="ROUTER",sysName_tag="VSP7432-1234",reason="LLDP neighbor removed",state_old="UP",state_new="DOWN",usage_subcore_cpu1_min=21,usage_subcore_cpu1_max=99,usage_subcore_cpu1_avg=56
```

will serialize into the following extr JSON ouput

```json
{
  "cpuStats": [
    {
      "device": {
        "serialnumber": "XYZ-1234"
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
      "ts": 1654791970
    },
    {
      "device": {
        "serialnumber": "XYZ-1234"
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
      "ts": 1654791970
    }
  ],
  "fanStateChanged": [
    {
      "device": {
        "serialnumber": "XYZ-1234"
      },
      "items": [
        {
          "keys": {
            "fan": 1,
            "tray": 4
          },
          "state": {
            "new": "Failed",
            "old": "Ok"
          },
          "tags": {
            "airFlowDirection": "FrontToBack",
            "partNumber": "1647G-00129 800751-00-01",
            "revision": "01"
          }
        },
        {
          "keys": {
            "fan": 2,
            "tray": 4
          },
          "state": {
            "new": "Ok",
            "old": "Failed"
          },
          "tags": {
            "airFlowDirection": "FrontToBack",
            "partNumber": "1647G-00129 800751-00-01",
            "revision": "01"
          }
        }
      ],
      "name": "FanStateChanged",
      "ts": 1654791970
    }
  ],
  "fanStats": [
    {
      "device": {
        "serialnumber": "XYZ-1234"
      },
      "items": [
        {
          "keys": {
            "fan": 10,
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
      "ts": 1654791970
    },
    {
      "device": {
        "serialnumber": "XYZ-1234"
      },
      "items": [
        {
          "keys": {
            "fan": 10,
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
      "ts": 1654791970
    }
  ],
  "interfaceStateChanged": [
    {
      "device": {
        "serialnumber": "XYZ-1234"
      },
      "items": [
        {
          "adminStatus": {
            "new": "Up",
            "old": "Down"
          },
          "keys": {
            "ifIndex": 1001,
            "name": "1:1"
          },
          "operStatus": {
            "new": "Up",
            "old": "Down"
          }
        },
        {
          "adminStatus": {
            "new": "Up",
            "old": "Down"
          },
          "keys": {
            "ifIndex": 1002,
            "name": "1:2"
          },
          "operStatus": {
            "new": "Up",
            "old": "Down"
          }
        }
      ],
      "name": "InterfaceStateChanged",
      "ts": 1654791970
    }
  ],
  "lldpStateChange": [
    {
      "device": {
        "serialnum": "\"ABCD-1234\""
      },
      "items": [
        {
          "cpu1": {
            "subcore": {
              "usage": {
                "avg": 56,
                "max": 99,
                "min": 21
              }
            }
          },
          "keys": {
            "ifIndex": 233,
            "index": 2,
            "name": "1:10",
            "timeMark": 0
          },
          "reason": "LLDP neighbor removed",
          "state": {
            "new": "DOWN",
            "old": "UP"
          },
          "tags": {
            "chassisId": "5420",
            "chassisIdSubtype": "MAC_ADDRESS",
            "macSrcAddress": "01:03:22:33:00:66",
            "medCapCurrent": [
              "EXTENDED_PD"
            ],
            "medCapSupported": [
              "EXTENDED_PD"
            ],
            "portDesc": "Extreme Networks Virtual Services Platform 7432CQ - 100GbCR4 Port 1/10",
            "portId": "port 1:10",
            "portIdSubtype": "INTERFACE_NAME",
            "sysCapEnabled": [
              "ROUTER"
            ],
            "sysCapSupported": [
              "ROUTER",
              "BRIDGE",
              "REPEATER"
            ],
            "sysDescription": "VSP-7432CQ (9.0.0.0_B024) (PRIVATE)",
            "sysName": "VSP7432-1234"
          }
        }
      ],
      "name": "LldpStateChange",
      "ts": 1684541930
    }
  ],
  "ospfNeighborStateChange": [
    {
      "device": {
        "reporterSerialnum": "\"XYZ-5678\"",
        "serialnum": "\"ABCD-1234\""
      },
      "items": [
        {
          "keys": {
            "neighborAddress": "10.20.4.1",
            "neighborAdressLessInterface": 0,
            "routerId": 10,
            "vrf": {
              "id": 99,
              "name": "vrf-1"
            }
          },
          "neighborRouterId": 10,
          "state": {
            "new": "2Way",
            "old": "Init"
          }
        },
        {
          "keys": {
            "neighborAddress": "10.20.66.1",
            "neighborAdressLessInterface": 0,
            "routerId": 10,
            "vrf": {
              "id": 33,
              "name": "vrf-4"
            }
          },
          "neighborRouterId": 20,
          "state": {
            "new": "Full",
            "old": "Exchange"
          }
        }
      ],
      "name": "OspfNeighborStateChange",
      "ts": 1671059160
    }
  ]
}
```
