# Example Input Plugin

This plugin gather services & hosts status using Icinga2 Remote API.

### Configuration:

```toml
# Description
[[inputs.icinga2]]
    server = "https://127.0.0.1:5665"
    filter = "services"
    username = "root"
    password = "icinga"
```

### Measurements & Fields:

- ll measurements have the following fields:
    - name (string)
    - status (int)

### Tags:

- All measurements have the following tags:
    - check_command (optional description)
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



[
  {
    "attrs": {
      "check_command": "check-bgp-juniper-netconf",
      "display_name": "tsh-core1.intercloud.fr - VRF-CSP-GOOGLE--TSH.inet6.0",
      "last_check": 1491827091.4058940411,
      "name": "ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
      "state": 0
    },
    "joins": {},
    "meta": {},
    "name": "tsh-core1.intercloud.fr!ef017af8-c684-4f3f-bb20-0dfe9fcd3dbe",
    "type": "Service"
  },
  {
    "attrs": {
      "check_command": "hostalive-custom",
      "display_name": "cpe-mgmt1.intercloud-test-dev.fr-vty.intercloud.fr",
      "last_check": 1491827181.4462339878,
      "name": "1bc2c4b7-6523-4d4d-a8ce-45a357ccd700",
      "state": 0
    },
    "joins": {},
    "meta": {},
    "name": "1bc2c4b7-6523-4d4d-a8ce-45a357ccd700",
    "type": "Host"
  }
]
