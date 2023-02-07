# Radius Input Plugin

The Radius plugin collects radius response times.

### Configuration:

```toml @sample.conf
[[inputs.radius]]
  ## An array of Server IPs and ports to gather from.
  servers = ["127.0.0.1:1812"]

  ## Credentials for radius authentication.
  # username = "myuser"
  # password = "mypassword"
  # secret = "mysecret"

  ## Maximum time to receive response.
  # response_timeout = "5s"
```

### Measurements & Fields:

- radius
  - responsetime (float)

### Tags:

- All measurements have the following tags:
    - port
    - server

### Example Output:

```
radius,port=1812,server=debian-stretch-radius responsetime=0.011 1502489900000000000
```
