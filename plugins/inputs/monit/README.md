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
process,address=http://127.0.0.1:2812,host=ubuntu,hostname=telegraf,platform_name=Linux,service=zookeeper,service_type=3,version=5.17.1 monitoring_mode="Active",cpu_percent_total=0,status="Running",monitoring_status_code=1i,mem_kb=174272i,service_uptime=11344202i,pid=17566i,parent_pid=0i,threads=31i,monitoring_status="Running",mem_percent_total=0.7,children=0i,status_code=0i,cpu_percent=0,mem_kb_total=174272i,mem_percent=0.7 1572256571000000000
file,address=http://127.0.0.1:2812,host=ubuntu,hostname=telegraf,platform_name=Linux,service=jboss-console-log,service_type=2,version=5.17.1 monitoring_status_code=1i,monitoring_mode="Active",monitoring_status="Accessible",size=1526i,permissions=644i,status_code=0i,status="Running" 1572256571000000000
program,address=http://127.0.0.1:2812,host=ubuntu,hostname=telegraf,platform_name=Linux,service=zookeeper-status,service_type=7,version=5.17.1 output="Zookeeper server is running on localhost. Zookeeper server status check query successful on localhost. Zookeeper Mode - follower",status_code=0i,status="Running",monitoring_status_code=1i,monitoring_mode="Active",last_started_time=1572256563i,program_status="Status OK" 1572256571000000000
```
