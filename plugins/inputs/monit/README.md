# Monit Plugin

The monit plugin gathers metrics and status information about local processes, 
remote hosts, file, file systems, directories and network interfaces managed and watched over by Monit.

To install Monit agent on the host please refer to the link https://mmonit.com/wiki/Monit/Installation

Minimum Version of Monit tested with is 5.16

### Configuration:

```toml
# Read metrics and status information about processes managed by Monit
 [[inputs.monit]]
   #SampleConfig
   address = "http://127.0.0.1:2812"
   basic_auth_username = "test"
   basic_auth_password = "test"
```

### Tags:
All measurements have the following tags:
- address
- version
- service
- paltform_name
- status
- monitoring_status
- monitoring_mode

### Measurements & Fields:

<optional description>

### Fields:
Fields for all Monit service types:
- status_code
- monitoring_status_code
- monitoring_mode_code 

### Measurement & Fields:
Fields for Monit service type Filesystem:
- Measurement:
  - monit_filesystem
- Fields:
  - mode
  - block_percent
  - block_usage
  - block_total
  - inode_percent
  - inode_usage
  - inode_total

Fields for Monit service type directory:
- Measurement:
  - monit_directory
- Fields:
  - permissions

Fields for Monit service type file:
- Measurement:
  - monit_file
- Fields:
  - size
  - permissions

Fields for Monit service type process:
- Measurement:
  - monit_process
- Fields:
  - cpu_percent
  - cpu_percent_total
  - mem_kb
  - mem_kb_total
  - mem_percent
  - mem_percent_total
  - pid
  - parent_pid
  - threads
  - children

Fields for Monit service type remote host:
- Measurement:
  - monit_remote_host
- Fields:
  - hostname
  - port_number
  - request
  - protocol
  - type

Fields for Monit service type system:
- Measurement:
  - monit_system
- Fields:
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

Fields for Monit service type fifo:
- Measurement:
  - monit_fifo
- Fields:
  - permissions

Fields for Monit service type program:
- Measurement:
  - monit_program
- Fields:
  - last_started_time
  - program_status

Fields for Monit service type network:
- Measurement:
  - monit_network
- Fields:
  - link_state
  - link_mode
  - link_speed
  - download_packets_now
  - download_packets_total
  - download_bytes_now
  - download_bytes_total
  - download_errors_now
  - download_errors_total
  - upload_packets_now
  - upload_packets_total
  - upload_bytes_now
  - upload_bytes_total
  - upload_errors_now
  - upload_errors_total

### Example Output:
```
$ ./telegraf -config telegraf.conf -input-filter monit -test
monit_system,address=http://localhost:2812,host=verizon-onap,hostname=verizon-onap,monitoring_mode=Monitoring\ mode:\ \ active,monitoring_status=Monitoring\ status:\ \ Monitored,platform_name=Linux,service=verizon-onap,status=Running,version=5.16 status_code=0i,cpu_system=1.9,cpu_user=4.7,cpu_wait=1.5,cpu_load_avg_1m=1.24,cpu_load_avg_5m=1.68,mem_percent=67.1,monitoring_status_code=1i,monitoring_mode_code=0i,cpu_load_avg_15m=1.64,mem_kb=10961012i,swap_kb=2322688,swap_percent=13.9 1578636430000000000
monit_remote_host,address=http://localhost:2812,host=verizon-onap,hostname=verizon-onap,monitoring_mode=Monitoring\ mode:\ \ passive,monitoring_status=Monitoring\ status:\ \ Monitored,platform_name=Linux,service=testing,status=Failure,version=5.16 status_code=32i,monitoring_status_code=1i,monitoring_mode_code=1i,remote_hostname="192.168.10.49",port_number=2220i,request="",protocol="DEFAULT",type="TCP" 1578636430000000000
monit_fifo,address=http://localhost:2812,host=verizon-onap,hostname=verizon-onap,monitoring_mode=Monitoring\ mode:\ \ active,monitoring_status=Monitoring\ status:\ \ Monitored,platform_name=Linux,service=test2,status=Running,version=5.16 status_code=0i,monitoring_status_code=1i,monitoring_mode_code=0i,permissions=664i 1578636430000000000
monit_network,address=http://localhost:2812,host=verizon-onap,hostname=verizon-onap,monitoring_mode=Monitoring\ mode:\ \ active,monitoring_status=Monitoring\ status:\ \ Monitored,platform_name=Linux,service=test1,status=Failure,version=5.16 monitoring_status_code=1i,monitoring_mode_code=0i,download_packets_total=0i,upload_bytes_now=0i,download_errors_total=0i,status_code=8388608i,link_speed=-1i,link_mode="Unknown Mode",download_bytes_now=0i,download_bytes_total=0i,download_errors_now=0i,upload_packets_total=0i,upload_bytes_total=0i,upload_errors_now=0i,upload_errors_total=0i,link_state=0i,download_packets_now=0i,upload_packets_now=0i 1578636430000000000
monit_directory,address=http://localhost:2812,host=verizon-onap,hostname=verizon-onap,monitoring_mode=Monitoring\ mode:\ \ passive,monitoring_status=Monitoring\ status:\ \ Monitored,platform_name=Linux,service=test,status=Running,version=5.16 status_code=0i,monitoring_status_code=1i,monitoring_mode_code=1i,permissions=755i 1578636430000000000
```
