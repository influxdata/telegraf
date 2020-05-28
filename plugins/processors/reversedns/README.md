# Reverse DNS Processor Plugin

The `reverse_dns` processor does a reverse-dns lookup on fields with IPs in them.

### Configuration:

```toml
[[processors.reverse_dns]]
  # For optimal performance, you may want to limit which metrics are passed to this
  # processor. eg:
  # namepass = ["my_metric_*"]

  # cache_ttl is how long the dns entries should stay cached for.
  # generally longer is better, but if you expect a large number of diverse lookups
  # you'll want to consider memory use.
  cache_ttl = "24h"

  # lookup_timeout is how long should you wait for a single dns request to repsond.
  # this is also the maximum acceptable latency for a metric travelling through
  # the reverse_dns processor. After lookup_timeout is exceeded, a metric will
  # be passed on unaltered.
  # multiple simultaneous resolution requests for the same IP will only make a
  # single rDNS request, and they will all wait for the answer for this long.
  lookup_timeout = "3s"

  max_parallel_lookups = 100

  [[processors.reverse_dns.lookup]]
    # get the ip from the field "source_ip", and put the result in the field "source_name"
    field = "source_ip"
    dest = "source_name"

  [[processors.reverse_dns.lookup]]
    # get the ip from the tag "destination_ip", and put the result in the tag 
    # "destination_name".
    tag = "destination_ip"
    dest = "destination_name"

    # If you would prefer destination_name to be a field you can use a subsequent 
    # converter like so:
    #   [[processors.converter.tags]]
    #     string = ["destination_name"]
    #     order = 2 # orders are necessary with multiple processors when order matters
```



### Example processing:

example config:

```toml
[[processors.reverse_dns]]
  [[processors.reverse_dns.lookup]]
    tag = "ip"
    dest = "domain"
```

```diff
- ping,ip=8.8.8.8 elapsed=300i 1502489900000000000
+ ping,ip=8.8.8.8,domain=dns.google. elapsed=300i 1502489900000000000
```
