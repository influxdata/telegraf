# LDAP Input Plugin

This plugin gathers metrics from LDAP servers' monitoring (`cn=Monitor`)
backend. Currently this plugin supports [OpenLDAP][openldap] and [389ds][389ds]
servers.

⭐ Telegraf v1.29.0
🏷️ network, server
💻 all

[openldap]: https://www.openldap.org/
[389ds]: https://www.port389.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# LDAP monitoring plugin
[[inputs.ldap]]
  ## Server to monitor
  ## The scheme determines the mode to use for connection with
  ##    ldap://...      -- unencrypted (non-TLS) connection
  ##    ldaps://...     -- TLS connection
  ##    starttls://...  --  StartTLS connection
  ## If no port is given, the default ports, 389 for ldap and starttls and
  ## 636 for ldaps, are used.
  server = "ldap://localhost"

  ## Server dialect, can be "openldap" or "389ds"
  # dialect = "openldap"

  # DN and password to bind with
  ## If bind_dn is empty an anonymous bind is performed.
  bind_dn = ""
  bind_password = ""

  ## Reverse the field names constructed from the monitoring DN
  # reverse_field_names = false

  ## Optional TLS Config
  ## Set to true/false to enforce TLS being enabled/disabled. If not set,
  ## enable TLS only if any of the other options are specified.
  # tls_enable =
  ## Trusted root certificates for server
  # tls_ca = "/path/to/cafile"
  ## Used for TLS client certificate authentication
  # tls_cert = "/path/to/certfile"
  ## Used for TLS client certificate authentication
  # tls_key = "/path/to/keyfile"
  ## Password for the key file if it is encrypted
  # tls_key_pwd = ""
  ## Send the specified TLS server name via SNI
  # tls_server_name = "kubernetes.example.com"
  ## Minimal TLS version to accept by the client
  # tls_min_version = "TLS12"
  ## List of ciphers to accept, by default all secure ciphers will be accepted
  ## See https://pkg.go.dev/crypto/tls#pkg-constants for supported values.
  ## Use "all", "secure" and "insecure" to add all support ciphers, secure
  ## suites or insecure suites respectively.
  # tls_cipher_suites = ["secure"]
  ## Renegotiation method, "never", "once" or "freely"
  # tls_renegotiation_method = "never"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

To use this plugin you must enable the monitoring backend/plugin of your LDAP
server. See [OpenLDAP][openldap_monitoring] or [389ds][389ds] documentation for
details.

[openldap_monitoring]: https://www.openldap.org/devel/admin/monitoringslapd.html

## Metrics

Depending on the server dialect, different metrics are produced. The metrics
are usually named according to the selected dialect.

### Tags

- server -- Server name or IP
- port   -- Port used for connecting

## Example Output

Using the `openldap` dialect

```text
openldap,server=localhost,port=389 operations_completed=63i,operations_initiated=98i,operations_bind_initiated=10i,operations_unbind_initiated=6i,operations_modrdn_completed=0i,operations_delete_initiated=0i,operations_add_completed=2i,operations_delete_completed=0i,operations_abandon_completed=0i,statistics_entries=1516i,threads_open=2i,threads_active=1i,waiters_read=1i,operations_modify_completed=0i,operations_extended_initiated=4i,threads_pending=0i,operations_search_initiated=36i,operations_compare_initiated=0i,connections_max_file_descriptors=4096i,operations_modify_initiated=0i,operations_modrdn_initiated=0i,threads_max=16i,time_uptime=6017i,connections_total=1037i,connections_current=1i,operations_add_initiated=2i,statistics_bytes=162071i,operations_unbind_completed=6i,operations_abandon_initiated=0i,statistics_pdu=1566i,threads_max_pending=0i,threads_backload=1i,waiters_write=0i,operations_bind_completed=10i,operations_search_completed=35i,operations_compare_completed=0i,operations_extended_completed=4i,statistics_referrals=0i,threads_starting=0i 1516912070000000000
```

Using the `389ds` dialect

```text
389ds,port=32805,server=localhost add_operations=0i,anonymous_binds=0i,backends=0i,bind_security_errors=0i,bytes_received=0i,bytes_sent=256i,cache_entries=0i,cache_hits=0i,chainings=0i,compare_operations=0i,connections=1i,connections_in_max_threads=0i,connections_max_threads=0i,copy_entries=0i,current_connections=1i,current_connections_at_max_threads=0i,delete_operations=0i,dtablesize=63936i,entries_returned=2i,entries_sent=2i,errors=2i,in_operations=11i,list_operations=0i,maxthreads_per_conn_hits=0i,modify_operations=1i,modrdn_operations=0i,onelevel_search_operations=0i,operations_completed=10i,operations_initiated=11i,read_operations=0i,read_waiters=0i,referrals=0i,referrals_returned=0i,search_operations=3i,security_errors=0i,simpleauth_binds=1i,strongauth_binds=2i,threads=17i,total_connections=4i,unauth_binds=0i,wholesubtree_search_operations=1i 1695637234047087280
```
