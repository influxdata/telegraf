# CDP (Common Data Platform)

The `cdp` output data format converts metrics into JSON payloads to be consumed by CDP (Common Data Platform) for
metered billing.

This serializer expects the following tags to be defined on the metric.

* `service`: Fills the CDP `serviceId` property.
* `project_id`: Fills the CDP `projectId` property.
* `environment_id`: Optional, fills the CDP `environmentId` property.
* `billing_region_id`: Optional, fills the CDP `tags.multiplayRegion` property.
* `fleet_id`: Optional, fills the CDP `tags.multiplayFleetId` property.
* `metering_event_machine`: Optional, fills the CDP `tags.multiplayMachineId` property.
* `virtual_type`: Optional, fills the CDP `tags.multiplayInfraType` property.

It also expects the following fields to be defined on the metric.

* `start_time`: Fills the CDP `startTime` property. Must be in the RFC3339 format.
* `quantity`: Fills the CDP `quantity` property. Must be a floating point value.

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
  "ts": 1636983542529,
  "eventId": "2669820a-5d85-4eb7-9a79-f6fd339b85c9",
  "fingerprint": "6199419141951568871",
  "serviceId": "MP",
  "projectId": "some-project-id",
  "environmentId": "dev",
  "playerId": "",
  "startTime": 1636983540000,
  "endTime": 1636983541000,
  "type": "network_usage_event",
  "amount": 1151801145,
  "tags":{
    "multiplayFleetId": "some-fleet-id",
    "multiplayMachineId": "some-machine-id",
    "multiplayInfraType": "some-infra-type",
    "multiplayProjectId": "ASID:123",
    "multiplayRegion": "some-billing-region-id",
    "analyticsEventType": "",
    "analyticsEventName": "",
  }
}
```
