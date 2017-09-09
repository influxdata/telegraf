# Telegraf Plugin: nginx_plus

### Configuration:

```
# Read Nginx Plus' advanced status information
[[inputs.nginx_plus]]
  ## An array of Nginx status URIs to gather stats.
  urls = ["http://localhost/status"]
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
[[inputs.nginx_plus]]
  ## An array of Nginx Plus status URIs to gather stats.
  urls = ["http://localhost/status"]
```

When run with:
```
./telegraf -config telegraf.conf -input-filter nginx_plus -test
```

It produces:
```
* Plugin: inputs.nginx_plus, Collection 1
> nginx_processes,server=localhost,port=12021,host=word.local respawned="0xc420075538" 1504922954000000000
> nginx_connections,server=localhost,port=12021,host=word.local accepted=4458727685i,dropped=10138424i,active=10256i,idle=29390i 1504922954000000000
> nginx_ssl,server=localhost,port=12021,host=word.local handshakes=0i,handshakes_failed=0i,session_reuses=0i 1504922954000000000
> nginx_requests,host=word.local,server=localhost,port=12021 total=147885504244i,current=10019i 1504922954000000000
> nginx_upstream,host=word.local,upstream=dataserver80,server=localhost,port=12021 zombies=0i,keepalive=0i 1504922954000000000
> nginx_upstream_peer,id=0,server=localhost,port=12021,host=word.local,upstream=dataserver80,serverAddress=10.10.102.181:80 responses_5xx=27831i,healthchecks_unhealthy=1i,downtime=484817i,healthchecks_last_passed=true,responses_1xx=0i,active=22i,requests=2620930i,responses_total=2620652i,fails=4i,downstart=0i,state="up",responses_4xx=16i,healthchecks_checks=14133i,selected="0xc4201b22e8",response_time=95i,responses_2xx=2592805i,weight=1i,responses_3xx=0i,sent=3802831967i,received=536695496i,unavail=4i,healthchecks_fails=27i,header_time=94i,backup=false 1504922954000000000
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
