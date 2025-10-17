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
    - offset_seconds (int64) - The offset from local and reference clock.
    - frequency (int64) - Local clock frequency offset.
    - maxerror_ns (int64) - The maximum error in nanoseconds.
    - estimated_error_ns (int64) - The estimated error in nanoseconds.
    - loop_time_constant (int64) - Phase-locked loop time constant.
    - tick_ns (int64) - Nanoseconds between clock ticks.
    - pps_frequency_hertz (float) - Pulse-per-second frequency in hertz.
    - pps_jitter_ns (int64) - Pulse-per-second jitter in nanoseconds.
    - pps_shift_seconds (int64) - Pulse-per-second interval duration in
    seconds.
    - pps_stability_hertz (float) - Pulse-per-second stability, average of
    relative.
    - pps_jitter_total (int64) - Pulse-per-second per second count of jitter
    limit.
    - pps_calibration_total (int64) - Pulse-per-second count of calibration
    intervals.
    - pps_error_total (int64) - Pulse-per-second count of calibration errors.
    - pps_stability_exceeded_total (int64) - Pulse-per-second total stability.
    - tai_offset_seconds (int64) - TAI offset in seconds.
    - sync_status (boolean) - Is clock synchronized with a server.
    - status (int) - Clock command/status.

## Example Output

```text
timex,host=testvm offset_ns=0i,loop_time_constant=2i,pps_frequency_hertz=0,pps_jitter_ns=0i,pps_jitter_total=0i,pps_calibration_total=0i,pps_stability_exceeded_total=0i,tai_offset_seconds=37i,pps_shift_seconds=0i,maxerror_ns=522000i,estimated_error_ns=9000i,status=0i,pps_stability_hertz=0,pps_error_total=0i,frequency=890113i,tick_ns=10000000i,sync_status=true 1760629900000000000
```
