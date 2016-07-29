# Check SSL Input Plugin

This input plugin will return how much time (in seconds) left for a SSL cert to expire.
Warning, this check doesnt verify if SSL is valid/secure or not.

### Configuration:

```
# SSL request given a list of servers (server:port) and a timeout
[[inputs.check_ssl]]
  ## Servers ( Default [] )
  servers = ["github.com:443"]
  ## Set response_timeout (default 5 seconds)
  response_timeout = 5s
```

### Measurements & Fields:

- ssl_cert
    - time_to_expire (int) # seconds left for the SSL cert to expire

### Tags:

- All measurements have the following tags:
    - server

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter check_ssl -test
> ssl_cert,server=www.google.com:443 time_to_expire=6185474.476944118 1468864305580596685
```
