# Ipset Input Plugin

This plugin gathers packets and bytes counters from [Linux IP sets][ipsets]
using the `ipset` command line tool.

> [!NOTE]
> IP sets created without the "counters" option are ignored.

⭐ Telegraf v1.6.0
🏷️ network, system
💻 linux

[ipsets]: https://ipset.netfilter.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather packets and bytes counters from Linux ipsets
  [[inputs.ipset]]
    ## By default, we only show sets which have already matched at least 1 packet.
    ## set include_unmatched_sets = true to gather them all.
    # include_unmatched_sets = false

    ## Adjust your sudo settings appropriately if using this option ("sudo ipset save")
    ## You can avoid using sudo or root, by setting appropriate privileges for
    ## the telegraf.service systemd service.
    # use_sudo = false

    ## Add number of entries and number of individual IPs (resolve CIDR syntax) for each ipset
    # count_per_ip_entries = false

    ## The default timeout of 1s for ipset execution can be overridden here:
    # timeout = "1s"
```

### Permissions

There are 3 ways to grant telegraf the right to run ipset:

- Run as root (strongly discouraged)
- Use sudo
- Configure systemd to run telegraf with CAP_NET_ADMIN and CAP_NET_RAW
  capabilities

#### Using sudo

To use sudo set the `use_sudo` option to `true` and update your sudoers file:

```bash
$ visudo
# Add the following line:
Cmnd_Alias IPSETSAVE = /sbin/ipset save
telegraf  ALL=(root) NOPASSWD: IPSETSAVE
Defaults!IPSETSAVE !logfile, !syslog, !pam_session
```

#### Using systemd capabilities

You may run `systemctl edit telegraf.service` and add the following:

```text
[Service]
CapabilityBoundingSet=CAP_NET_RAW CAP_NET_ADMIN
AmbientCapabilities=CAP_NET_RAW CAP_NET_ADMIN
```

## Metrics

- ipset
  - tags:
    - rule
    - set
  - fields:
    - timeout
    - packets
    - bytes

- ipset (for `count_per_ip_entries = true`)
  - tags:
    - set
  - fields:
    - entries
    - ips

## Example Output

```sh
$ sudo ipset save
create myset hash:net family inet hashsize 1024 maxelem 65536 counters comment
add myset 10.69.152.1 packets 8 bytes 672 comment "machine A"
```

```text
ipset,rule=10.69.152.1,host=trashme,set=myset bytes_total=8i,packets_total=672i 1507615028000000000
```
