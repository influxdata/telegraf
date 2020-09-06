# GTFS

This input plugin can be used to gather vehicle positions, trip updates and
service alerts from a GTFS-realtime feed. [GTFS-realtime](https://developers.google.com/transit/gtfs-realtime) 
is an extension of the [General Transit Feed Specification](https://developers.google.com/transit/gtfs).

### Configuration

This plugin can be configured to collect any combination of vehicle positions, trip updates
and service alerts. Enable any one of these three input types by populating its URL in the
plugin configuration. At least one of the three must be configured.

```toml
[[inputs.gtfs]]
  ## API Key
  # key = "${GTFS_API_KEY}"
  ## URL for fetching vehicle positions
  # vehicle_positions_url = "https://host.test/VehiclePositions.pb"
  ## URL for fetching vehicle positions
  # trip_updates_url = "https://host.test/TripUpdates.pb"
  ## URL for fetching vehicle positions
  # service_alerts_url = "https://host.test/ServiceAlerts.pb"

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"
```

### Metrics

Metrics can be collected for each of three GTFS-realtime feed types: vehicle positions, trip updates and service alerts.

#### Vehicle Positions

The vehicle `position` measurement captures information about the vehicles including location and congestion level.

##### Fields 

##### Tags

#### Trip Updates

The trip `update` measurement captures information about delays, cancellations and changed routes.

##### Fields 

##### Tags

#### Service Alerts

The service `alert` measurement captures information about stop moves and unforeseen events 
affecting a station, route or the entire network.

##### Fields 

##### Tags

### Examples

An example of vehicle positions from the Boston's MBTA GTFS-realtime feed:

```csv
position,host=redacted,route_id=47 longitude=-71.09119415283203,bearing=297,trip_id="43759936",route_id="47",latitude=42.33601379394531,vehicle_id="y1853",vehicle_label="1853" 1590961364000000000
position,host=redacted,route_id=86 longitude=-71.07364654541016,bearing=0,vehicle_id="y1407",latitude=42.38750076293945,vehicle_label="1407",trip_id="44563840",route_id="86" 1590961165000000000
position,host=redacted,route_id=117 trip_id="43589269",route_id="117",latitude=42.40916442871094,longitude=-70.99650573730469,bearing=289,vehicle_id="y0776",vehicle_label="0776" 1590961365000000000
position,host=redacted,route_id=Green-B bearing=135,speed=2.700000047683716,vehicle_label="3680-3889",route_id="Green-B",latitude=42.340179443359375,longitude=-71.1670913696289,vehicle_id="G-10177",trip_id="43832291-LechmereNorthStation" 1590961361000000000
position,host=redacted,route_id=95 vehicle_id="y2050",vehicle_label="2050",longitude=-71.13471221923828,trip_id="44557959",route_id="95",latitude=42.424129486083984,bearing=180 1590961244000000000
position,host=redacted,route_id=101 vehicle_label="1452",trip_id="44557600",route_id="101",latitude=42.40914535522461,longitude=-71.10965728759766,vehicle_id="y1452",bearing=166 1590961362000000000
position,host=redacted,route_id=94 longitude=-71.1059341430664,vehicle_id="y1400",vehicle_label="1400",trip_id="44557405",latitude=42.417030334472656,bearing=101,route_id="94" 1590961363000000000
position,host=redacted,route_id=32 latitude=42.24892044067383,vehicle_label="1617",trip_id="43725081",longitude=-71.12676239013672,bearing=0,vehicle_id="y1617",route_id="32" 1590961361000000000
```
