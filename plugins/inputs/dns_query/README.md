# DNS Query Input Plugin

The DNS plugin gathers dns query times in miliseconds - like [Dig](https://en.wikipedia.org/wiki/Dig_\(command\))

### Configuration:
```toml
# Query given DNS server and gives statistics
[[inputs.dns_query]]
  ## servers to query
  servers = ["8.8.8.8"]

  ## Network is the network protocol name.
  # network = "udp"

  ## Domains or subdomains to query.
  # domains = ["."]

  ## Query record type.
  ## Posible values: A, AAAA, CNAME, MX, NS, PTR, TXT, SOA, SPF, SRV.
  # record_type = "A"

  ## Dns server port.
  # port = 53

  ## Query timeout in seconds.
  # timeout = 2
```

### Metrics:

- dns_query
  - tags:
    - server
    - domain
    - record_type
    - result
  - fields:
    - query_time_ms (float)
    - result_code (int, success = 0, timeout = 1, error = 2)

### Example Output:

```
dns_query,domain=mjasion.pl,record_type=A,server=8.8.8.8 query_time_ms=67.189842 1456082743585760680
```
