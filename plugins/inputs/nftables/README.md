# Nftables Plugin

This plugin gathers packets and bytes counters for rules within
Linux's [nftables][nftables] firewall.

Rules are identified through associated comment.
**Rules without comment are ignored**.

Before using this plugin **you must ensure that the rules you want to monitor
are named with a unique comment**. Comments are added using the 'comment
"my comment"' nftables options.


⭐ Telegraf v1.37.0
🏷️ network, system
💻 linux

[nftables]: https://wiki.nftables.org/wiki-nftables/index.php/Main_Page

## Configuration

```toml @sample.conf
[[inputs.nftables]]
  ## Use sudo for command execution, can be restricted to "nft --json list table"
  # use_sudo = false

  ## Use this alternative binary
  ## By default the nft command in PATH is used
  # binary = "nft"

  ## A List of nftables to monitor. 
  ## have a counter and comment declared on it.
  # tables = [ "filter" ]
```

Since telegraf will fork a process to run nftables, `AmbientCapabilities` is
required to transmit the capabilities bounding set to the forked process.

### Using sudo

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

## Metrics

* nftables
  * tags:
    * table
    * chain
    * ruleid -- comment associated to the rule
  * fields:
    * pkts (integer, count)
    * bytes (integer, bytes)

## Example Output

```text
> nftables,chain=incoming,host=my_hostname,ruleid=comment_val_1,table=filter bytes=66435845i,pkts=133882i 1757367516000000000
> nftables,chain=outgoing,host=my_hostname,ruleid=comment_val2,table=filter bytes=25596512i,pkts=145129i 1757367516000000000
```
