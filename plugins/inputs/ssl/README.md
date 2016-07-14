# Example Input Plugin

This input plugin will return how many days left for a SSL cert to expire.

### Configuration:

```
# SSL request given a server, a Port, a timeout and a skipverify flag
[[inputs.check_ssl]]
  ## Server (default github.com)
  server = "github.com"
  ## Set response_timeout (default 5 seconds)
  response_timeout = 5
  ## Port (Default 443)
  port = "443"
  ## SSL Skip Verification of certificates
  skip_verify = false
```

### Measurements & Fields:

- expire_time
    - days_to_expire (int) # Days left for the SSL cert to expire

### Tags:

- All measurements have the following tags:
    - server

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter check_ssl -test
> expire_time,server=github.com days_to_expire=248i 1468251532250867718
```
