Here are a few configuration examples for different use cases.

### Switch/router interface metrics

This setup will collect data on all interfaces from three different tables, `IF-MIB::ifTable`, `IF-MIB::ifXTable` and `EtherLike-MIB::dot3StatsTable`. It will also add the name from `IF-MIB::ifDescr` and use that as a tag. Depending on your needs and preferences you can easily use `IF-MIB::ifName` or `IF-MIB::ifAlias` instead or in addition. The values of these are typically:

    IF-MIB::ifName = Gi0/0/0
    IF-MIB::ifDescr = GigabitEthernet0/0/0
    IF-MIB::ifAlias = ### LAN ###

This configuration also collects the hostname from the device (`RFC1213-MIB::sysName.0`) and adds as a tag. So each metric will both have the configured host/IP as `agent_host` as well as the device self-reported hostname as `hostname` and the name of the host that has collected these metrics as `host`.

Here is the configuration that you add to your `telegraf.conf`:

```
[[inputs.snmp]]
  agents = [ "host.example.com" ]
  version = 2
  community = "public"

  [[inputs.snmp.field]]
    name = "hostname"
    oid = "RFC1213-MIB::sysName.0"
    is_tag = true

  [[inputs.snmp.field]]
    name = "uptime"
    oid = "DISMAN-EXPRESSION-MIB::sysUpTimeInstance"

  # IF-MIB::ifTable contains counters on input and output traffic as well as errors and discards.
  [[inputs.snmp.table]]
    name = "interface"
    inherit_tags = [ "hostname" ]
    oid = "IF-MIB::ifTable"

    # Interface tag - used to identify interface in metrics database
    [[inputs.snmp.table.field]]
      name = "ifDescr"
      oid = "IF-MIB::ifDescr"
      is_tag = true

  # IF-MIB::ifXTable contains newer High Capacity (HC) counters that do not overflow as fast for a few of the ifTable counters
  [[inputs.snmp.table]]
    name = "interface"
    inherit_tags = [ "hostname" ]
    oid = "IF-MIB::ifXTable"

    # Interface tag - used to identify interface in metrics database
    [[inputs.snmp.table.field]]
      name = "ifDescr"
      oid = "IF-MIB::ifDescr"
      is_tag = true

  # EtherLike-MIB::dot3StatsTable contains detailed ethernet-level information about what kind of errors have been logged on an interface (such as FCS error, frame too long, etc)
  [[inputs.snmp.table]]
    name = "interface"
    inherit_tags = [ "hostname" ]
    oid = "EtherLike-MIB::dot3StatsTable"

    # Interface tag - used to identify interface in metrics database
    [[inputs.snmp.table.field]]
      name = "ifDescr"
      oid = "IF-MIB::ifDescr"
      is_tag = true
```
