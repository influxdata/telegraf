# Nftables Plugin

The nftables plugin gathers packets and bytes counters for rules within
Linux's nftables firewall.

Rules are identified through associated comment.
**Rules without comment are ignored**.

Before using this plugin **you must ensure that the rules you want to monitor
are named with a unique comment**. Comments are added using the 'comment
"my comment"' nftables options.

The nftables command requires CAP_NET_ADMIN and CAP_NET_RAW capabilities.
You have several options to grant telegraf to run nftables:

* Run telegraf as root. This is strongly discouraged.
* Configure systemd to run telegraf with CAP_NET_ADMIN and CAP_NET_RAW.
This is the simplest and recommended option.
* Configure sudo to grant telegraf to run nftables. This is the most restrictive
 option, but require sudo setup.

⭐ Telegraf v1.1.0
🏷️  network, system
💻 linux

## Using systemd capabilities

You may run `systemctl edit telegraf.service` and add the following:

```text
[Service]
CapabilityBoundingSet=CAP_NET_RAW CAP_NET_ADMIN
AmbientCapabilities=CAP_NET_RAW CAP_NET_ADMIN
```

Since telegraf will fork a process to run nftables, `AmbientCapabilities` is
required to transmit the capabilities bounding set to the forked process.

## Using sudo

You may edit your sudo configuration with the following:

```sudo
telegraf ALL=(root) NOPASSWD: /usr/bin/nft *
```

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml
  # use sudo to run nftables
  use_sudo = false
  # defines the table to monitor:
  tables = ["filter"]
```

## Tags

* All measurements have the following tags:
  * table
  * chain
  * ruleid

The `ruleid` is the comment associated to the rule.

## Metrics

* nftables
  * pkts (integer, count)
  * bytes (integer, bytes)

## Example Output

```bash
$ ./telegraf --config telegraf.conf --input-filter nftables --test
> nftables,chain=incoming,host=my_hostname,ruleid=comment_val_1,table=filter bytes=66435845i,pkts=133882i 1757367516000000000
> nftables,chain=outgoing,host=my_hostname,ruleid=comment_val2,table=filter bytes=25596512i,pkts=145129i 1757367516000000000
```
