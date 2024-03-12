# SNMP Lookup Processor Plugin

The `snmp_lookup` plugin looks up extra tags using SNMP and caches them.

Telegraf minimum version: Telegraf 1.30.0

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `auth_password` and
`priv_password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Lookup extra tags via SNMP based on the table index
[[processors.snmp_lookup]]
  ## Name of tag of the SNMP agent to do the lookup on
  # agent_tag = "source"

  ## Name of tag holding the table row index
  # index_tag = "index"

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

  ## The maximum number of SNMP requests to make at the same time.
  # max_parallel_lookups = 16

  ## The amount of agents to cache entries for. If limit is reached, 
  ## oldest will be removed first. 0 means no limit.
  # max_cache_entries = 100

  ## Control whether the metrics need to stay in the same order this plugin
  ## received them in. If false, this plugin may change the order when data is
  ## cached. If you need metrics to stay in order set this to true. Keeping the
  ## metrics ordered may be slightly slower.
  # ordered = false

  ## The amount of time entries are cached for a given agent. After this period
  ## elapses if tags are needed they will be retrieved again.
  # cache_ttl = "8h"

  ## Minimum time between requests to an agent in case an index could not be
  ## resolved. If set to zero no request on missing indices will be triggered.
  # min_time_between_updates = "5m"

  ## List of tags to be looked up.
  [[processors.snmp_lookup.tag]]
    ## Object identifier of the variable as a numeric or textual OID.
    oid = "IF-MIB::ifName"

    ## Name of the tag to create.  If not specified, it defaults to the value of 'oid'.
    ## If 'oid' is numeric, an attempt to translate the numeric OID into a textual OID
    ## will be made.
    # name = ""

    ## Apply one of the following conversions to the variable value:
    ##   hwaddr:  Convert the value to a MAC address.
    ##   ipaddr:  Convert the value to an IP address.
    ##   enum(1): Convert the value according to its syntax in the MIB (full).
    ##   enum:    Convert the value according to its syntax in the MIB.
    ##
    # conversion = ""
```

## Examples

### Sample config

```diff
- foo,index=2,source=127.0.0.1 field=123
+ foo,ifName=eth0,index=2,source=127.0.0.1 field=123
```

### processors.ifname replacement

The following config will use the same `ifDescr` fallback as `processors.ifname`
when there is no `ifName` value on the device.

```toml
[[processors.snmp_lookup]]
  agent_tag = "agent"
  index_tag = "ifIndex"

  [[processors.snmp_lookup.tag]]
    oid = ".1.3.6.1.2.1.2.2.1.2"
    name = "ifName"

  [[processors.snmp_lookup.tag]]
    oid = ".1.3.6.1.2.1.31.1.1.1.1"
    name = "ifName"
```

```diff
- foo,agent=127.0.0.1,ifIndex=2 field=123
+ foo,agent=127.0.0.1,ifIndex=2,ifName=eth0 field=123
```
