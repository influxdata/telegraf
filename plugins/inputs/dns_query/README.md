# DNS Query Input Plugin

The DNS plugin gathers dns query times in miliseconds - like
[Dig](https://en.wikipedia.org/wiki/Dig_\(command\))

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Query given DNS server and gives statistics
[[inputs.dns_query]]
  ## servers to query
  servers = ["8.8.8.8"]

  ## Network is the network protocol name.
  # network = "udp"

  ## Domains or subdomains to query.
  # domains = ["."]

  ## Query record type.
  ## Possible values: A, AAAA, CNAME, MX, NS, PTR, TXT, SOA, SPF, SRV.
  # record_type = "A"

  ## Dns server port.
  # port = 53

  ## Query timeout
  # timeout = "2s"

  ## Include the specified additional properties in the resulting metric.
  ## The following values are supported:
  ##    "first_ip" -- return IP of the first A and AAAA answer
  ##    "all_ips"  -- return IPs of all A and AAAA answers
  # include_fields = []
```

## Metrics

- dns_query
  - tags:
    - server
    - domain
    - record_type
    - result
    - rcode
  - fields:
    - query_time_ms (float)
    - result_code (int, success = 0, timeout = 1, error = 2)
    - rcode_value (int)

## Rcode Descriptions

|rcode_value|rcode|Description|
|---|-----------|-----------------------------------|
|0  | NoError   | No Error                          |
|1  | FormErr   | Format Error                      |
|2  | ServFail  | Server Failure                    |
|3  | NXDomain  | Non-Existent Domain               |
|4  | NotImp    | Not Implemented                   |
|5  | Refused   | Query Refused                     |
|6  | YXDomain  | Name Exists when it should not    |
|7  | YXRRSet   | RR Set Exists when it should not  |
|8  | NXRRSet   | RR Set that should exist does not |
|9  | NotAuth   | Server Not Authoritative for zone |
|10 | NotZone   | Name not contained in zone        |
|16 | BADSIG    | TSIG Signature Failure            |
|16 | BADVERS   | Bad OPT Version                   |
|17 | BADKEY    | Key not recognized                |
|18 | BADTIME   | Signature out of time window      |
|19 | BADMODE   | Bad TKEY Mode                     |
|20 | BADNAME   | Duplicate key name                |
|21 | BADALG    | Algorithm not supported           |
|22 | BADTRUNC  | Bad Truncation                    |
|23 | BADCOOKIE | Bad/missing Server Cookie         |

## Example Output

```text
dns_query,domain=google.com,rcode=NOERROR,record_type=A,result=success,server=127.0.0.1 rcode_value=0i,result_code=0i,query_time_ms=0.13746 1550020750001000000
```
