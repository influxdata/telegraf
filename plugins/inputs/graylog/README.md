# GrayLog plugin

The Graylog plugin can collect data from remote Graylog service URLs. 

Plugin currently support two type of end points:-

- multiple  (Ex http://[graylog-server-ip]:12900/system/metrics/multiple)
- namespace (Ex http://[graylog-server-ip]:12900/system/metrics/namespace/{namespace})

End Point can be a mixe of one  multiple end point  and several namespaces end points


Note: if namespace end point specified metrics array will be ignored for that call.

Sample configration
```
[[inputs.graylog]]
  ## API End Point, currently supported API:
  ## - multiple  (Ex http://[graylog-server-ip]:12900/system/metrics/multiple)
  ## - namespace (Ex http://[graylog-server-ip]:12900/system/metrics/namespace/{namespace})
  ## Note if namespace end point specified metrics array will be ignored for that call.
  ## End point can contain namespace and multiple type calls
  ## Please check http://[graylog-server-ip]:12900/api-browser for full list end points

  servers = [
    "http://10.224.162.16:12900/system/metrics/multiple"
  ]

  #Metrics define metric which will be pulled from GrayLog and reported to the defined Output 
  metrics = [
    "jvm.cl.loaded",
    "jvm.memory.pools.Metaspace.committed"
  ]
  ## User name and password  
  username = "put-username-here"
  password = "put-password-here"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
```

Please refer to GrayLog metrics api browser for full metric end points http://10.224.162.16:12900/api-browser
