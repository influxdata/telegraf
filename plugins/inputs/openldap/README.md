# Openldap Input Plugin

This plugin gathers metrics from OpenLDAP's cn=Monitor backend.

### Configuration:

```toml
[[inputs.openldap]]
  host = "localhost"
  port = 389

  # ldaps, starttls, or no encryption. default is an empty string, disabling all encryption.
  # note that port will likely need to be changed to 636 for ldaps
  # valid options: "" | "starttls" | "ldaps"
  ssl = ""

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
> openldap,server=localhost,port=389,host=zirzla search_operations_completed=2i,delete_operations_completed=0i,read_waiters=1i,total_connections=1004i,bind_operations_completed=3i,unbind_operations_completed=3i,referrals_statistics=0i,current_connections=1i,bind_operations_initiated=3i,compare_operations_completed=0i,add_operations_completed=2i,delete_operations_initiated=0i,unbind_operations_initiated=3i,search_operations_initiated=3i,add_operations_initiated=2i,max_file_descriptors_connections=4096i,abandon_operations_initiated=0i,write_waiters=0i,modrdn_operations_completed=0i,abandon_operations_completed=0i,pdu_statistics=23i,modify_operations_initiated=0i,bytes_statistics=1660i,entries_statistics=17i,compare_operations_initiated=0i,modrdn_operations_initiated=0i,extended_operations_completed=0i,modify_operations_completed=0i,extended_operations_initiated=0i 1499990455000000000
```
