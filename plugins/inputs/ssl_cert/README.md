# SSL Cert Input Plugin

This plugin provides information about SSL certificate accessible via local
file or network connection.


### Configuration

```toml
# Reads metrics from a SSL certificate
[[inputs.ssl_cert]]
  ## List of local SSL files
  #files = []
  ## List of servers
  #servers = []
  ## Timeout for SSL connection
  #timeout = 5
```


### Metrics

- `ssl_cert`
  - tags:
    - `server` (only if `servers` parameter is defined)
    - `file` (only if `files` parameter is defined)
  - fields:
    - `expiry` (int, seconds)
    - `age` (int, seconds)
    - `startdate` (int, seconds)
    - `enddate` (int, seconds)


### Example output

```
ssl_cert,server=google.com:443,host=myhost age=1753627i,expiry=5503972i,startdate=1516092060i,enddate=1523349660i 1517845687000000000
ssl_cert,host=myhost,file=/path/to/the.crt age=7522207i,expiry=308002732i,startdate=1510323480i,enddate=1825848420i 1517845687000000000
```
