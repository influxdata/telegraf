# Telegraf Plugin: NATS Monitor

NATS Server [Monitoring](http://nats.io/documentation/server/gnatsd-monitoring/)

### Configuration:

```toml
# Read metrics from one or many NATS servers
[[inputs.nats_monitor]]
  ## An array of NATS monitors.
  # urls = ["http://localhost:8222"]
```

### Measurements & Fields:

- nats_varz
    - cpu
	- mem
	- subscriptions
	- connections
	- in_msgs
	- out_msgs
	- in_bytes
	- out_bytes

### Tags:

- All measurements have the following tags:
    - url   
    
### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter nats_monitor -test
* Plugin: inputs.nats_monitor, Collection 1
> nats_varz,host=ing,url=http://localhost:8222 \
out_msgs=2050000000i,in_msgs=211100000i,in_bytes=6577600000i,out_bytes=64800000000i,\
cpu=0i,mem=7278592i,subscriptions=0i,connections=0i 1487340557000000000
```
