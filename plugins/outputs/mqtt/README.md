# MQTT Producer Output Plugin

This plugin writes to a [MQTT Broker](http://http://mqtt.org/) acting as a mqtt Producer.

```
[[outputs.mqtt]]
  ## URLs of mqtt brokers
  servers = ["localhost:1883"]
  
  ## topic for producer messages
  topic = "telegraf"
  
  ## QoS policy for messages
  qos = 2
  
  ## username and password to connect MQTT server.
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"
   
  ## client ID, if not set a random ID is generated
  # client_id = ""
    
  ## Timeout for write operations. default: 5s
  # timeout = "5s"
  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
  
  ## Data format to output.
  data_format = "influx"


```

### Required parameters:

* `servers`: List of strings, this is for speaking to a cluster of `mqtt` brokers. On each flush interval, Telegraf will randomly choose one of the urls to write to. Each URL should just include host and port e.g. -> `["{host}:{port}","{host2}:{port2}"]`
* `topic_prefix`: The `mqtt` topic prefix to publish to. MQTT outputs send metrics to this topic format "<topic_prefix>/<hostname>/<pluginname>/" ( ex: prefix/web01.example.com/mem)
* `qos`: The `mqtt` QoS policy for sending messages. See https://www.ibm.com/support/knowledgecenter/en/SSFKSJ_9.0.0/com.ibm.mq.dev.doc/q029090_.htm for details.

### Optional parameters:
* `username`: The username to connect MQTT server.
* `password`: The password to connect MQTT server.
* `client_id`: The unique client id to connect MQTT server. If this paramater is not set then a random ID is generated.
* `timeout`: Timeout for write operations. default: 5s
* `ssl_ca`: SSL CA
* `ssl_cert`: SSL CERT
* `ssl_key`: SSL key
* `insecure_skip_verify`: Use SSL but skip chain & host verification (default: false)
* `data_format`: [About Telegraf data formats](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md)
