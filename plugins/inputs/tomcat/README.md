# Tomcat Input Plugin

The Tomcat plugin collects statistics available from the tomcat manager status
page from the `http://<host>/manager/status/all?XML=true URL.` (`XML=true` will
return only xml data).

See the [Tomcat documentation][1] for details of these statistics.

[1]: https://tomcat.apache.org/tomcat-9.0-doc/manager-howto.html#Server_Status

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather metrics from the Tomcat server status page.
[[inputs.tomcat]]
  ## URL of the Tomcat server status
  # url = "http://127.0.0.1:8080/manager/status/all?XML=true"

  ## HTTP Basic Auth Credentials
  # username = "tomcat"
  # password = "s3cret"

  ## Request timeout
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Metrics

- tomcat_jvm_memory
  - free
  - max
  - total
- tomcat_jvm_memorypool
  - committed
  - init
  - max
  - used
- tomcat_connector
  - bytes_received
  - bytes_sent
  - current_thread_busy
  - current_thread_count
  - error_count
  - max_threads
  - max_time
  - processing_time
  - request_count

### Tags

- tomcat_jvm_memory
  - source
- tomcat_jvm_memorypool has the following tags:
  - name
  - type
  - source
- tomcat_connector
  - name
  - source

## Example Output

```text
tomcat_jvm_memory,host=N8-MBP free=20014352i,max=127729664i,total=41459712i 1474663361000000000
tomcat_jvm_memorypool,host=N8-MBP,name=Eden\ Space,type=Heap\ memory committed=11534336i,init=2228224i,max=35258368i,used=1941200i 1474663361000000000
tomcat_jvm_memorypool,host=N8-MBP,name=Survivor\ Space,type=Heap\ memory committed=1376256i,init=262144i,max=4390912i,used=1376248i 1474663361000000000
tomcat_jvm_memorypool,host=N8-MBP,name=Tenured\ Gen,type=Heap\ memory committed=28549120i,init=5636096i,max=88080384i,used=18127912i 1474663361000000000
tomcat_jvm_memorypool,host=N8-MBP,name=Code\ Cache,type=Non-heap\ memory committed=6946816i,init=2555904i,max=251658240i,used=6406528i 1474663361000000000
tomcat_jvm_memorypool,host=N8-MBP,name=Compressed\ Class\ Space,type=Non-heap\ memory committed=1966080i,init=0i,max=1073741824i,used=1816120i 1474663361000000000
tomcat_jvm_memorypool,host=N8-MBP,name=Metaspace,type=Non-heap\ memory committed=18219008i,init=0i,max=-1i,used=17559376i 1474663361000000000
tomcat_connector,host=N8-MBP,name=ajp-bio-8009 bytes_received=0i,bytes_sent=0i,current_thread_count=0i,current_threads_busy=0i,error_count=0i,max_threads=200i,max_time=0i,processing_time=0i,request_count=0i 1474663361000000000
tomcat_connector,host=N8-MBP,name=http-bio-8080 bytes_received=0i,bytes_sent=86435i,current_thread_count=10i,current_threads_busy=1i,error_count=2i,max_threads=200i,max_time=167i,processing_time=245i,request_count=15i 1474663361000000000
```
