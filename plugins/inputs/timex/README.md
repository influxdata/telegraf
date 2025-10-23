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
  - tags:
    - status (string) - Clock command/status.
  - fields:
    - offset_ns (int64) - The offset from local and reference clock.
    - frequency_hz (float) - Local clock frequency offset.
    - maxerror_ns (int64) - The maximum error in nanoseconds.
    - estimated_error_ns (int64) - The estimated error in nanoseconds.
    - loop_time_constant (int64) - Phase-locked loop time constant.
    - tick_ns (int64) - Nanoseconds between clock ticks.
    - pps_frequency_hz (float) - Pulse-per-second frequency in hertz.
    - pps_jitter_ns (int64) - Pulse-per-second jitter in nanoseconds.
    - pps_shift_sec (int64) - Pulse-per-second interval duration in
    seconds.
    - pps_stability_hz (float) - Pulse-per-second stability, average of
    relative.
    - pps_jitter_total (int64) - Pulse-per-second per second count of jitter
    limit.
    - pps_calibration_total (int64) - Pulse-per-second count of calibration
    intervals.
    - pps_error_total (int64) - Pulse-per-second count of calibration errors.
    - pps_stability_exceeded_total (int64) - Pulse-per-second total stability.
    - tai_offset_sec (int64) - TAI offset in seconds.
    - synchronized (boolean) - Is clock synchronized with a server.
    - status (int) - Clock command/status.

## Example Output

```text
timex,host=testvm,status=ok maxerror_ns=1516000i,estimated_error_ns=4000i,tick_ns=10000000i,loop_time_constant=2i,pps_jitter_total=0i,synchronized=true,offset_ns=0i,frequency_hz=55.05543,pps_shift_sec=0i,pps_stability_hz=0,tai_offset_sec=37i,status=0i,pps_frequency_hz=0,pps_jitter_ns=0i,pps_calibration_total=0i,pps_error_total=0i,pps_stability_exceeded_total=0i 1761121800000000000
```
