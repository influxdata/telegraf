# Fireboard Input Plugin

This plugin gathers real-time temperature data from [fireboard][fireboard]
thermometers.

> [!NOTE]
> You will need to sign up to for the [Fireboard REST API][api] in order to use
> this plugin.

⭐ Telegraf v1.12.0
🏷️ iot
💻 all

[fireboard]: https://www.fireboard.com
[api]: https://docs.fireboard.io/reference/restapi.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read real time temps from fireboard.io servers
[[inputs.fireboard]]
  ## Specify auth token for your account
  auth_token = "invalidAuthToken"
  ## You can override the fireboard server URL if necessary
  # url = https://fireboard.io/api/v1/devices.json
  ## You can set a different http_timeout if you need to
  ## You should set a string using an number and time indicator
  ## for example "12s" for 12 seconds.
  # http_timeout = "4s"
```

### auth_token

In lieu of requiring a username and password, this plugin requires an
authentication token that you can generate using the [Fireboard REST
API](https://docs.fireboard.io/reference/restapi.html#Authentication).

### url

While there should be no reason to override the URL, the option is available
in case Fireboard changes their site, etc.

### http_timeout

If you need to increase the HTTP timeout, you can do so here. You can set this
value in seconds. The default value is four (4) seconds.

## Rate Limiting

The Fireboard API enforces a **17-call / 5-minute** rate limit. Set your
collection interval accordingly:

| Interval | Calls per 5 min | Status |
|----------|-----------------|--------|
| `30s`    | 10              | ✅ Safe |
| `15s`    | 20              | ❌ Over limit |

## Metrics

All four measurements are populated from a single API call to
`/api/v1/devices.json` — no additional rate-limit cost.

- fireboard
  - tags:
    - channel (channel number)
    - channel_name (human-readable probe label; empty string if not found)
    - scale (`Celcius` or `Fahrenheit`)
    - title (name of the Fireboard device)
    - uuid (UUID of the Fireboard device)
  - fields:
    - temperature (float)

- fireboard_drive (emitted per device when `last_drivelog` is non-null)
  - tags:
    - model (device model, e.g. `FBXPD`)
    - pel_profile (active pellet profile name, e.g. `Hickory`)
    - pel_state (pellet grill state machine, e.g. `RUNNING`, `COOLDOWN`)
    - title
    - uuid
  - fields:
    - current_temp (float)
    - drive_percent (float) — fan/auger output %
    - drive_type (int)
    - fan_speed_target (float)
    - lid_paused (bool)
    - mode_type (int)
    - pid_error_sum (float)
    - pid_kd (float)
    - pid_kd_output (float)
    - pid_ki (float)
    - pid_ki_output (float)
    - pid_kp (float)
    - pid_kp_output (float)
    - pid_last_error (float)
    - pid_output (float)
    - power_mode (int)
    - setpoint (float) — target temperature
    - tied_channel (int) — control channel number
    - vbatt (float) — battery voltage

- fireboard_alert (emitted per alert entry per channel)
  - tags:
    - channel
    - channel_name
    - scale
    - title
    - uuid
  - fields:
    - enabled (bool)
    - temp_max (float, omitted when null)
    - temp_min (float, omitted when null)

- fireboard_device (emitted per device when `device_log` is non-null)
  - tags:
    - model
    - title
    - uuid
  - fields:
    - onboard_temp (float) — device internal temperature
    - signal_level (int) — WiFi RSSI in dBm
    - vbatt (float) — battery voltage
    - vbatt_percent (float) — battery fraction (0.0–1.0)

## Example Output

```text
fireboard,channel=1,channel_name=Instant\ Probe,host=myhost,scale=Fahrenheit,title=telegraf-FireBoard,uuid=b55e766c-b308-49b5-93a4-df89fe31efd0 temperature=79.9 1561690040000000000
fireboard_drive,model=FBXPD,pel_state=RUNNING,title=pellet-grill,uuid=aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa drive_percent=42.5,setpoint=350,current_temp=348.2,fan_speed_target=60 1561690040000000000
fireboard_alert,channel=1,channel_name=Instant\ Probe,scale=Fahrenheit,title=telegraf-FireBoard,uuid=b55e766c-b308-49b5-93a4-df89fe31efd0 enabled=true,temp_max=205 1561690040000000000
fireboard_device,model=FBXPD,title=pellet-grill,uuid=aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa onboard_temp=95,signal_level=-65,vbatt=12.3,vbatt_percent=0.85 1561690040000000000
```
