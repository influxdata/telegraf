# Tacacs Input Plugin

The Tacacs plugin collects tacacs authentication response times.

### Configuration:

```toml @sample.conf
[[inputs.tacacs]]
  ## An array of Server IPs and ports to gather from. If none specified, defaults to localhost.
  servers = ["127.0.0.1:49","hostname.domain.com:49"]

  ## Request source server IP, normally the server running telegraf.
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
    - source

### Example Output:

```
tacacs,source=debian-stretch-tacacs responsetime=0.011 1502489900000000000
```
