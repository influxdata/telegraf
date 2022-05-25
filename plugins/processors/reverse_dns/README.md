# Reverse DNS Processor Plugin

The `reverse_dns` processor does a reverse-dns lookup on tags (or fields) with
IPs in them.

Telegraf minimum version: Telegraf 1.15.0

## Configuration

```toml
# ReverseDNS does a reverse lookup on IP addresses to retrieve the DNS name
[[processors.reverse_dns]]
  ## For optimal performance, you may want to limit which metrics are passed to this
  ## processor. eg:
  ## namepass = ["my_metric_*"]

  ## cache_ttl is how long the dns entries should stay cached for.
  ## generally longer is better, but if you expect a large number of diverse lookups
  ## you'll want to consider memory use.
  cache_ttl = "24h"

  ## lookup_timeout is how long should you wait for a single dns request to repsond.
  ## this is also the maximum acceptable latency for a metric travelling through
  ## the reverse_dns processor. After lookup_timeout is exceeded, a metric will
  ## be passed on unaltered.
  ## multiple simultaneous resolution requests for the same IP will only make a
  ## single rDNS request, and they will all wait for the answer for this long.
  lookup_timeout = "3s"

  ## max_parallel_lookups is the maximum number of dns requests to be in flight
  ## at the same time. Requesting hitting cached values do not count against this
  ## total, and neither do mulptiple requests for the same IP.
  ## It's probably best to keep this number fairly low.
  max_parallel_lookups = 10

  ## ordered controls whether or not the metrics need to stay in the same order
  ## this plugin received them in. If false, this plugin will change the order
  ## with requests hitting cached results moving through immediately and not
  ## waiting on slower lookups. This may cause issues for you if you are
  ## depending on the order of metrics staying the same. If so, set this to true.
  ## keeping the metrics ordered may be slightly slower.
  ordered = false

  [[processors.reverse_dns.lookup]]
    ## get the ip from the field "source_ip", and put the result in the field "source_name"
    field = "source_ip"
    dest = "source_name"

  [[processors.reverse_dns.lookup]]
    ## get the ip from the tag "destination_ip", and put the result in the tag
    ## "destination_name".
    tag = "destination_ip"
    dest = "destination_name"

    ## If you would prefer destination_name to be a field instead, you can use a
    ## processors.converter after this one, specifying the order attribute.
```

## Example

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
