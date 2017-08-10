# Tomcat Input Plugin

The Tomcat plugin collects statistics available from the tomcat manager status page from the `http://<host>/manager/status/all?XML=true URL.`
(`XML=true` will return only xml data). See the [Tomcat documentation](https://tomcat.apache.org/tomcat-9.0-doc/manager-howto.html#Server_Status) for details of these statistics.

### Configuration:

```toml
# A Telegraf plugin to collect tomcat metrics.
[[inputs.tomcat]]
  # A Tomcat status URI to gather stats.
  # Default is "http://127.0.0.1:8080/manager/status/all?XML=true".
  url = "http://127.0.0.1:8080/manager/status/all?XML=true"
  # Credentials for status URI.
  # Default is tomcat/s3cret.
  username = "tomcat"
  password = "s3cret"
```

### Measurements & Fields:

- tomcat\_jvm\_memory
    - free
    - total
    - max
- tomcat\_jvm\_memorypool
  - max\_threads
  - current\_thread\_count
  - current\_threads\_busy
  - max\_time
  - processing\_time
  - request\_count
  - error\_count
  - bytes\_received
  - bytes\_sent
- tomcat\_connector
  - max\_threads
  - current\_thread\_count
  - current\_thread\_busy
  - max\_time
  - processing\_time
  - request\_count
  - error\_count
  - bytes\_received
  - bytes\_sent

### Tags:

- tomcat\_jvm\_memorypool has the following tags:
  - name
  - type
- tomcat\_connector
  - name

### Sample Queries:

TODO

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter tomcat -test
* Plugin: tomcat, Collection 1
> tomcat_jvm_memory,host=N8-MBP free=20014352i,max=127729664i,total=41459712i 1474663361000000000
> tomcat_jvm_memorypool,host=N8-MBP,name=Eden\ Space,type=Heap\ memory committed=11534336i,init=2228224i,max=35258368i,used=1941200i 1474663361000000000
> tomcat_jvm_memorypool,host=N8-MBP,name=Survivor\ Space,type=Heap\ memory committed=1376256i,init=262144i,max=4390912i,used=1376248i 1474663361000000000
> tomcat_jvm_memorypool,host=N8-MBP,name=Tenured\ Gen,type=Heap\ memory committed=28549120i,init=5636096i,max=88080384i,used=18127912i 1474663361000000000
> tomcat_jvm_memorypool,host=N8-MBP,name=Code\ Cache,type=Non-heap\ memory committed=6946816i,init=2555904i,max=251658240i,used=6406528i 1474663361000000000
> tomcat_jvm_memorypool,host=N8-MBP,name=Compressed\ Class\ Space,type=Non-heap\ memory committed=1966080i,init=0i,max=1073741824i,used=1816120i 1474663361000000000
> tomcat_jvm_memorypool,host=N8-MBP,name=Metaspace,type=Non-heap\ memory committed=18219008i,init=0i,max=-1i,used=17559376i 1474663361000000000
> tomcat_connector,host=N8-MBP,name=ajp-bio-8009 bytes_received=0i,bytes_sent=0i,current_thread_count=0i,current_threads_busy=0i,error_count=0i,max_threads=200i,max_time=0i,processing_time=0i,request_count=0i 1474663361000000000
> tomcat_connector,host=N8-MBP,name=http-bio-8080 bytes_received=0i,bytes_sent=86435i,current_thread_count=10i,current_threads_busy=1i,error_count=2i,max_threads=200i,max_time=167i,processing_time=245i,request_count=15i 1474663361000000000
```
