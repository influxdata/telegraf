# Telegraf Plugin: SSL

### Configuration:

```
# Check expiration date and domains of ssl certificate
[[inputs.ssl]]
  ## Server to check
  [[inputs.ssl.servers]]
    host = "google.com:443"
    timeout = 5
  ## Server to check
  [[inputs.ssl.servers]]
    host = "github.com"
    timeout = 5
```

### Tags:

- domain
- port

### Fields:

- time_to_expiration(int)

### Example Output:

If ssl certificate is valid:

```
* Plugin: inputs.ssl, Collection 1
> ssl,domain=example.com,port=443,host=host time_to_expiration=3907728i 1517213967000000000
```

If ssl certificate and domain mismatch:

```
* Plugin: inputs.ssl, Collection 1
2018-01-29T08:20:33Z E! Error in plugin [inputs.ssl]: [example.com:443] cert and domain mismatch
> ssl,domain=example.com,port=443,host=host time_to_expiration=0i 1517214033000000000
```

If ssl certificate has expired:

```
* Plugin: inputs.ssl, Collection 1
2018-01-29T08:20:33Z E! Error in plugin [inputs.ssl]: [example.com:443] cert has expired
> ssl,domain=example.com,port=443,host=host time_to_expiration=0i 1517214033000000000
```
