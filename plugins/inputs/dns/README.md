# DNS Input Plugin

The DNS plugin gathers dns query times in miliseconds - like [Dig](https://en.wikipedia.org/wiki/Dig_\(command\))

### Configuration:

```
# Sample Config:
[[inputs.dns]]
  ### Domains or subdomains to query
  domains = ["mjasion.pl"] # required

  ### servers to query
  servers = ["8.8.8.8"] # required

  ### Query record type. Posible values: A, CNAME, MX, TXT, NS. Default is "A"
  recordType = "A" # optional

  ### Dns server port. 53 is default
  port = 53 # optional

  ### Query timeout in seconds. Default is 2 seconds
  timeout = 2 # optional
```

For querying more than one record type make:
 
```
[[inputs.dns]]
  domains = ["mjasion.pl"]
  servers = ["8.8.8.8", "8.8.4.4"]
  recordType = "A"

[[inputs.dns]]
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
./telegraf -config telegraf.conf -test -input-filter dns -test
> dns,domain=mjasion.pl,recordType=A,server=8.8.8.8 value=25.236181 1455452083165126877
```
