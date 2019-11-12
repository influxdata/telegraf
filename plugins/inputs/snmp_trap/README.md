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
	- source (string, IP address of trap source)
	- trap_name (string, value from SNMPv2-MIB::snmpTrapOID.0 PDU)
	- trap_mib (string, mib from SNMPv2-MIB::snmpTrapOID.0 PDU)
	- trap_oid (string, oid string from SNMPv2-MIB::snmpTrapOID.0 PDU)
	- trap_version (string, "1" or "2c" or "3")
  - fields:
	- $NAME (the type is variable and depends on the PDU)
	- $NAME_type (string, description of the Asn1BER type of the PDU.  Examples: "Integer", "TimeTicks", "IPAddress")

### Example Output
```
snmp_trap,source=192.168.122.102,trap_mib=SNMPv2-MIB,trap_name=coldStart,trap_oid=.1.3.6.1.6.3.1.1.5.1,trap_version=2c sysUpTimeInstance=0i,sysUpTimeInstance_type="TimeTicks",snmpTrapEnterprise.0="linux",snmpTrapEnterprise.0_type="ObjectIdentifier" 1573586012665359513
snmp_trap,source=192.168.122.102,trap_mib=NET-SNMP-AGENT-MIB,trap_name=nsNotifyShutdown,trap_oid=.1.3.6.1.4.1.8072.4.0.2,trap_version=2c sysUpTimeInstance=2196i,sysUpTimeInstance_type="TimeTicks",snmpTrapEnterprise.0="netSnmpNotificationPrefix",snmpTrapEnterprise.0_type="ObjectIdentifier" 1573586012076284951
```
