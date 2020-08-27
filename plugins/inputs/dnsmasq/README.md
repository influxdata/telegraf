# Dnsmasq Input Plugin

The Dnsmasq plugin gathers Dnsmasq statistics about DNS side.

See "cache statistics" section in [https://manpages.debian.org/stretch/dnsmasq-base/dnsmasq.8.en.html#NOTES](https://manpages.debian.org/stretch/dnsmasq-base/dnsmasq.8.en.html#NOTES)

An example command to query this, using the dig utility would be

``` shell
dig +short chaos txt cachesize.bind
```


### Configuration:
```toml
# Read metrics about dnsmasq dns side.
[[inputs.dnsmasq]]
  # Dnsmasq server IP address.
  server = "127.0.0.1"
  
  # Dnsmasq server port.
  port = 53
```

### Metrics:

- dnsmasq
  - tags:
    - server
  - fields:
    - auth (float)
    - cachesize (float)
    - evictions (float)
    - hits (float)
    - insertions (float)
    - misses (float)
	- queries (float)
	- queries_failed (float)

### Example Output:

```
dnsmasq,host=localhost,server=127.0.0.1,port=53 insertions=0,evictions=0,misses=0,hits=12,auth=0,queries=0,queries_failed=0,cachesize=150 1598519060000000000
```
