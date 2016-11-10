# Monit Plugin

The monit plugin gathers metrics and status information about local processes managed and watched over by Monit.

### Configuration:

```toml
# Read metrics and status information about processes managed by Monit
 [[inputs.monit]]
   #SampleConfig
   address = "http://127.0.0.1:2812"
   basic_auth_username = "test"
   basic_auth_password = "test"
```

### Measurements & Fields:

<optional description>

Measurement:  monit

Fields for all Monit service types:
- status
- status_text 
- monitoring_status
- monitoring_status_text

Fields for Monit service type 3:
- cpu_percent
- cpu_percent_total
- mem_kb
- mem_kb_total
- mem_percent
- mem_percent_total
- service_uptime
      
Fields for Monit service type 5:
- cpu_system
- cpu_user
- cpu_wait
- cpu_load_avg_1m
- cpu_load_avg_5m
- cpu_load_avg_15m
- mem_kb
- mem_percent
- swap_kb
- swap_percent

### Tags:
All measurements have the following tags:
- address
- version
- service
- service_type

### Example Output:
```
$ ./telegraf -config telegraf.conf -input-filter monit -test
monit,address=http://127.0.0.1:2812,host=ubuntu,service=telegraf,service_type=3,version=5.20.0 cpu_percent=0,cpu_percent_total=0,mem_kb=0i,mem_kb_total=0i,mem_percent=0,mem_percent_total=0,monitoring_status=1i,monitoring_status_decoded="Running",service_uptime=0i,status=4608i,status_decoded="Failure" 1478782321000000000
```
