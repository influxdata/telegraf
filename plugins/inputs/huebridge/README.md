# HueBridge Input Plugin

This input plugin gathers status from [Hue Bridge][hue] devices
using the [CLIP API][hue_api] interface of the devices.

⭐ Telegraf v1.34.0
🏷️ iot
💻 all

[hue]: https://www.philips-hue.com/
[hue_api]: https://developers.meethue.com/develop/hue-api-v2/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather Hue smart home status
[[inputs.huebridge]]
  ## The Hue bridges to query.
  ## See README file for all addressing options.
  
  ## Manual device to room assignments to apply during status evaluation.
  ## E.g. for motion sensors which are reported without a room assignment.
  # room_assignments = { "Motion sensor 1" = "Living room", "Motion sensor 2" = "Corridor" }
  
  ## Timeout for gathering information
  # timeout = "10s"
  
  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  # tls_key_pwd = "secret"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Extended bridge access options

The Hue bridges to query can be defined URLs of the following form:

```toml
  <locator scheme>://<bridge id>:<user name>@<locator dependent address>/
```

where:

&lt;locator scheme&gt; is one of

- **address**: To identify the bridge via the DNS name or ip address within
the URLs address part (see example below).
- **cloud**: To identify the bridge via its cloud registration. The address
part defines the discovery endpoint. If empty the standard endpoint
[https://discovery.meethue.com/](https://discovery.meethue.com/) is used.
- **mdns**: To identify the bridge via mDNS. The URL's address part is always
empty in this case.
- **remote**: To identify the bridge via the Cloud Remote API. The address
part defines the cloud API endpoint. If empty the standard endpoint
[https://api.meethue.com/](https://api.meethue.com/) is used.

&lt;bridge id&gt; is the unique bridge id as returned in

```bash
curl -k https://<bridge address>/api/config/0
```

&lt;user name&gt; is the secret user name returned during application
authentication. To create a new user name issue the following command
after pressing the bridge's link button:

```bash
  curl -k -X POST http://<bridge address>/api \
    -H 'Content-Type: application/json' \
    -d '{"devicetype":"huebridge-telegraf-plugin"}'
```

Examples URLs:

- "address://0123456789ABCDEF:sFlEGnMAFXO6RtZV17aViNUB95G2uXWw64texDzD@mybridge/"
- "cloud://0123456789ABCDEF:sFlEGnMAFXO6RtZV17aViNUB95G2uXWw64texDzD@/"
- "mdns://0123456789ABCDEF:sFlEGnMAFXO6RtZV17aViNUB95G2uXWw64texDzD@/"
- "remote://0123456789ABCDEF:sFlEGnMAFXO6RtZV17aViNUB95G2uXWw64texDzD@/"

### Remote bridge access

A bridge registered to the cloud can be accessed remotely.
In order to access a bridge remotely a Hue Developer Account is required,
a Remote App must be registered and the corresponding Authorization flow must
be completed. See [Cloud2Cloud Getting Started][] in the developer
documentation for full details (Developer registration required).

Beside using a remote access URL as described above the following parameters
must be set in the configuration file exactly as used during App registration:

```toml
remote_client_id = ""
remote_client_secret = ""
remote_callback_url = ""
remote_token_dir = ""
```

The remote_token_dir points to the directory used to persist the token data.

[Cloud2Cloud Getting Started]: https://developers.meethue.com/develop/hue-api-v2/cloud2cloud-getting-started/

## Metrics

- `huebridge_light`
  - tags
    - `bridge_id` - The bridge id (this metrics has been queried from)
    - `room` - The name of the room
    - `device` - The name of the device
  - fields
    - `on` (int) - 0: light is off 1: light is on
- `huebridge_temperature`
  - tags
    - `bridge_id` - The bridge id (this metrics has been queried from)
    - `room` - The name of the room
    - `device` - The name of the device
    - `enabled` - The current status of sensor (active: true|false)
  - fields
    - `temperature` (float) - The current temperatue (in °Celsius)
- `huebridge_light_level`
  - tags
    - `bridge_id` - The bridge id (this metrics has been queried from)
    - `room` - The name of the room
    - `device` - The name of the device
    - `enabled` - The current status of sensor (active: true|false)
  - fields
    - `light_level` (int) - The current light level (in human friendly scale 10.000*log10(lux)+1)
    - `light_level_lux` (float) - The current light level (in lux)
- `huebridge_motion_sensor`
  - tags
    - `bridge_id` - The bridge id (this metrics has been queried from)
    - `room` - The name of the room
    - `device` - The name of the device
    - `enabled` - The current status of sensor (active: true|false)
  - fields
    - `motion` (int) - 0: no motion detected 1: motion detected
- `huebridge_device_power`
  - tags
    - `bridge_id` - The bridge id (this metrics has been queried from)
    - `room` - The name of the room
    - `device` - The name of the device
  - fields
    - `battery_level` (int) - Power source status (normal, low, critical)
    - `battery_state` (string) - Battery charge level (in %)

## Example Output

```text
huebridge_light,huebridge_bridge_id=0123456789ABCDEF,huebridge_room=Name#15,huebridge_device=Name#3 on=0 1734880329
huebridge_temperature,huebridge_room=Name#15,huebridge_device=Name#7,huebridge_device_enabled=true,huebridge_bridge_id=0123456789ABCDEF temperature=17.63 1734880329
huebridge_light_level,huebridge_bridge_id=0123456789ABCDEF,huebridge_room=Name#15,huebridge_device=Name#7,huebridge_device_enabled=true light_level=18948,light_level_lux=78.46934003526889 1734880329
huebridge_motion_sensor,huebridge_bridge_id=0123456789ABCDEF,huebridge_room=Name#15,huebridge_device=Name#7,huebridge_device_enabled=true motion=0 1734880329
huebridge_device_power,huebridge_bridge_id=0123456789ABCDEF,huebridge_room=Name#15,huebridge_device=Name#7 battery_level=100,battery_state=normal 1734880329
```
