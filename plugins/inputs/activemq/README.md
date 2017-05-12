# Telegraf plugin: Activemq

#### Plugin arguments:
- **urls** []string: List of activemq metrics URLs to collect from. Default is "http://localhost:8161/api/jolokia/read/org.apache.activemq:type=Broker,brokerName=localhost".
- **username** string: Username for HTTP basic authentication
- **password** string: Password for HTTP basic authentication
- **timeout** duration: time that the HTTP connection will remain waiting for response. Defalt 4 seconds ("4s")

##### Optional SSL Config

- **ssl_ca** string: the full path for the SSL CA certicate
- **ssl_cert** string: the full path for the SSL certificate
- **ssl_key** string: the full path for the key file
- **insecure_skip_verify** bool: if true HTTP client will skip all SSL verifications related to peer and host. Default to false

#### Description

The Activemq plugin collects from the /api/jolokia/read/org.apache.activemq:type=Broker,brokerName=localhost URL.

# Measurements:

Meta:
- tags: `port=<port>`, `server=url`

- TotalConnectionsCount
- TotalProducerCount
- CurrentConnectionsCount
- TotalDequeueCount
- AverageMessageSize
- MinMessageSize
- TotalConsumerCount
- MaxMessageSize
- TotalMessageCount
- MemoryPercentUsage
- TotalEnqueueCount

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter activemq -test
* Plugin: inputs.activemq, Collection 1
> activemq,host=luoxiaojun1992-OptiPlex-7020,port=8161,server=119.254.210.186 AverageMessageSize=1601,CurrentConnectionsCount=66i,MaxMessageSize=4376,MemoryPercentUsage=0,MinMessageSize=1024,TotalConnectionsCount=3939i,TotalConsumerCount=3i,TotalDequeueCount=3965i,TotalEnqueueCount=11753i,TotalMessageCount=0i,TotalProducerCount=0i 1480739077000000000
```
