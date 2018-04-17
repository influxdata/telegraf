# SNMP Plugin

The SNMP input plugin gathers metrics from SNMP agents.

## Configuration:

See additional SNMP plugin configuration examples [here](./CONFIG-EXAMPLES.md).

### Example:

SNMP data:
```
.1.0.0.0.1.1.0 octet_str "foo"
.1.0.0.0.1.1.1 octet_str "bar"
.1.0.0.0.1.102 octet_str "bad"
.1.0.0.0.1.2.0 integer 1
.1.0.0.0.1.2.1 integer 2
.1.0.0.0.1.3.0 octet_str "0.123"
.1.0.0.0.1.3.1 octet_str "0.456"
.1.0.0.0.1.3.2 octet_str "9.999"
.1.0.0.1.1 octet_str "baz"
.1.0.0.1.2 uinteger 54321
.1.0.0.1.3 uinteger 234
```

Telegraf config:
```toml
[[inputs.snmp]]
  agents = [ "127.0.0.1:161" ]
  version = 2
  community = "public"

  name = "system"
  [[inputs.snmp.field]]
    name = "hostname"
    oid = ".1.0.0.1.1"
    is_tag = true
  [[inputs.snmp.field]]
    name = "uptime"
    oid = ".1.0.0.1.2"
  [[inputs.snmp.field]]
    name = "loadavg"
    oid = ".1.0.0.1.3"
    conversion = "float(2)"

  [[inputs.snmp.table]]
    name = "remote_servers"
    inherit_tags = [ "hostname" ]
    [[inputs.snmp.table.field]]
      name = "server"
      oid = ".1.0.0.0.1.1"
      is_tag = true
    [[inputs.snmp.table.field]]
      name = "connections"
      oid = ".1.0.0.0.1.2"
    [[inputs.snmp.table.field]]
      name = "latency"
      oid = ".1.0.0.0.1.3"
      conversion = "float"
```

Resulting output:
```
* Plugin: snmp, Collection 1
> system,agent_host=127.0.0.1,host=mylocalhost,hostname=baz loadavg=2.34,uptime=54321i 1468953135000000000
> remote_servers,agent_host=127.0.0.1,host=mylocalhost,hostname=baz,server=foo connections=1i,latency=0.123 1468953135000000000
> remote_servers,agent_host=127.0.0.1,host=mylocalhost,hostname=baz,server=bar connections=2i,latency=0.456 1468953135000000000
```

#### Configuration via MIB:

This example uses the SNMP data above, but is configured via the MIB.
The example MIB file can be found in the `testdata` directory. See the [MIB lookups](#mib-lookups) section for more information.

Telegraf config:
```toml
[[inputs.snmp]]
  agents = [ "127.0.0.1:161" ]
  version = 2
  community = "public"

  [[inputs.snmp.field]]
    oid = "TEST::hostname"
    is_tag = true

  [[inputs.snmp.table]]
    oid = "TEST::testTable"
    inherit_tags = [ "hostname" ]
```

Resulting output:
```
* Plugin: snmp, Collection 1
> testTable,agent_host=127.0.0.1,host=mylocalhost,hostname=baz,server=foo connections=1i,latency="0.123" 1468953135000000000
> testTable,agent_host=127.0.0.1,host=mylocalhost,hostname=baz,server=bar connections=2i,latency="0.456" 1468953135000000000
```

### Config parameters

* `agents`: Default: `[]`
List of SNMP agents to connect to in the form of `IP[:PORT]`. If `:PORT` is unspecified, it defaults to `161`.

* `version`: Default: `2`
SNMP protocol version to use.

* `community`: Default: `"public"`
SNMP community to use.

* `max_repetitions`: Default: `50`
Maximum number of iterations for repeating variables.

* `sec_name`:
Security name for authenticated SNMPv3 requests.

* `auth_protocol`: Values: `"MD5"`,`"SHA"`,`""`. Default: `""`
Authentication protocol for authenticated SNMPv3 requests.

* `auth_password`:
Authentication password for authenticated SNMPv3 requests.

* `sec_level`: Values: `"noAuthNoPriv"`,`"authNoPriv"`,`"authPriv"`. Default: `"noAuthNoPriv"`
Security level used for SNMPv3 messages.

* `context_name`:
Context name used for SNMPv3 requests.

* `priv_protocol`: Values: `"DES"`,`"AES"`,`""`. Default: `""`
Privacy protocol used for encrypted SNMPv3 messages.

* `priv_password`:
Privacy password used for encrypted SNMPv3 messages.


* `name`:
Output measurement name.

#### Field parameters:
* `oid`:
OID to get. May be a numeric or textual OID.

* `oid_index_suffix`:
The OID sub-identifier to strip off so that the index can be matched against other fields in the table.

* `oid_index_length`:
Specifies the length of the index after the supplied table OID (in OID path segments). Truncates the index after this point to remove non-fixed value or length index suffixes.

* `name`:
Output field/tag name.
If not specified, it defaults to the value of `oid`. If `oid` is numeric, an attempt to translate the numeric OID into a texual OID will be made.

* `is_tag`:
Output this field as a tag.

* `conversion`: Values: `"float(X)"`,`"float"`,`"int"`,`""`. Default: `""`
Converts the value according to the given specification.

    - `float(X)`: Converts the input value into a float and divides by the Xth power of 10. Efficively just moves the decimal left X places. For example a value of `123` with `float(2)` will result in `1.23`.
    - `float`: Converts the value into a float with no adjustment. Same as `float(0)`.
    - `int`: Convertes the value into an integer.
    - `hwaddr`: Converts the value to a MAC address.
    - `ipaddr`: Converts the value to an IP address.

#### Table parameters:
* `oid`:
Automatically populates the table's fields using data from the MIB.

* `name`:
Output measurement name.
If not specified, it defaults to the value of `oid`.  If `oid` is numeric, an attempt to translate the numeric OID into a texual OID will be made.

* `inherit_tags`:
Which tags to inherit from the top-level config and to use in the output of this table's measurement.

* `index_as_tag`:
Adds each row's index within the table as a tag.  

### MIB lookups
If the plugin is configured such that it needs to perform lookups from the MIB, it will use the net-snmp utilities `snmptranslate` and `snmptable`.

When performing the lookups, the plugin will load all available MIBs. If your MIB files are in a custom path, you may add the path using the `MIBDIRS` environment variable. See [`man 1 snmpcmd`](http://net-snmp.sourceforge.net/docs/man/snmpcmd.html#lbAK) for more information on the variable.
