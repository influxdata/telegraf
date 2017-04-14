# Telegraf Input Plugin: Icinga2

This plugin gather services & hosts status using Icinga2 Remote API.

The icinga2 plugin uses the icinga2 remote API to gather status on running
services and hosts. You can read Icinga2's documentation for their remote API
[here](https://docs.icinga.com/icinga2/latest/doc/module/icinga2/chapter/icinga2-api)

### Configuration:

```toml
# Description
[[inputs.icinga2]]
    ## Icing2 Endpoint
    server = "https://127.0.0.1:5665"
    ## Required Icinga2 object type ("services" or "hosts, default "services")
    filter = "services"
    ## Required username used for request HTTP Basic Authentication (default: "")
    username = "root"
    ## Required password used for HTTP Basic Authentication (default: "")
    password = "icinga"
```

### Measurements & Fields:

- ll measurements have the following fields:
    - name (string)
    - status (int)

### Tags:

- All measurements have the following tags:
    - check_command
    - display_name

### Sample Queries:

```
SELECT * FROM "icinga2_services_status" WHERE status = 0 AND time > now() - 24h // Service with OK status
SELECT * FROM "icinga2_services_status" WHERE status = 1 AND time > now() - 24h // Service with WARNING status
SELECT * FROM "icinga2_services_status" WHERE status = 2 AND time > now() - 24h // Service with Critical status
SELECT * FROM "icinga2_services_status" WHERE status = 3 AND time > now() - 24h // Service with UNKNOWN status
```

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter icinga2 -test
icinga2_hosts_status,display_name=router-fr.eqx.fr,check_command=hostalive-custom,host=test-vm name="router-fr.eqx.fr",status=0 1492021603000000000
```
