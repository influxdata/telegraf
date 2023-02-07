# Tacacs Input Plugin

The Tacacs plugin collects tacacs authentication response times.

### Configuration:

```toml @sample.conf
[[inputs.tacacs]]
  ## An array of Server IPs to gather from, default localhost
  servers = ["127.0.0.1"]

  ## Request source server IP, normally the server running telegraf
  remaddr = "127.0.0.1"

  ## Credentials for tacacs authentication.
  # username = "myuser"
  # password = "mypassword"
  # secret = "mysecret"

  ## Maximum time to receive response.
  # response_timeout = "5s"
```

### Measurements & Fields:

- tacacs
  - responsetime (float)

### Tags:

- All measurements have the following tags:
    - server

### Example Output:

```
tacacs,server=debian-stretch-tacacs responsetime=0.011 1502489900000000000
```
