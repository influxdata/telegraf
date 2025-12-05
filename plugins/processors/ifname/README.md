# Network Interface Name Processor Plugin

This plugin looks up network interface names using SNMP.

⭐ Telegraf v1.15.0
🏷️ annotation
💻 all

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `auth_password` and
`priv_password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Add a tag of the network interface name looked up over SNMP by interface number
[[processors.ifname]]
  ## Name of tag holding the interface number
  # tag = "ifIndex"

  ## Name of output tag where service name will be added
  # dest = "ifName"

  ## Name of tag of the SNMP agent to request the interface name from
  ##   example: agent = "source"
  # agent = "agent"

  ## Timeout for each request.
  # timeout = "5s"

  ## SNMP version; can be 1, 2, or 3.
  # version = 2

  ## SNMP community string.
  # community = "public"

  ## Number of retries to attempt.
  # retries = 3

  ## The GETBULK max-repetitions parameter.
  # max_repetitions = 10

  ## SNMPv3 authentication and encryption options.
  ##
  ## Security Name.
  # sec_name = "myuser"
  ## Authentication protocol; one of "MD5", "SHA", or "".
  # auth_protocol = "MD5"
  ## Authentication password.
  # auth_password = "pass"
  ## Security Level; one of "noAuthNoPriv", "authNoPriv", or "authPriv".
  # sec_level = "authNoPriv"
  ## Context Name.
  # context_name = ""
  ## Privacy protocol used for encrypted messages; one of "DES", "AES" or "".
  # priv_protocol = ""
  ## Privacy password used for encrypted messages.
  # priv_password = ""

  ## max_parallel_lookups is the maximum number of SNMP requests to
  ## make at the same time.
  # max_parallel_lookups = 100

  ## ordered controls whether or not the metrics need to stay in the
  ## same order this plugin received them in. If false, this plugin
  ## may change the order when data is cached.  If you need metrics to
  ## stay in order set this to true.  keeping the metrics ordered may
  ## be slightly slower
  # ordered = false

  ## cache_ttl is the amount of time interface names are cached for a
  ## given agent.  After this period elapses if names are needed they
  ## will be retrieved again.
  # cache_ttl = "8h"
```

## Example

Example config:

```toml
[[processors.ifname]]
  tag = "ifIndex"
  dest = "ifName"
```

```diff
- foo,ifIndex=2,agent=127.0.0.1 field=123 1502489900000000000
+ foo,ifIndex=2,agent=127.0.0.1,ifName=eth0 field=123 1502489900000000000
```
