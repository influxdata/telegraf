# DNS Query Input Plugin

The DNS plugin gathers dns query times in miliseconds - like [Dig](https://en.wikipedia.org/wiki/Dig_\(command\))

### Configuration:

```
# Sample Config:
[[inputs.dns_query]]
  ### servers to query
  servers = ["8.8.8.8"] # required

  ### Domains or subdomains to query. "."(root) is default
  domains = ["mjasion.pl"] #optional

  ### Query record type. Posible values: A, AAAA, ANY, CNAME, MX,  NS, PTR, SOA, SPF, SRV, TXT. Default is "NS"
  recordType = "A" # optional

  ### Dns server port. 53 is default
  port = 53 # optional

  ### Query timeout in seconds. Default is 2 seconds
  timeout = 2 # optional
```

For querying more than one record type make:
 
```
[[inputs.dns_query]]
  domains = ["mjasion.pl"]
  servers = ["8.8.8.8", "8.8.4.4"]
  recordType = "A"

[[inputs.dns_query]]
  domains = ["mjasion.pl"]
  servers = ["8.8.8.8", "8.8.4.4"]
  recordType = "MX"
```

### Tags:

- server 
- domain
- recordType

### Example output:

```
./telegraf -config telegraf.conf -test -input-filter dns_query -test
> dns_query,domain=mjasion.pl,record_type=A,server=8.8.8.8 query_time_ms=67.189842 1456082743585760680
```