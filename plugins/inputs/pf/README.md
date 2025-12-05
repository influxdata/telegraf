# PF Input Plugin

This plugin gathers information from the FreeBSD or OpenBSD pf firewall like
the number of current entries in the table, counters for the number of searches,
inserts, and removals to tables using the `pfctl` command.

> [!NOTE]
> This plugin requires the `pfctl` binary to be executable by Telegraf. It
> requires read access to the device file `/dev/pf`.

⭐ Telegraf v1.5.0
🏷️ system, network
💻 freebsd

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather counters from PF
[[inputs.pf]]
  ## PF require root access on most systems.
  ## Setting 'use_sudo' to true will make use of sudo to run pfctl.
  ## Users must configure sudo to allow telegraf user to run pfctl with no password.
  ## pfctl can be restricted to only list command "pfctl -s info".
  use_sudo = false
```

### Permissions

You have several options to grant Telegraf the permissions to run `pfctl`:

- Run telegraf as root. This is strongly discouraged.
- Change the ownership and permissions for `/dev/pf` to allow being read by the
  Telegraf user. This is discouraged.
- Configure sudo to allow running `pfctl` as root by the Telegraf user.
  This is the most restrictive option, but require sudo setup.
- Add the Telegraf user to the `proxy` group as `/dev/pf`.

For the `sudo` option you may add the following to the sudo configuration:

```sudo
telegraf ALL=(root) NOPASSWD: /sbin/pfctl -s info
```

## Metrics

- pf
  - entries (integer, count)
  - searches (integer, count)
  - inserts (integer, count)
  - removals (integer, count)
  - match (integer, count)
  - bad-offset (integer, count)
  - fragment (integer, count)
  - short (integer, count)
  - normalize (integer, count)
  - memory (integer, count)
  - bad-timestamp (integer, count)
  - congestion (integer, count)
  - ip-option (integer, count)
  - proto-cksum (integer, count)
  - state-mismatch (integer, count)
  - state-insert (integer, count)
  - state-limit (integer, count)
  - src-limit (integer, count)
  - synproxy (integer, count)

## Example Output

```shell
> pfctl -s info
Status: Enabled for 0 days 00:26:05           Debug: Urgent

State Table                          Total             Rate
  current entries                        2
  searches                           11325            7.2/s
  inserts                                5            0.0/s
  removals                               3            0.0/s
Counters
  match                              11226            7.2/s
  bad-offset                             0            0.0/s
  fragment                               0            0.0/s
  short                                  0            0.0/s
  normalize                              0            0.0/s
  memory                                 0            0.0/s
  bad-timestamp                          0            0.0/s
  congestion                             0            0.0/s
  ip-option                              0            0.0/s
  proto-cksum                            0            0.0/s
  state-mismatch                         0            0.0/s
  state-insert                           0            0.0/s
  state-limit                            0            0.0/s
  src-limit                              0            0.0/s
  synproxy                               0            0.0/s
```

```text
pf,host=columbia entries=3i,searches=2668i,inserts=12i,removals=9i 1510941775000000000
```
