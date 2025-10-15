# Timex Input Plugin

This plugin gathers metrics on system time.

⭐ Telegraf v1.37.0
🏷️ hardware, system
💻 linux

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about temperature
[[inputs.timex]]
  ## No input configuration
```

## Metrics

- timex
  - fields:
    - offset_seconds (float)
    - frequency_adjustment_ratio (float)
    - maxerror_seconds (float)
    - estimated_error_seconds (float)
    - loop_time_constant (float)
    - tick_seconds (float)
    - pps_frequency_hertz (float)
    - pps_jitter_seconds (float)
    - pps_shift_seconds (float)
    - pps_stability_hertz (float)
    - pps_jitter_total (float)
    - pps_calibration_total (float)
    - pps_error_total (float)
    - pps_stability_exceeded_total (float)
    - tai_offset_seconds (float)
    - sync_status (int)
    - status (float)

**Fields**
- offset_seconds - The offset from local and reference clock.
- frequency_adjustment_ratio - Local clock frequency adjustment ratio.
- maxerror_seconds - The maximum error in seconds.
- estimated_error_seconds - The estimated error in seconds.
- loop_time_constant - Phase-locked loop time constant.
- tick_seconds - Seconds between clock ticks.
- pps_frequency_hertz - Pulse-per-second frequency in hertz.
- pps_jitter_seconds - Pulse-per-second jitter in seconds.
- pps_shift_seconds - Pulse-per-second interval duration in seconds.
- pps_stability_hertz - Pulse-per-second stability, average of relative
frequency changes.
- pps_jitter_total - Pulse-per-second per second count of jitter limit
 exceeded events.
- pps_calibration_total - Pulse-per-second count of calibration intervals.
- pps_error_total - Pulse-per-second count of calibration errors.
- pps_stability_exceeded_total - Pulse-per-second total stability
exceeded in seconds.
- tai_offset_seconds - TAI offset in seconds.
- sync_status - Is clock synchronized with a server (1 = yes, 0 = no).
- status - Clock command/status

## Example Output

```text
timex,host=testvm pps_error_total=0,sync_status=1i,estimated_error_seconds=0.000006,loop_time_constant=2,pps_jitter_seconds=0,pps_shift_seconds=0,pps_jitter_total=0,pps_calibration_total=0,maxerror_seconds=0.004021,pps_frequency_hertz=0,pps_stability_exceeded_total=0,tai_offset_seconds=37,offset_seconds=0,frequency_adjustment_ratio=1.0000166235961914,status=0,tick_seconds=0.01,pps_stability_hertz=0 1760440880000000000
```
