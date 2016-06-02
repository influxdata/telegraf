## Telegraf Plugin: MEM

#### Description

The mem plugin collects memory metrics, defined as follows. For a more complete
explanation of the difference between `used` and `actual_used` RAM, see
[Linux ate my ram](http://www.linuxatemyram.com/).

- **total**: total physical memory available
- **available**: the actual amount of available memory that can be given instantly
to processes that request more memory in bytes; In linux kernel 3.14+, this
is available natively in /proc/meminfo. In other platforms, this is calculated by
summing different memory values depending on the platform
(e.g. free + buffers + cached on Linux).
It is supposed to be used to monitor actual memory usage in a cross platform fashion.
- **available_percent**: Percent of memory available, `available / total * 100`
- **used**: memory used, calculated differently depending on the platform and
designed for informational purposes only.
- **free**: memory not being used at all (zeroed) that is readily available; note
that this doesn't reflect the actual memory available (use 'available' instead).
- **used_percent**: the percentage usage calculated as `used / total * 100`

## Measurements:
#### Raw Memory measurements:

Meta:
- units: bytes
- tags: `nil`

Measurement names:
- mem_total
- mem_available
- mem_used
- mem_free

#### Derived usage percentages:

Meta:
- units: percent (out of 100)
- tags: `nil`

Measurement names:
- mem_used_percent
- mem_available_percent
