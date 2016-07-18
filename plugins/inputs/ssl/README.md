# Example Input Plugin

This input plugin will return how much time (in seconds) left for a SSL cert to expire.
A string will be returned as field to show the error message for those servers in the 
list that have failed for some reason. If site is ok, "error" field will be empty

### Configuration:

```
# SSL request given a server, a Port, a timeout and a skipverify flag
[[inputs.check_ssl]]
  ## Servers ( Default [] )
  servers = ["github.com:443"]
  ## Set response_timeout (default 5 seconds)
  response_timeout = 5s
```

### Measurements & Fields:

- ssl_cert
    - time_to_expire (int) # seconds left for the SSL cert to expire
    - error (string) # error message if something fail (Nil if OK)

### Tags:

- All measurements have the following tags:
    - server

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter check_ssl -test
> ssl_cert,server=www.google.com:443 error=,time_to_expire=6185474.476944118 1468864305580596685
```
