# Timex Input Plugin

This plugin gathers metrics on system time using the Linux Kernel [adjtimex syscall][timex].

The call gets the information of the kernel time variables that are controlled
 by the ntpd, systemd-timesyncd, chrony or other time synchronization services.

⭐ Telegraf v1.37.0
🏷️ hardware, system
💻 linux

[timex]: https://man7.org/linux/man-pages/man2/adjtimex.2.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read time metrics from linux timex interface.
[[inputs.timex]]
  ## No input configuration
```

## Metrics

- timex
  - fields:
    - offset_seconds (float) - The offset from local and reference clock.
    - frequency_adjustment_ratio (float) - Local clock frequency adjustment
     ratio.
    - maxerror_seconds (float) - The maximum error in seconds.
    - estimated_error_seconds (float) - The estimated error in seconds.
    - loop_time_constant (float) - Phase-locked loop time constant.
    - tick_seconds (float) - Seconds between clock ticks.
    - pps_frequency_hertz (float) - Pulse-per-second frequency in hertz.
    - pps_jitter_seconds (float) - Pulse-per-second jitter in seconds.
    - pps_shift_seconds (float) - Pulse-per-second interval duration in
    seconds.
    - pps_stability_hertz (float) - Pulse-per-second stability, average of
    relative.
    - pps_jitter_total (float) - Pulse-per-second per second count of jitter
    limit.
    - pps_calibration_total (float) - Pulse-per-second count of calibration
    intervals.
    - pps_error_total (float) - Pulse-per-second count of calibration errors.
    - pps_stability_exceeded_total (float) - Pulse-per-second total stability.
    - tai_offset_seconds (float) - TAI offset in seconds.
    - sync_status (boolean) - Is clock synchronized with a server.
    - status (float) - Clock command/status.

## Example Output

```text
timex,host=testvm tick_ns=10000000i,loop_time_constant=2i,pps_stability_hertz=0,pps_stability_exceeded_total=0i,maxerror_ns=2014000i,pps_jitter_ns=0i,pps_shift_seconds=0i,pps_jitter_total=0i,estimated_error_ns=5000i,pps_frequency_hertz=0,pps_calibration_total=0i,pps_error_total=0i,tai_offset_seconds=37i,sync_status=true,offset_ns=0i,frequency_adjustment_ratio=1.0000130641326905,status=0i 1760609100000000000
```
