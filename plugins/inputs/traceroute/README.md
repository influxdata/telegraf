# traceroute Input Plugin

The traceroute plugin provides routing information given end host.

### Configuration:

```toml
# NOTE: this plugin forks the traceroute command. You may need to set capabilities
# via setcap cap_net_raw+p /bin/traceroute
[[inputs.traceroute]]
  ## List of urls to traceroute
  urls = ["www.google.com"] # required
  ## per-traceroute timeout, in s. 0 == no timeout
  ## it is highly recommended to set this value to match the telegraf interval
  # response_timeout = 0.0
  ## wait time per probe in seconds (traceroute -w <WAITTIME>)
  # waittime = 5.0
  ## starting TTL of packet (traceroute -f <FIRST_TTL>)
  # first_ttl = 1
  ## maximum number of hops (hence TTL) traceroute will probe (traceroute -m <MAX_TTL>)
  # max_ttl = 30
  ## number of probe packets sent per hop (traceroute -q <NQUERIES>)
  # nqueries = 3
  ## do not try to map IP addresses to host names (traceroute -n)
  # no_host_name = false
  ## use ICMP packets (traceroute -I)
  # icmp = false
  ## Lookup AS path in routes (traceroute -A)
  # as_path_lookups = false
  ## source interface/address to traceroute from (traceroute -i <INTERFACE/SRC_ADDR>)
  # interface = ""
```

### Metrics:

- traceroute
  - tags:
    - target_fqdn 
    - target_ip (IPv4 string)
  - fields:
    - result_code
        - 0:success
      	- 1:no such host
    - number_of_hops (int, # of hops made)

- traceroute_hop_data
  - tags:
    - target_fqdn
    - target_ip (IPv4 string)
    - column_number (zero-indexed value representing which column of the traceroute output the data resides in)
    - hop_fqdn
    - hop_ip (IPv4 string)
    - hop_number (string)
  - fields:
    - hop_rtt_ms (round trip time in ms)
    - hop_asn (ASN number ex. "AS1234" or multiple ASN's ex. "AS1234/AS5678" of hop ip)

### Sample Queries:

Get traceroute information given host
```
SELECT *
FROM "traceroute"
WHERE "target_fqdn"='www.google.com'
```

Get average round trip team for each top given time
```
SELECT MEAN("hop_rtt_ms")
FROM "traceroute_hop_data"
WHERE "time"=1453831884664956455
GROUP BY "hop_number"
```

### Example Output:

#### traceroute
```
> traceroute,host=m1.cloudpbx.ca,target_fqdn=www.google.com,target_ip=172.217.0.100 number_of_hops=6i 1525474707000000000
```

#### traceroute_hop_data
```
> traceroute,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 number_of_hops=9i,result_code=0i 1529342902000000000
> traceroute_hop_data,column_number=0,hop_fqdn=167.99.176.254,hop_ip=167.99.176.254,hop_number=1,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 hop_asn="",hop_rtt_ms=4.804999828338623 1529342902000000000
> traceroute_hop_data,column_number=0,hop_fqdn=138.197.249.90,hop_ip=138.197.249.90,hop_number=2,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 hop_asn="",hop_rtt_ms=0.9390000104904175 1529342902000000000
> traceroute_hop_data,column_number=0,hop_fqdn=162.243.190.33,hop_ip=162.243.190.33,hop_number=3,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 hop_asn="",hop_rtt_ms=1.1859999895095825 1529342902000000000
> traceroute_hop_data,column_number=0,hop_fqdn=108.170.250.227,hop_ip=108.170.250.227,hop_number=4,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 hop_asn="",hop_rtt_ms=1.125 1529342902000000000
> traceroute_hop_data,column_number=0,hop_fqdn=74.125.252.132,hop_ip=74.125.252.132,hop_number=5,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 hop_asn="",hop_rtt_ms=15.60099983215332 1529342902000000000
> traceroute_hop_data,column_number=0,hop_fqdn=209.85.249.137,hop_ip=209.85.249.137,hop_number=6,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 hop_asn="",hop_rtt_ms=16.5 1529342902000000000
> traceroute_hop_data,column_number=0,hop_fqdn=108.170.243.174,hop_ip=108.170.243.174,hop_number=7,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 hop_asn="",hop_rtt_ms=17.04599952697754 1529342902000000000
> traceroute_hop_data,column_number=0,hop_fqdn=72.14.239.115,hop_ip=72.14.239.115,hop_number=8,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 hop_asn="",hop_rtt_ms=14.781000137329102 1529342902000000000
> traceroute_hop_data,column_number=0,hop_fqdn=ord38s09-in-f4.1e100.net,hop_ip=172.217.9.68,hop_number=9,host=m1_cloudpbx,target_fqdn=www.google.com,target_ip=172.217.9.68 hop_asn="",hop_rtt_ms=18.89900016784668 1529342902000000000
```


Sponsored by [CloudPBX](http://CloudPBX.ca) with generous support by the NSERC Experience Award.
