# ah_airrm Input Plugin

The `ah_airrm` plugin gather metrics on the rrm neighbor list from HiveOS.

## Configuration

```toml
# Read metrics about wireless stats
[[inputs.ah_airrm]]
  # Sample interval
  interval = "5s"
  # Interface names from where stats to be collected
  ifname = ["wifi0","wifi1"]
```

## Metrics

type: object
description: The RF neighbor state information.
properties:
  keys:
    $ref: "../../../interfaces/components/schemas/InterfaceKeysElement.yaml"
  rrmId:
    description: RRM ID of the neighbor AP.
    type: integer
    format: int32
  channel:
    description: channel.
    type: integer
    format: int32
  channelWidth:
    description: channel.
    type: integer
    format: int32
  channelUtilization:
    description: total channel utilization in percentage
    type: integer
    format: int8
  interferenceUtilization:
    description: Non-WIFI interference in percentage
    type: integer
    format: int8
  rxObssUtilization:
    description: Receive overlapping BSS channel utilization in percent
    type: integer
    format: int8
  wifinterferenceUtilization:
    description: Total WIFI interference in percentage.
    type: integer
    format: int8
  packetErrorRate:
    description: packet error rate in percentage
    type: integer
    format: int8
  aggregationSize:
    description: number of frames in ampdu.
    type: integer
    format: int32
  clientCount:
    description: Number of connected clients.
    type: integer
    format: int32

  required:
    keys

example:
  keys:
    name: "wifi0"
    ifIndex: 1001
  rrmId: 12345
  channel: 36
  channelWidth: 20
  channelUtilization: 75
  interferenceUtilization: 10
  rxObssUtilization: 5
  wifinterferenceUtilization: 15
  packetErrorRate: 2
  aggregationSize: 10
  clientCount: 50
