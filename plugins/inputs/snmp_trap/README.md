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
# Snmp trap listener
[[inputs.snmp_trap]]
  ## Transport, local address, and port to listen on.  Transport must
  ## be "udp://".  Omit local address to listen on all interfaces.
  ##   example: "udp://127.0.0.1:1234"
  # service_address = udp://:162
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
snmp_trap,mib=SNMPv2-MIB,name=coldStart,oid=.1.3.6.1.6.3.1.1.5.1,source=192.168.122.102,version=2c snmpTrapEnterprise.0="linux",sysUpTimeInstance=1i 1574109187723429814
snmp_trap,mib=NET-SNMP-AGENT-MIB,name=nsNotifyShutdown,oid=.1.3.6.1.4.1.8072.4.0.2,source=192.168.122.102,version=2c sysUpTimeInstance=5803i,snmpTrapEnterprise.0="netSnmpNotificationPrefix" 1574109186555115459
```
