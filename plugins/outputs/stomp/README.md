# STOMP Producer Output Plugin

This plugin writes to a [Active MQ Broker](http://activemq.apache.org/) Applying STOMP Protocol.

It also support Amazon MQ  (https://aws.amazon.com/amazon-mq/)


## CONFIGURATON

```toml
[[outputs.STOMP]]
  ## Host of Active Mq broker
    host = "localhost:61613"

  ## Queue name for producer messages
    queueName = "telegraf"


  ## Optional username and password if Required to connect Active MQ server.
    username = ""
    password = ""


  ## Default No TLS Connecton 
    # SSL = false

  ## Optional TLS Config
    # SSL = true
    # tls_ca = "/etc/telegraf/ca.pem"
    # tls_cert = "/etc/telegraf/cert.pem"
    # tls_key = "/etc/telegraf/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false



  ## Data format to output.
    # data_format = "json"
```

