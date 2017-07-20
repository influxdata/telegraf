# Telegraf Plugin: Nginx

### Configuration:

```
# Read Nginx's basic status information (ngx_http_stub_status_module)
[[inputs.nginx]]
  ## An array of Nginx stub_status URI to gather stats.
  urls = ["http://localhost/server_status"]

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP response timeout (default: 5s)
  response_timeout = "5s"
```

### Measurements & Fields:

- Measurement
    - accepts
    - active
    - handled
    - reading
    - requests
    - waiting
    - writing

### Tags:

- All measurements have the following tags:
    - port
    - server

### Example Output:

Using this configuration:
```
[[inputs.nginx]]
  ## An array of Nginx stub_status URI to gather stats.
  urls = ["http://localhost/status"]
```

When run with:
```
./telegraf --config telegraf.conf --input-filter nginx --test
```

It produces:
```
* Plugin: nginx, Collection 1
> nginx,port=80,server=localhost accepts=605i,active=2i,handled=605i,reading=0i,requests=12132i,waiting=1i,writing=1i 1456690994701784331
```

### Reference material

Structures for Nginx Plus have been built based on history of
[status module documentation](http://nginx.org/en/docs/http/ngx_http_status_module.html)

Subsequent versions of status response structure available here:

- [version 1](http://web.archive.org/web/20130805111222/http://nginx.org/en/docs/http/ngx_http_status_module.html)

- [version 2](http://web.archive.org/web/20131218101504/http://nginx.org/en/docs/http/ngx_http_status_module.html)

- version 3 - not available

- [version 4](http://web.archive.org/web/20141218170938/http://nginx.org/en/docs/http/ngx_http_status_module.html)

- [version 5](http://web.archive.org/web/20150414043916/http://nginx.org/en/docs/http/ngx_http_status_module.html)

- [version 6](http://web.archive.org/web/20150918163811/http://nginx.org/en/docs/http/ngx_http_status_module.html)

- [version 7](http://web.archive.org/web/20161107221028/http://nginx.org/en/docs/http/ngx_http_status_module.html)
