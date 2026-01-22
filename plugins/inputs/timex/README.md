# Timex Input Plugin

This plugin gathers metrics on system time using the Linux Kernel
[adjtimex syscall][timex].

The call gets the information of the kernel time variables that are controlled
by the ntpd, systemd-timesyncd, chrony or other time synchronization services.

⭐ Telegraf v1.37.0
🏷️ hardware, system
💻 linux

[timex]: https://man7.org/linux/man-pages/man2/adjtimex.2.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read time metrics from linux timex interface.
[[inputs.timex]]
  ## No input configuration
```

## Metrics

Metric fields usually have a suffix denoting the unit of the field with
`ns` being nanoseconds, `sec` being seconds and `ppm` being parts-per-million.
The parts-per-million unit is defined as
`1 ppm` corresponds to `10^-6` or `0.0001 %`.

- timex
  - tags:
    - status (string) - Clock command/status.
  - fields:
    - offset_ns (int64) - The offset from local and reference clock.
    - frequency_offset_ppm (float) - Local clock frequency offset in parts per
    million.
    - maxerror_ns (int64) - The maximum error in nanoseconds.
    - estimated_error_ns (int64) - The estimated error in nanoseconds.
    - loop_time_constant (int64) - Phase-locked loop time constant.
    - tick_ns (int64) - Nanoseconds between clock ticks.
    - pps_frequency_ppm (float) - Pulse-per-second frequency in parts per
    million.
    - pps_jitter_ns (int64) - Pulse-per-second jitter in nanoseconds.
    - pps_shift_sec (int64) - Pulse-per-second interval duration in
    seconds.
    - pps_stability_ppm (float) - Pulse-per-second stability, average of
    relative in parts per million.
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
timex,host=testvm,status=ok maxerror_ns=1516000i,estimated_error_ns=4000i,tick_ns=10000000i,loop_time_constant=2i,pps_jitter_total=0i,synchronized=true,offset_ns=0i,frequency_offset_ppm=55.05543,pps_shift_sec=0i,pps_stability_ppm=0,tai_offset_sec=37i,status=0i,pps_frequency_ppm=0,pps_jitter_ns=0i,pps_calibration_total=0i,pps_error_total=0i,pps_stability_exceeded_total=0i 1761121800000000000
```
