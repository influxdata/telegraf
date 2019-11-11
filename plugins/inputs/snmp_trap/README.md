# SNMP Trap Input Plugin

The SNMP Trap plugin is a service input plugin that receives SNMP
notifications (traps and inform requests).

Notifications are received on plain UDP. The port to listen is
configurable.

OIDs can be resolved to strings using system MIB files. This is done
in same way as the SNMP input plugin. See the section "MIB Lookups" in
the SNMP [README.md](../snmp/README.md) for details.

### Configuration
```toml
  ## Local address and port to listen on.  Omit address to listen on
  ## all interfaces.  Example "127.0.0.1:1234", default ":162"
  #service_address = :162
```

### Metrics

- snmp_trap
  - tags:
	- trap_name (string, value from SNMPv2-MIB::snmpTrapOID.0 PDU)
	- trap_version (string, "1" or "2c" or "3")
  - fields:
	- $NAME (the type is variable and depends on the PDU)
	- $NAME_type (string, description of the Asn1BER type of the PDU.  Examples: "Integer", "TimeTicks", "IPAddress")

### Example Output
```
> snmp_trap,host=debian,trap_name=coldStart,trap_version=2c sysUpTimeInstance_type="TimeTicks",snmpTrapEnterprise.0="linux",snmpTrapEnterprise.0_type="ObjectIdentifier",sysUpTimeInstance=1i 1573078928344595213
> snmp_trap,host=debian,trap_name=nsNotifyShutdown,trap_version=2c sysUpTimeInstance=1224i,sysUpTimeInstance_type="TimeTicks",snmpTrapEnterprise.0="netSnmpNotificationPrefix",snmpTrapEnterprise.0_type="ObjectIdentifier" 1573078928299642679
```
