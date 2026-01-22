# Nftables Plugin

This plugin gathers packets and bytes counters for rules within
Linux's [nftables][nftables] firewall, as well as set element counts.

> [!IMPORTANT]
> Rules are identified by the associated comment so those **comments have to be
> unique**! Rules without comment are ignored.

⭐ Telegraf v1.37.0
🏷️ network, system
💻 linux

[nftables]: https://wiki.nftables.org/wiki-nftables/index.php/Main_Page

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
[[inputs.nftables]]
  ## Use the specified binary which will be looked-up in PATH
  # binary = "nft"

  ## Use sudo for command execution, can be restricted to
  ## "nft --json list table"
  # use_sudo = false

  ## Tables to monitor (may use "family table" format, e.g., "inet filter")
  # tables = [ "filter" ]
```

Since telegraf will fork a process to run nftables, `AmbientCapabilities` is
required to transmit the capabilities bounding set to the forked process.

### Using sudo

You may edit your sudo configuration with the following:

```sudo
telegraf ALL=(root) NOPASSWD: /usr/bin/nft *
```

## Metrics

Rules:

* nftables
  * tags:
    * table
    * chain
    * rule -- comment associated to the rule
  * fields:
    * pkts (integer, count)
    * bytes (integer, bytes)

Sets:

* nftables
  * tags:
    * table
    * set
  * field:
    * count (integer, count)

## Example Output

```text
> nftables,chain=incoming,host=my_hostname,rule=comment_val_1,table=filter bytes=66435845i,pkts=133882i 1757367516000000000
> nftables,chain=outgoing,host=my_hostname,rule=comment_val2,table=filter bytes=25596512i,pkts=145129i 1757367516000000000
> nftables,host=my_hostname,set=my_set,table=filter count=10i 1757367516000000000
```
