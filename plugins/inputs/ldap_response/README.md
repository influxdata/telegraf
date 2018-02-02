# Ldap_Response Input Plugin

This plugin is generic and should work with most LDAP servers. In the same way the http_response plugin captures basic timing metrics of an http endpoint, this plugin captures basic timing metrics of an LDAP endpoint: connection, bind, search, and the total. 

Much of the connectivity and field gathering is borrowed from the openldap plugin.

### Configuration:

```toml
[[inputs.ldap_response]]
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

  # base entry for searches
  search_base = ""

  # ldap search to perform. defaults to "(objectClass=*)" if unspecified.
  search_filter = ""

  # the attributes to return as fields. defaults to "objectclass" if unspecified.
  search_attributes = [
    "attribute1",
    "attribute2",
  ]
```

### Measurements & Fields:

This plugin produces a single measurement: **ldap_response** with the following fields:

- **connect_time_ms** - time in milliseconds to connect to the ldap server
- **bind_time_ms** - time in milliseconds to perform a bind (i.e. authentication) with the ldap server
- **query_time_ms** - time in milliseconds to perform the requested search with the ldap server
- **total_time_ms** - time in milliseconds for the entire interaction with the ldap server
- *any search_attributes* - attributes that are specified as *search_attributes* will be looked up and the entry's CN will be suffixed with `_` and the name of the attribute **IF** the field value is an integer.

### Tags:

- server= # value from config
- port= # value from config

### Example Output:

```
$ telegraf -config telegraf.conf -input-filter ldap_response -test --debug
* Plugin: inputs.ldap_response, Collection 1
> ldap_response,server=localhost,port=389,environment=prod query_time_ms=33.8776,total_time_ms=107.7321,connect_time_ms=54.0426,bind_time_ms=19.7915 1518821618000000000
```
