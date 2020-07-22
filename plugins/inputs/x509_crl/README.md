# X509 CRL Input Plugin

This plugin provides information about X509 CRL accessible via file.


### Configuration

```toml
# Reads metrics from a X509 CRL
[[inputs.x509_crl]]
  ## List certificate sources
  sources = ["/etc/nginx/ssl/mycrl.pem"]
```


### Metrics

- x509_crl
  - tags:
    - source - source of the CRL
    - issuer - the CA's DN that generated the CRL
    - version
  - fields:
    - startdate (int, seconds) - when the CRL was generated 
    - enddate (int, seconds) - when the CRL has to be renewed
    - expiry (int, seconds) - time to expiration - can be negative when expired
    - has_expired (boolean) - is it still valid
    - revoked_certificates (int) - number of revoked certificated 


### Example output

```
x509_crl,issuer:CN=ac,O=Alsace\ Reseau\ Neutre,L=Strasbourg,ST=Alsace,C=FR,source:/tmp/x509crl_tmp_file241751718,version:0 end_date:1583509523i,has_expired:false,revoked_certificates:0i,start_date:1580917523i 1563582256000000000
```
