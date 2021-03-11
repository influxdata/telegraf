# JBoss plugin

The JBoss plugin can collect data from JBoss management API.

Plugin currently support JBoss Application server in domain modes:-

- domaincontroller  (Ex http://[jboss-server-ip]:9990/management)

Currently tested on JBoss EAP 6.X , JBoss AS 7.X

### Configuration:

```toml
# Read flattened metrics from one or more JBoss HTTP endpoints
[[inputs.jboss]]
  ## API endpoint:
  ##
  servers = [
    "http://localhost:9990/management",
  ]

  ## Username and password
  #username = ""
  #password = ""

  ## Optional SSL Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
  
  ## Metric selection
  metrics =[
    "jvm",
    "web",
    "deployment",
    "database",
    "transaction",
    "jms",
  ]
```

Some examples from JBoss:
https://docs.jboss.org/author/display/WFLY10/The%20HTTP%20management%20API.html


### Metrics:

The jboss_jvm measurement has static fields (those related to threading and memory info ) and dinamic fields ( those related to garbage-collector info) , these dinamic fields has as many  (count/time) pairs as diferent collection modes.

These fields will be in the form

      - [garbage_collector_mode]`_count`  (float)
      - [garbage_collector_mode]`_time` (float)

garbage_colletor_mode could be any of the following depending on the jvm GC configuration

      - PS_MarkSweep
      - PS_Scavenge
      - ParNew
      - ConcurrentMarkSweep
      - etc


- jboss_jvm
    - tags:
        - jboss_host  domain host (only for domain profiles)
        - jboss_server | instance name (only for domain profiles)
    - fields:
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
            - [garbage_collector_mode]`_count`  (float)
            - [garbage_collector_mode]`_time` (float)

- jboss_web_con
    - tags:
        - jboss_host  domain host (only for domain profiles)
        - jboss_server  instance name (only for domain profiles)
        - type  could be only "http"  or "ajp"
    - fields:
        - bytesSent (float):	Number of byte sent by the connector
        - bytesReceived(float):	Number of byte received by the connector (POST data).
        - processingTime(float):	Processing time used by the connector. Im milli-seconds.
        - errorCount(float):	Number of error that occurs when processing requests by the connector.
        - maxTime(float):	Max time spent to process a request.
        - requestCount(float):	Number of requests processed by the connector.

- jboss_web_app
    - tags:
        - jboss_host  domain host (only for domain profiles)
        - jboss_server  instance name (only for domain profiles)
        - name  Web Module name (war)
        - context-root| URL context root for this web app
        - runtime_name   |  Application Name (ear/war)
    - fields:
        - active-sessions (float)
        - expired-sessions (float)
        - max-active-sessions (float)
        - sessions-created (float)

- jboss_ejb
    - tags:
        - jboss_host  domain host (only for domain profiles)
        - jboss_server  instance name (only for domain profiles)
        - name   module Name
        - ejb   EJB name
        - runtime_name  Application Name (ear)
    - fields:
        - invocations (float)
        - peak-concurrent-invocations (float)
        - pool-available-count (float)
        - pool-create-count (float)
        - pool-current-size (float)
        - pool-max-size (float)
        - pool-remove-count (float)
        - wait-time (float)

- jboss_database
    - tags:
        - jboss_host  domain host (only for domain profiles)
        - jboss_server  instance name (only for domain profiles)
        - name  datasource name
    - fields
        - in-use-count (integer)
        - active-count: (integer)
        - available-count:(integer)

- jboss_jms
    - tags:
        - jboss_host  domain host (only for domain profiles)
        - jboss_server  instance name (only for domain profiles)
        - name   queue/topic name
    - fields:
        - message-count (float)  The number of messages currently in this queue.
        - messages-added (float) The number of messages added to this queue since it was created.
        - consumer-count (float) [for Queue stats] The number of consumers consuming messages from this queue.
        - subscription-count (float) [for Topic stats] The number of (durable and non-durable) subscribers for this topic.
        - scheduled-count (float) The number of scheduled messages in this queue.

- jboss_transaction
    - tags:
        - jboss_host  domain host (only for domain profiles)
        - jboss_server  instance name (only for domain profiles)
    - fields:
        - number-of-aborted-transactions (integer)  The number of aborted (ie rolledback) transactions.
        - number-of-application-rollbacks (integer) The number of transactions that have been aborted by application request.
        - number-of-committed-transactions (integer) The number of commited transactions.
        - number-of-heuristics (integer) The number of transactions which have termineted with heuristics outcome.
        - number-of-inflight-transactions (integer) The number of transactions that have begun but not yet terminated.
        - number-of-nested-transactions (integer) The number of nested (sub) transactions created.
        - number-of-resource-rollbacks (integer) The number of transactions that have been rolled back due to resource failure.
        - number-of-system-rollbacks (integer) The number of transactions that have been rolled back due to internal system errors.
        - number-of-timed-out-transactions (integer) The number of transactions that have been rolled back due to timeout.
        - number-of-transactions (integer) The total number of transactions created.
    
