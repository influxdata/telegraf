# X509 Cert Input Plugin

This plugin provides information about X509 certificate accessible via local
file or network connection.


### Configuration

```toml
# Reads metrics from a SSL certificate
[[inputs.x509_cert]]
  ## List certificate sources
  sources = ["/etc/ssl/certs/ssl-cert-snakeoil.pem", "https://example.org:443"]

  ## Timeout for SSL connection
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```


### Metrics

- x509_cert
  - tags:
    - source - source of the certificate
    - organization
    - organizational_unit
    - country
    - province
    - locality
  - fields:
    - expiry (int, seconds)
    - age (int, seconds)
    - startdate (int, seconds)
    - enddate (int, seconds)


### Example output

```
x509_cert,host=myhost,source=https://example.org age=1753627i,expiry=5503972i,startdate=1516092060i,enddate=1523349660i 1517845687000000000
x509_cert,host=myhost,source=/etc/ssl/certs/ssl-cert-snakeoil.pem age=7522207i,expiry=308002732i,startdate=1510323480i,enddate=1825848420i 1517845687000000000
```
