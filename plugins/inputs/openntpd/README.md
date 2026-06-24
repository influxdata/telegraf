# OpenNTPD Input Plugin

This plugin gathers metrics from [OpenNTPD][openntpd] using the `ntpctl`
command.

> [!NOTE]
> The `ntpctl` binary must be present on the system and executable by Telegraf.
> The plugin supports using `sudo` for execution.

⭐ Telegraf v1.12.0
🏷️ server, network
💻 all

[openntpd]: http://www.openntpd.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Get standard NTP query metrics from OpenNTPD.
[[inputs.openntpd]]
  ## Run ntpctl binary with sudo.
  # use_sudo = false

  ## Location of the ntpctl binary.
  # binary = "/usr/sbin/ntpctl"

  ## Maximum time the ntpctl binary is allowed to run.
  # timeout = "5s"
```

### Permissions

It's important to note that this plugin references `ntpctl`, which may require
additional permissions to execute successfully. Depending on the user/group
permissions of the telegraf user executing this plugin, you may need to alter
the group membership, set facls, or use sudo.

#### Group membership (recommended)

```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G ntpd telegraf

$ groups telegraf
telegraf : telegraf ntpd
```

#### Sudo privileges

If you use this method, you will need the following in your telegraf config:

```toml
[[inputs.openntpd]]
  use_sudo = true
```

You will also need to update your sudoers file:

```bash
$ visudo
# Add the following lines:
Cmnd_Alias NTPCTL = /usr/sbin/ntpctl
telegraf ALL=(ALL) NOPASSWD: NTPCTL
Defaults!NTPCTL !logfile, !syslog, !pam_session
```

Please use the solution you see as most appropriate.

## Metrics

- openntpd
  - tags:
    - remote (remote peer address or hostname)
    - stratum (remote peer stratum)
    - state_prefix (ntpctl state indicator, e.g. `*` for active peer;
      omitted when absent)
  - fields:
    - delay (round trip delay to the remote peer in milliseconds; `float`)
    - jitter (mean deviation (jitter) for remote peer in milliseconds; `float`)
    - offset (mean offset (phase) to remote peer in milliseconds; `float`)
    - poll (polling interval in seconds; `int`)
    - next (number of seconds until the next poll; `int`)
    - wt (peer weight; `int`)
    - tl (peer trust level; `int`)

- openntpd_sensors (one metric per hardware/GPS sensor)
  - tags:
    - sensor (sensor device name, e.g. `nmea0`)
    - refid (sensor reference ID, e.g. `GPS`)
    - state_prefix (ntpctl state indicator, e.g. `*` for active sensor;
      omitted when absent)
  - fields:
    - wt (sensor weight; `int`)
    - gd (sensor good count; `int`)
    - st (sensor stratum; `int`)
    - next (seconds until next poll; `int`)
    - poll (polling interval in seconds; `int`)
    - offset (sensor offset in milliseconds; `float`)
    - correction (sensor correction in milliseconds; `float`)

- openntpd_status (one metric per gather, system-level summary)
  - fields:
    - peers_valid (number of currently valid peers; `int`)
    - peers_total (total number of configured peers; `int`)
    - sensors_valid (number of currently valid sensors; `int`)
    - sensors_total (total number of configured sensors; `int`)
    - constraint_offset_s (constraint offset in seconds; `int`)
    - clock_synced (1 if the local clock is synced, 0 otherwise; `int`)
    - stratum (local clock stratum; `int`)

## Example Output

```text
openntpd,remote=194.57.169.1,stratum=2 tl=10i,poll=1007i,offset=2.295,jitter=3.896,delay=53.766,next=266i,wt=1i 1514454299000000000
openntpd_sensors,sensor=nmea0,refid=GPS,state_prefix=* wt=10i,gd=1i,st=0i,next=1i,poll=15i,offset=-0.673,correction=0.6 1514454299000000000
openntpd_status peers_valid=12i,peers_total=12i,sensors_valid=1i,sensors_total=1i,constraint_offset_s=-1i,clock_synced=1i,stratum=1i 1514454299000000000
```
