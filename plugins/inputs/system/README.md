# System Input Plugin

This plugin gathers general system statistics like system load, uptime or the
number of users logged in. It is similar to the unix `uptime` command.

⭐ Telegraf v0.1.6
🏷️ system
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read metrics about system load & uptime
[[inputs.system]]
  ## Information to collect; available options are:
  ##   legacy - legacy layout of system metrics; see README for details
  ##   cpus   - CPU counts of the system
  ##   dmi    - BIOS, baseboard, chassis and product information from DMI/SMBIOS
  ##   load   - 1, 5 and 15-minute load averages
  ##   os     - operating system release and uname information
  ##   uptime - system uptime
  ##   users  - logged-in user counts
  # include = ["legacy"]

  ## How long to cache the result of the "os" group between gathers.
  ## Set higher to reduce the number of os-release/uname reads, lower to
  ## surface distro upgrades and kexec'd kernels faster. Set to zero to
  ## re-read the data on every gather.
  # os_cache_ttl = "8h"

  ## How long to cache the result of the "dmi" group between gathers.
  ## DMI/SMBIOS data is effectively static for the life of the machine,
  ## so a long cache is typical. Set to zero to re-read on every gather.
  # dmi_cache_ttl = "8h"
```

### `legacy` include vs. fine-grained options

While `legacy` is the default `include` for compatibility reasons, it is _not_
recommended as it creates three, sparse metrics, one containing the CPU, load
and user information, one for the uptime and one for the formatted uptime.
However, if you are using Prometheus as an output, it might be your preferred
choice as the metrics are typed correctly.

If you are using the fine-grained options such as `cpus` you will get a _single_,
untyped metric containing all selected information. To reproduce the fields of
the `legacy` set, use `include = ["load", "users", "cpus", "uptime"]`.

Both `legacy` and the fine-grained options can be used at the same time
resulting in four metrics in total (one for the fine-grained set and three for
the `legacy` setting).

### Permissions

The `n_users` field requires read access to `/var/run/utmp`, and may require the
`telegraf` user to be added to the `utmp` group on some systems. If this file
does not exist `n_users` will be skipped.

The `n_unique_users` shows the count of unique usernames logged in. This way if
a user has multiple sessions open/started they would only get counted once. The
same requirements for `n_users` apply.

The `os` group reads `/etc/os-release` on Linux (typically world-readable) and
calls the `uname` syscall on POSIX systems. The `os` field is always populated
from Go's runtime, and `arch` falls back to the runtime architecture when the
kernel cannot be queried, so both are always present. On platforms where
gopsutil cannot provide platform release or kernel data (e.g. parts of
FreeBSD/OpenBSD/Solaris) the `platform`, `platform_family`, `platform_version`
and `kernel_version` fields may be empty. Results are cached between gathers,
see `os_cache_ttl` above.

The `dmi` group exposes BIOS, baseboard, chassis and product information from
DMI/SMBIOS. On Linux the data is read from `/sys/class/dmi/id/` and does not
require root access for most fields; serial numbers and asset tags are
generally restricted by the kernel. On Windows the data is read via WMI.
macOS, BSD and Solaris are not supported and the `dmi` value is ignored
there. Results are cached between gathers, see `dmi_cache_ttl` above.

## Metrics

The `include` option controls which measurements and fields are gathered.
The fine-grained options like `cpus`, `load` or `users` populate the `system`
metric, while the `legacy` setting emits three `system` metrics.

### fine-grained options

When `os` is included, the values reflect operating system release information
together with `uname`-style kernel data. Fields are reported as strings. The
`os` and `arch` fields are always populated; the `platform`, `platform_family`,
`platform_version` and `kernel_version` fields may be empty on unsupported
platforms.

When `dmi` is included fields may be empty if the field is not exposed by the
system, or `unknown` when it is restricted by the kernel (typical for serial
numbers, asset tags and the product UUID on Linux without root).

#### `system`

| Field               | Include option | Type    | Description                                    |
|---------------------|----------------|---------|------------------------------------------------|
| `n_cpus`            | `cpus`         | integer | Number of logical CPUs                         |
| `n_physical_cpus`   | `cpus`         | integer | Number of physical CPUs                        |
| `bios_vendor`       | `dmi`          | string  | BIOS vendor (e.g. `Dell Inc.`)                 |
| `bios_version`      | `dmi`          | string  | BIOS version (e.g. `2.18.0`)                   |
| `bios_date`         | `dmi`          | string  | BIOS release date (e.g. `04/12/2024`)          |
| `board_vendor`      | `dmi`          | string  | Baseboard / motherboard vendor                 |
| `board_product`     | `dmi`          | string  | Baseboard product name (e.g. `0X3D66`)         |
| `board_version`     | `dmi`          | string  | Baseboard version                              |
| `board_serial`      | `dmi`          | string  | Baseboard serial number (restricted)           |
| `board_asset_tag`   | `dmi`          | string  | Baseboard asset tag (restricted)               |
| `chassis_vendor`    | `dmi`          | string  | Chassis vendor                                 |
| `chassis_type_code` | `dmi`          | string  | Chassis type code as defined by SMBIOS DSP0134 |
| `chassis_type`      | `dmi`          | string  | Human-readable chassis type description        |
| `chassis_version`   | `dmi`          | string  | Chassis version                                |
| `chassis_serial`    | `dmi`          | string  | Chassis serial number (restricted)             |
| `chassis_asset_tag` | `dmi`          | string  | Chassis asset tag (restricted)                 |
| `product_vendor`    | `dmi`          | string  | System product vendor (e.g. `Dell Inc.`)       |
| `product_name`      | `dmi`          | string  | System product name (e.g. `PowerEdge R750`)    |
| `product_family`    | `dmi`          | string  | System product family                          |
| `product_version`   | `dmi`          | string  | System product version                         |
| `product_serial`    | `dmi`          | string  | System product serial number (restricted)      |
| `product_sku`       | `dmi`          | string  | System product SKU                             |
| `product_uuid`      | `dmi`          | string  | System product UUID (restricted)               |
| `load1`             | `load`         | float   | 1-minute load average                          |
| `load5`             | `load`         | float   | 5-minute load average                          |
| `load15`            | `load`         | float   | 15-minute load average                         |
| `os`                | `os`           | string  | OS family                                      |
| `arch`              | `os`           | string  | Architecture                                   |
| `platform`          | `os`           | string  | OS distribution / platform identifier          |
| `platform_family`   | `os`           | string  | Platform family (e.g. `debian`, `rhel`)        |
| `platform_version`  | `os`           | string  | Platform / distribution version                |
| `kernel_version`    | `os`           | string  | Kernel release as returned by `uname -r`       |
| `uptime`            | `uptime`       | integer | System uptime in seconds                       |
| `n_users`           | `users`        | integer | Number of logged-in user sessions              |
| `n_unique_users`    | `users`        | integer | Number of unique logged-in usernames           |

The resulting metric is untyped.

### legacy setting

The following three metrics are emitted if `include` contains the `legacy`
setting.

#### `system` (gauge)

| Field             | Type    | Description                                    |
|-------------------|---------|------------------------------------------------|
| `load1`           | float   | 1-minute load average                          |
| `load5`           | float   | 5-minute load average                          |
| `load15`          | float   | 15-minute load average                         |
| `n_users`         | integer | Number of logged-in user sessions              |
| `n_unique_users`  | integer | Number of unique logged-in usernames           |
| `n_cpus`          | integer | Number of logical CPUs                         |
| `n_physical_cpus` | integer | Number of physical CPUs                        |
| `uptime`          | integer | System uptime in seconds                       |

#### `system` (counter)

| Field             | Type    | Description                                    |
|-------------------|---------|------------------------------------------------|
| `uptime`          | integer | System uptime in seconds                       |

#### `system` (untyped)

| Field             | Type    | Description                                    |
|-------------------|---------|------------------------------------------------|
| `uptime_format`   | string  | Human-readable uptime                          |

## Example Output

### Default configuration

With the default `include = ["legacy"]` the output is backward-compatible with
previous versions:

```text
system,host=worker-01 load1=3.72,load5=2.4,load15=2.1,n_users=3i,n_unique_users=2i,n_cpus=4i,n_physical_cpus=2i 1748000000000000000
system,host=worker-01 uptime=1249632i 1748000000000000000
system,host=worker-01 uptime_format="14 days, 11:07" 1748000000000000000
```

### Recommended configuration

With `include = ["cpus", "load", "users", "uptime"]`, all fields are emitted
in a single metric with the new field names:

```text
system,host=worker-01 load1=3.72,load5=2.4,load15=2.1,n_users=3i,n_unique_users=2i,n_cpus=4i,n_physical_cpus=2i,uptime=1249632i 1748000000000000000
```

When including all options the emitted metric will be

```text
system,host=worker-01 load1=3.72,load5=2.4,load15=2.1,n_users=3i,n_unique_users=2i,n_cpus=4i,n_physical_cpus=2i,uptime=1249632i os="linux",arch="x86_64",platform="ubuntu",platform_family="debian",platform_version="26.04",kernel_version="7.0.0-7-generic" bios_vendor="Dell Inc.",bios_version="2.18.0",bios_date="04/12/2024",board_vendor="Dell Inc.",board_product="0X3D66",board_version="A00",board_serial="CN747503AB0123",board_asset_tag="",chassis_vendor="Dell Inc.",chassis_type="23",chassis_type_description="Rack mount chassis",chassis_version="",chassis_serial="7XK4P03",chassis_asset_tag="",product_vendor="Dell Inc.",product_name="PowerEdge R750",product_family="PowerEdge",product_version="",product_serial="7XK4P03",product_sku="SKU=NotProvided;ModelName=PowerEdge R750",product_uuid="4c4c4544-0058-4b10-8034-b3c04f503033" 1748000000000000000
```
