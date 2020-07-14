# Network Interface Name Processor Plugin

The `ifname` plugin looks up network interface names using SNMP.

### Configuration:

```toml
[[processors.ifname]]
  ## Name of tag holding the interface number
  # tag = "ifIndex"

  ## Name of output tag where service name will be added
  # dest = "ifName"

  ## Name of tag of the SNMP agent to request the interface name from
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

### Example processing:

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
