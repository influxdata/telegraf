# Iptables Input Plugin

This plugin gathers packets and bytes counters for rules within a set of table
and chain from the Linux's iptables firewall.

> [!IMPORTANT]
> Rules are identified through associated comment, so you must ensure that the
> rules you want to monitor do have a **unique** comment using the `--comment`
> flag when adding them. Rules without comments are ignored.

The rule number cannot be used as identifier as it is not constant and
may vary when rules are inserted/deleted at start-up or by automatic tools
(interactive firewalls, fail2ban, ...).

> [!IMPORTANT]
> The `iptables` command requires `CAP_NET_ADMIN` and `CAP_NET_RAW`
> capabilities. Check the [permissions section](#permissions) for ways to
> grant them.

⭐ Telegraf v1.1.0
🏷️ network, system
💻 linux

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather packets and bytes throughput from iptables
# This plugin ONLY supports Linux
[[inputs.iptables]]
  ## iptables require root access on most systems.
  ## Setting 'use_sudo' to true will make use of sudo to run iptables.
  ## Users must configure sudo to allow telegraf user to run iptables with
  ## no password.
  ## iptables can be restricted to only list command "iptables -nvL".
  # use_sudo = false

  ## Setting 'use_lock' to true runs iptables with the "-w" option.
  ## Adjust your sudo settings appropriately if using this option
  ## ("iptables -w 5 -nvl")
  # use_lock = false

  ## Define an alternate executable, such as "ip6tables". Default is "iptables".
  # binary = "ip6tables"
  ## defines the table to monitor:
  table = "filter"

  ## defines the chains to monitor.
  ## NOTE: iptables rules without a comment will not be monitored.
  ## Read the plugin documentation for more information.
  chains = [ "INPUT" ]
```

### Permissions

The `iptables` command requires `CAP_NET_ADMIN` and `CAP_NET_RAW capabilities`.
You have several options to grant permissions to telegraf:

- Run telegraf as root. This is strongly discouraged.
- Configure systemd to run telegraf with CAP_NET_ADMIN and CAP_NET_RAW. This is
  the simplest and recommended option.
- Configure sudo to grant telegraf to run iptables. This is the most
  restrictive option, but require sudo setup.

#### Using systemd capabilities

You may run `systemctl edit telegraf.service` and add the following:

```shell
[Service]
CapabilityBoundingSet=CAP_NET_RAW CAP_NET_ADMIN
AmbientCapabilities=CAP_NET_RAW CAP_NET_ADMIN
```

Since telegraf will fork a process to run iptables, `AmbientCapabilities` is
required to transmit the capabilities bounding set to the forked process.

#### Using sudo

To use sudo set the `use_sudo` option to `true` and update your sudoers file:

```bash
$ visudo
# Add the following line:
Cmnd_Alias IPTABLESSHOW = /usr/bin/iptables -nvL *
telegraf  ALL=(root) NOPASSWD: IPTABLESSHOW
Defaults!IPTABLESSHOW !logfile, !syslog, !pam_session
```

### Using IPtables lock feature

Defining multiple instances of this plugin in telegraf.conf can lead to
concurrent IPtables access resulting in "ERROR in input [inputs.iptables]: exit
status 4" messages in telegraf.log and missing metrics. Setting 'use_lock =
true' in the plugin configuration will run IPtables with the '-w' switch,
allowing a lock usage to prevent this error.

## Metrics

- iptables
  - tags:
    - table
    - chain
    - ruleid (comment associated to the rule)
  - fields:
    - pkts (integer, count)
    - bytes (integer, bytes)

## Example Output

```shell
iptables -nvL INPUT
```

```text
Chain INPUT (policy DROP 0 packets, 0 bytes)
pkts bytes target     prot opt in     out     source               destination
100   1024   ACCEPT     tcp  --  *      *       192.168.0.0/24       0.0.0.0/0            tcp dpt:22 /* ssh */
 42   2048   ACCEPT     tcp  --  *      *       192.168.0.0/24       0.0.0.0/0            tcp dpt:80 /* httpd */
```

```text
iptables,table=filter,chain=INPUT,ruleid=ssh pkts=100i,bytes=1024i 1453831884664956455
iptables,table=filter,chain=INPUT,ruleid=httpd pkts=42i,bytes=2048i 1453831884664956455
```
