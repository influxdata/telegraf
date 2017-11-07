# JBoss plugin

The JBoss plugin can collect data from JBoss management API.

Plugin currently support JBoss Application server in domain modes:-

- domaincontroller  (Ex http://[jboss-server-ip]:9990/management)

### Configuration:

```toml
# Read flattened metrics from one or more JBoss HTTP endpoints
[[inputs.jboss]]
  ## API endpoint:
  ##
  servers = [
    "http://[jboss-server-ip]:9990/management",
  ]
  ## Execution Mode
	exec_as_domain = false
  ## Username and password
  username = ""
  password = ""

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
 ## Metric selection
  metrics =[
    "jvm",
    "web_con",
    "deployment",
    "database",
    "jms",
  ]
```

Please refer to JBoss management API for full documentation, https://docs.jboss.org/author/display/AS71/The+HTTP+management+API


### Measurements & Fields:

#### Common Tags for all measurements

Tag Name | Description
---------|------------
jboss_host | domain host (only for domain profiles)
jboss_server | instance name (only for domain profiles)



#### Measurement : jboss_jvm
  - (threading info)
    - thread-count (float)
    - peak-thread-count (float)
    - daemon-thread-count (float)
  - (memory info)
    - heap_committed (float)
    - heap_init (float)
    - heap_max (float)
    - heap_used (float)
    - nonheap_committed (float)
    - nonheap_init (float)
    - nonheap_max (float)
    - nonheap_used (float)
  - (garbage-collector)
  There are as many  (count/time) pairs as diferent collection modes
    - [garbage_collector_mode]`_count`  (float)
    - [garbage_collector_mode]`_time` (float)

garbage_colletor_mode could be any of the following depending on the jvm GC configuration   

- PS_MarkSweep
- PS_Scavenge
- ParNew
- ConcurrentMarkSweep
- etc

#### Measurement : jboss_web_con
  - bytesSent (float):	Number of byte sent by the connector
  - bytesReceived(float):	Number of byte received by the connector (POST data).
  - processingTime(float):	Processing time used by the connector. Im milli-seconds.
  - errorCount(float):	Number of error that occurs when processing requests by the connector.
  - maxTime(float):	Max time spent to process a request.
  - requestCount(float):	Number of requests processed by the connector.

  Tag Name | Description
  ---------|------------
  type | could be only "http"  or "ajp"


#### Measurement : jboss_web_app

  - active-sessions (float)
  - expired-sessions (float)
  - max-active-sessions (float)
  - sessions-created (float)


Tag Name | Description
---------|------------
name     | Web Module name (war)
context-root| URL context root for this web app
runtime_name   |  Application Name (ear/war)

#### Measurement : jboss_ejb

  - invocations (float)
  - peak-concurrent-invocations (float)
  - pool-available-count (float)
  - pool-create-count (float)
  - pool-current-size (float)
  - pool-max-size (float)
  - pool-remove-count (float)
  - wait-time (float)


Tag Name | Description
---------|------------
 name|   module Name
 ejb|    EJB name
 runtime_name|  Application Name (ear)

#### Measurement : jboss_database

  - in-use-count (integer)
  - active-count: (integer)
  - available-count:(integer)


   Tag Name | Description
   ---------|------------
    "name"|   datasource name

#### Measurement : jboss_jms

  - message-count (float)  The number of messages currently in this queue.
  - messages-added (float) The number of messages added to this queue since it was created.
  - consumer-count (float) [for Queue stats] The number of consumers consuming messages from this queue.
  - subscription-count (float) [for Topic stats] The number of (durable and non-durable) subscribers for this topic.
  - scheduled-count (float) The number of scheduled messages in this queue.


Tag Name | Description
  ---------|------------
  name|   queue/topic name
