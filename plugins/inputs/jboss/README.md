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

  ## Username and password
  username = ""
  password = ""

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
```

Please refer to JBoss management API for full documentation, https://docs.jboss.org/author/display/AS71/The+HTTP+management+API

