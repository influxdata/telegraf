# CDP (Common Data Platform)

The `cdp` output data format converts metrics into JSON payloads to be consumed by CDP (Common Data Platform) for
metered billing.

This serializer expects the following tags to be defined on the metric.

* `service`: Fills the CDP `serviceId` property.
* `project_id`: Fills the CDP `projectGenesisId` property.
* `environment_id`: Optional, fills the CDP `environmentId` property.
* `billing_region_id`: Optional, fills the CDP `region` property.
* `fleet_id`: Optional, fills the CDP `tags.fleetId` property.
* `metering_event_machine`: Optional, fills the CDP `tags.machineId` property.
* `virtual_type`: Optional, fills the CDP `tags.infraType` property.

It also expects the following fields to be defined on the metric.

* `start_time`: Fills the CDP `startTime` property. Must be in the RFC3339 format.
* `quantity_total`: Fills the CDP `quantity` property. Must be a floating point value.

### Configuration

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "cdp"
```

### Examples:

```json
{
  "type": "unity.services.systemUsage.v1",
  "msg": {
    "ts": 1636983542529,
    "eventId": "2669820a-5d85-4eb7-9a79-f6fd339b85c9",
    "serviceId": "MP",
    "projectId": "",
    "projectGenesisId": "some-project-id",
    "environmentId": "dev",
    "region": "some-billing-region-id",
    "startTime": 1636983540000,
    "endTime": 1636983541000,
    "type": "egress",
    "amount": 4822.75390625,
    "tags":{
      "fleetId": "some-fleet-id",
      "machineId": "some-machine-id",
      "infraType": "some-infra-type"
    }
  }
}
```
