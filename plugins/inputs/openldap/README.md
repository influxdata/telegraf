# Openldap Input Plugin

This plugin gathers metrics from OpenLDAP's cn=Monitor backend.

### Configuration:

```toml
# Description
[[inputs.openldap]]
  host = "localhost"
  port = 389

  # ldaps, starttls. default is an empty string, disabling all encryption.
  # note that port will likely need to be changed to 636 for ldaps
  ssl = "" | "starttls" | "ldaps"
  
  # skip peer certificate verification. Default is false.
  insecure_skip_verify = false
 
  # Path to PEM-encoded Root certificate to use to verify server certificate
  ssl_ca = "/etc/ssl/certs.pem"

  # dn/password to bind with. If bind_dn is empty, an anonymous bind is performed.
  bind_dn = ""
  bind_password = ""
```

### Measurements & Fields:

All **monitorCounter**, **monitorOpInitiated**, and **monitorOpCompleted** attributes are gathered based on this LDAP query:

```(|(objectClass=monitorCounterObject)(objectClass=monitorOperation))```

Metric names are based on their entry DN.

Metrics for the **monitorOp*** attributes have **_initiated** and **_completed** added to the base name.

An OpenLDAP 2.4 server will provide these metrics:

- openldap
	- max_file_descriptors_connections
	- current_connections
	- total_connections
	- abandon_operations_completed
	- abandon_operations_initiated
	- add_operations_completed
	- add_operations_initiated
	- bind_operations_completed
	- bind_operations_initiated
	- compare_operations_completed
	- compare_operations_initiated
	- delete_operations_completed
	- delete_operations_initiated
	- extended_operations_completed
	- extended_operations_initiated
	- modify_operations_completed
	- modify_operations_initiated
	- modrdn_operations_completed
	- modrdn_operations_initiated
	- search_operations_completed
	- search_operations_initiated
	- unbind_operations_completed
	- unbind_operations_initiated
	- bytes_statistics
	- entries_statistics
	- pdu_statistics
	- referrals_statistics
	- read_waiters
	- write_waiters

### Tags:

- server= # value from config
- port= # value from config
    
### Example Output:

```
$ telegraf -config telegraf.conf -input-filter openldap -test --debug
* Plugin: inputs.openldap, Collection 1
> openldap,port=389,host=localhost,server=localhost abandon_operations_initiated=4,extended_operations_completed=125963,bytes_statistics=595939321,pdu_statistics=17028251,modify_operations_initiated=0,delete_operations_completed=0,compare_operations_completed=0,max_file_descriptors_connections=4096,unbind_operations_completed=7981688,extended_operations_initiated=125963,referrals_statistics=0,modify_operations_completed=0,delete_operations_initiated=0,bind_operations_completed=8115329,search_operations_completed=4385841,add_operations_completed=0,abandon_operations_completed=4,write_waiters=0,bind_operations_initiated=8115329,modrdn_operations_initiated=0,compare_operations_initiated=0,entries_statistics=4401128,read_waiters=1,current_connections=3,search_operations_initiated=4385842,modrdn_operations_completed=0,add_operations_initiated=0,total_connections=8147531,unbind_operations_initiated=7981688 1491189665000000000
```
