# SNMP Input Plugin

The `snmp` input plugin uses polling to gather metrics from SNMP agents.
Support for gathering individual OIDs as well as complete SNMP tables is
included.

## Note about Paths

Path is a global variable, separate snmp instances will append the specified
path onto the global path variable

## Configuration

```toml
# Retrieves SNMP values from remote agents
[[inputs.snmp]]
  ## Agent addresses to retrieve values from.
  ##   format:  agents = ["<scheme://><hostname>:<port>"]
  ##   scheme:  optional, either udp, udp4, udp6, tcp, tcp4, tcp6.
  ##            default is udp
  ##   port:    optional
  ##   example: agents = ["udp://127.0.0.1:161"]
  ##            agents = ["tcp://127.0.0.1:161"]
  ##            agents = ["udp4://v4only-snmp-agent"]
  agents = ["udp://127.0.0.1:161"]

  ## Timeout for each request.
  # timeout = "5s"

  ## SNMP version; can be 1, 2, or 3.
  # version = 2

  ## Path to mib files
  ## Used by the gosmi translator.
  ## To add paths when translating with netsnmp, use the MIBDIRS environment variable
  # path = ["/usr/share/snmp/mibs"]

  ## SNMP community string.
  # community = "public"

  ## Agent host tag
  # agent_host_tag = "agent_host"

  ## Number of retries to attempt.
  # retries = 3

  ## The GETBULK max-repetitions parameter.
  # max_repetitions = 10

  ## SNMPv3 authentication and encryption options.
  ##
  ## Security Name.
  # sec_name = "myuser"
  ## Authentication protocol; one of "MD5", "SHA", "SHA224", "SHA256", "SHA384", "SHA512" or "".
  # auth_protocol = "MD5"
  ## Authentication password.
  # auth_password = "pass"
  ## Security Level; one of "noAuthNoPriv", "authNoPriv", or "authPriv".
  # sec_level = "authNoPriv"
  ## Context Name.
  # context_name = ""
  ## Privacy protocol used for encrypted messages; one of "DES", "AES", "AES192", "AES192C", "AES256", "AES256C", or "".
  ### Protocols "AES192", "AES192", "AES256", and "AES256C" require the underlying net-snmp tools
  ### to be compiled with --enable-blumenthal-aes (http://www.net-snmp.org/docs/INSTALL.html)
  # priv_protocol = ""
  ## Privacy password used for encrypted messages.
  # priv_password = ""

  ## Add fields and tables defining the variables you wish to collect.  This
  ## example collects the system uptime and interface variables.  Reference the
  ## full plugin documentation for configuration details.
  [[inputs.snmp.field]]
    oid = "RFC1213-MIB::sysUpTime.0"
    name = "uptime"

  [[inputs.snmp.field]]
    oid = "RFC1213-MIB::sysName.0"
    name = "source"
    is_tag = true

  [[inputs.snmp.table]]
    oid = "IF-MIB::ifTable"
    name = "interface"
    inherit_tags = ["source"]

    [[inputs.snmp.table.field]]
      oid = "IF-MIB::ifDescr"
      name = "ifDescr"
      is_tag = true
```

### Configure SNMP Requests

This plugin provides two methods for configuring the SNMP requests: `fields`
and `tables`.  Use the `field` option to gather single ad-hoc variables.
To collect SNMP tables, use the `table` option.

#### Field

Use a `field` to collect a variable by OID.  Requests specified with this
option operate similar to the `snmpget` utility.

```toml
[[inputs.snmp]]
  # ... snip ...

  [[inputs.snmp.field]]
    ## Object identifier of the variable as a numeric or textual OID.
    oid = "RFC1213-MIB::sysName.0"

    ## Name of the field or tag to create.  If not specified, it defaults to
    ## the value of 'oid'. If 'oid' is numeric, an attempt to translate the
    ## numeric OID into a textual OID will be made.
    # name = ""

    ## If true the variable will be added as a tag, otherwise a field will be
    ## created.
    # is_tag = false

    ## Apply one of the following conversions to the variable value:
    ##   float(X):    Convert the input value into a float and divides by the
    ##                Xth power of 10. Effectively just moves the decimal left
    ##                X places. For example a value of `123` with `float(2)`
    ##                will result in `1.23`.
    ##   float:       Convert the value into a float with no adjustment. Same
    ##                as `float(0)`.
    ##   int:         Convert the value into an integer.
    ##   hwaddr:      Convert the value to a MAC address.
    ##   ipaddr:      Convert the value to an IP address.
    ##   hextoint:X:Y Convert a hex string value to integer. Where X is the Endian
    ##                and Y the bit size. For example: hextoint:LittleEndian:uint64
    ##                or hextoint:BigEndian:uint32. Valid options for the Endian are:
    ##                BigEndian and LittleEndian. For the bit size: uint16, uint32
    ##                and uint64.
    ##
    # conversion = ""
```

#### Table

Use a `table` to configure the collection of a SNMP table.  SNMP requests
formed with this option operate similarly way to the `snmptable` command.

Control the handling of specific table columns using a nested `field`.  These
nested fields are specified similarly to a top-level `field`.

By default all columns of the SNMP table will be collected - it is not required
to add a nested field for each column, only those which you wish to modify. To
*only* collect certain columns, omit the `oid` from the `table` section and only
include `oid` settings in `field` sections. For more complex include/exclude
cases for columns use [metric filtering][].

One [metric][] is created for each row of the SNMP table.

```toml
[[inputs.snmp]]
  # ... snip ...

  [[inputs.snmp.table]]
    ## Object identifier of the SNMP table as a numeric or textual OID.
    oid = "IF-MIB::ifTable"

    ## Name of the field or tag to create.  If not specified, it defaults to
    ## the value of 'oid'.  If 'oid' is numeric an attempt to translate the
    ## numeric OID into a textual OID will be made.
    # name = ""

    ## Which tags to inherit from the top-level config and to use in the output
    ## of this table's measurement.
    ## example: inherit_tags = ["source"]
    # inherit_tags = []

    ## Add an 'index' tag with the table row number.  Use this if the table has
    ## no indexes or if you are excluding them.  This option is normally not
    ## required as any index columns are automatically added as tags.
    # index_as_tag = false

    [[inputs.snmp.table.field]]
      ## OID to get. May be a numeric or textual module-qualified OID.
      oid = "IF-MIB::ifDescr"

      ## Name of the field or tag to create.  If not specified, it defaults to
      ## the value of 'oid'. If 'oid' is numeric an attempt to translate the
      ## numeric OID into a textual OID will be made.
      # name = ""

      ## Output this field as a tag.
      # is_tag = false

      ## The OID sub-identifier to strip off so that the index can be matched
      ## against other fields in the table.
      # oid_index_suffix = ""

      ## Specifies the length of the index after the supplied table OID (in OID
      ## path segments). Truncates the index after this point to remove non-fixed
      ## value or length index suffixes.
      # oid_index_length = 0

      ## Specifies if the value of given field should be snmptranslated
      ## by default no field values are translated
      # translate = true

      ## Secondary index table allows to merge data from two tables with
      ## different index that this filed will be used to join them. There can
      ## be only one secondary index table.
      # secondary_index_table = false

      ## This field is using secondary index, and will be later merged with
      ## primary index using SecondaryIndexTable. SecondaryIndexTable and
      ## SecondaryIndexUse are exclusive.
      # secondary_index_use = false

      ## Controls if entries from secondary table should be added or not
      ## if joining index is present or not. I set to true, means that join
      ## is outer, and index is prepended with "Secondary." for missing values
      ## to avoid overlaping indexes from both tables. Can be set per field or
      ## globally with SecondaryIndexTable, global true overrides per field false.
      # secondary_outer_join = false
```

#### Two Table Join

Snmp plugin can join two snmp tables that have different indexes. For this to work one table
should have translation field that return index of second table as value. Examples
of such fields are:

* Cisco portTable with translation field: `CISCO-STACK-MIB::portIfIndex`,
which value is IfIndex from ifTable
* Adva entityFacilityTable with translation field: `ADVA-FSPR7-MIB::entityFacilityOneIndex`,
which value is IfIndex from ifTable
* Cisco cpeExtPsePortTable with translation field: `CISCO-POWER-ETHERNET-EXT-MIB::cpeExtPsePortEntPhyIndex`,
which value is index from entPhysicalTable

Such field can be used to translate index to secondary table with `secondary_index_table = true`
and all fields from secondary table (with index pointed from translation field), should have added option
`secondary_index_use = true`. Telegraf cannot duplicate entries during join so translation
must be 1-to-1 (not 1-to-many). To add fields from secondary table with index that is not present
in translation table (outer join), there is a second option for translation index `secondary_outer_join = true`.

##### Example configuration for table joins

CISCO-POWER-ETHERNET-EXT-MIB table before join:

```toml
[[inputs.snmp.table]]
name = "ciscoPower"
index_as_tag = true

[[inputs.snmp.table.field]]
name = "PortPwrConsumption"
oid = "CISCO-POWER-ETHERNET-EXT-MIB::cpeExtPsePortPwrConsumption"

[[inputs.snmp.table.field]]
name = "EntPhyIndex"
oid = "CISCO-POWER-ETHERNET-EXT-MIB::cpeExtPsePortEntPhyIndex"
```

Partial result (removed agent_host and host columns from all following outputs in this section):

```text
> ciscoPower,index=1.2 EntPhyIndex=1002i,PortPwrConsumption=6643i 1621460628000000000
> ciscoPower,index=1.6 EntPhyIndex=1006i,PortPwrConsumption=10287i 1621460628000000000
> ciscoPower,index=1.5 EntPhyIndex=1005i,PortPwrConsumption=8358i 1621460628000000000
```

Note here that EntPhyIndex column carries index from ENTITY-MIB table, config for it:

```toml
[[inputs.snmp.table]]
name = "entityTable"
index_as_tag = true

[[inputs.snmp.table.field]]
name = "EntPhysicalName"
oid = "ENTITY-MIB::entPhysicalName"
```

Partial result:

```text
> entityTable,index=1006 EntPhysicalName="GigabitEthernet1/6" 1621460809000000000
> entityTable,index=1002 EntPhysicalName="GigabitEthernet1/2" 1621460809000000000
> entityTable,index=1005 EntPhysicalName="GigabitEthernet1/5" 1621460809000000000
```

Now, lets attempt to join these results into one table. EntPhyIndex matches index
from second table, and lets convert EntPhysicalName into tag, so second table will
only provide tags into result. Configuration:

```toml
[[inputs.snmp.table]]
name = "ciscoPowerEntity"
index_as_tag = true

[[inputs.snmp.table.field]]
name = "PortPwrConsumption"
oid = "CISCO-POWER-ETHERNET-EXT-MIB::cpeExtPsePortPwrConsumption"

[[inputs.snmp.table.field]]
name = "EntPhyIndex"
oid = "CISCO-POWER-ETHERNET-EXT-MIB::cpeExtPsePortEntPhyIndex"
secondary_index_table = true    # enables joining

[[inputs.snmp.table.field]]
name = "EntPhysicalName"
oid = "ENTITY-MIB::entPhysicalName"
secondary_index_use = true      # this tag is indexed from secondary table
is_tag = true
```

Result:

```text
> ciscoPowerEntity,EntPhysicalName=GigabitEthernet1/2,index=1.2 EntPhyIndex=1002i,PortPwrConsumption=6643i 1621461148000000000
> ciscoPowerEntity,EntPhysicalName=GigabitEthernet1/6,index=1.6 EntPhyIndex=1006i,PortPwrConsumption=10287i 1621461148000000000
> ciscoPowerEntity,EntPhysicalName=GigabitEthernet1/5,index=1.5 EntPhyIndex=1005i,PortPwrConsumption=8358i 1621461148000000000
```

## Troubleshooting

Check that a numeric field can be translated to a textual field:

```sh
$ snmptranslate .1.3.6.1.2.1.1.3.0
DISMAN-EVENT-MIB::sysUpTimeInstance
```

Request a top-level field:

```sh
snmpget -v2c -c public 127.0.0.1 sysUpTime.0
```

Request a table:

```sh
snmptable -v2c -c public 127.0.0.1 ifTable
```

To collect a packet capture, run this command in the background while running
Telegraf or one of the above commands.  Adjust the interface, host and port as
needed:

```sh
sudo tcpdump -s 0 -i eth0 -w telegraf-snmp.pcap host 127.0.0.1 and port 161
```

## Example Output

```shell
snmp,agent_host=127.0.0.1,source=loaner uptime=11331974i 1575509815000000000
interface,agent_host=127.0.0.1,ifDescr=wlan0,ifIndex=3,source=example.org ifAdminStatus=1i,ifInDiscards=0i,ifInErrors=0i,ifInNUcastPkts=0i,ifInOctets=3436617431i,ifInUcastPkts=2717778i,ifInUnknownProtos=0i,ifLastChange=0i,ifMtu=1500i,ifOperStatus=1i,ifOutDiscards=0i,ifOutErrors=0i,ifOutNUcastPkts=0i,ifOutOctets=581368041i,ifOutQLen=0i,ifOutUcastPkts=1354338i,ifPhysAddress="c8:5b:76:c9:e6:8c",ifSpecific=".0.0",ifSpeed=0i,ifType=6i 1575509815000000000
interface,agent_host=127.0.0.1,ifDescr=eth0,ifIndex=2,source=example.org ifAdminStatus=1i,ifInDiscards=0i,ifInErrors=0i,ifInNUcastPkts=21i,ifInOctets=3852386380i,ifInUcastPkts=3634004i,ifInUnknownProtos=0i,ifLastChange=9088763i,ifMtu=1500i,ifOperStatus=1i,ifOutDiscards=0i,ifOutErrors=0i,ifOutNUcastPkts=0i,ifOutOctets=434865441i,ifOutQLen=0i,ifOutUcastPkts=2110394i,ifPhysAddress="c8:5b:76:c9:e6:8c",ifSpecific=".0.0",ifSpeed=1000000000i,ifType=6i 1575509815000000000
interface,agent_host=127.0.0.1,ifDescr=lo,ifIndex=1,source=example.org ifAdminStatus=1i,ifInDiscards=0i,ifInErrors=0i,ifInNUcastPkts=0i,ifInOctets=51555569i,ifInUcastPkts=339097i,ifInUnknownProtos=0i,ifLastChange=0i,ifMtu=65536i,ifOperStatus=1i,ifOutDiscards=0i,ifOutErrors=0i,ifOutNUcastPkts=0i,ifOutOctets=51555569i,ifOutQLen=0i,ifOutUcastPkts=339097i,ifSpecific=".0.0",ifSpeed=10000000i,ifType=24i 1575509815000000000
```

[metric filtering]: /docs/CONFIGURATION.md#metric-filtering
[metric]: /docs/METRICS.md
