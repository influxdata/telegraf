# Circonus Output Plugin

This plugin writes metrics data to the Circonus platform. In order to use this
plugin, an HTTPTrap check must be configured on a Circonus broker. This check
can be automatically created by the plugin or manually configured (see the
plugin configuration information). For information about Circonus HTTPTrap
check configuration click [here][docs].

### Configuration

```toml
[[outputs.circonus]]
  ## Connection timeout:
  # timeout = "5s"

  ## Checks is a map of regexp patterns and submission URL's of Circonus
  ## HTTPTrap checks to which metrics with names mattching the patterns will
  ## be sent:
  # checks = { ".*" = "https://broker1.example.net:43191/module/httptrap/11223344-5566-7788-9900-aabbccddeeff/example" }
  
  ## If the CID of a broker is provided:
  # broker = "/broker/1"
  ## or automatic broker lookup can be used if broker is set to "auto":
  # broker = "auto"
  ## brokers can be excluded by adding their CID to the exclude list:
  # exclude_brokers = [ "/broker/2" ]
  ## then a check can be automatically created for metrics collected with
  ## this Telegraf plugin by entering "auto" for the submission URL:
  checks = { ".*" = "auto" }
  
  ## Optional Broker TLS Configuration, note any brokers used by this plugin
  ## must share the same CA and certificate files, if this info is not provided,
  ## the broker CA data will be retrived using the API:
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification:
  # insecure_skip_verify = false

  ## Circonus API Configuration, this is required for automatic check creation
  ## and automatic check lookup, and retrieving broker CA information:
  # api_url = "https://api.circonus.com/"
  # api_token = "11223344-5566-7788-9900-aabbccddeeff"
  # api_app = "telegraf"
  ## Optional API TLS Configuration: 
  # api_tls_ca = "/etc/telegraf/api_ca.pem"
  ## Use TLS but skip chain & host verification:
  # api_insecure_skip_verify = false
```

### Configuration Options

|Setting|Description|
|-------|-----------|
|`timeout`|The timeout to use when connecting to the Circonus broker.|
|`broker`|The CID of a Circonus broker to use when automatically creating a check. If omitted or set to `"auto"`, then a random eligible broker will be selected.|
|`exclude_brokers`|A list of broker CID's that will be excluded from automatic broker selection.|
|`checks`|A map of regexp patterns to Circonus check submission URL's. The regexp pattern will match to the name of the metric that will be populated in the Circonus system, and this can include tag matching (i.e. category:value). The submission URL should match the submission URL of an HTTPTrap check as shown in the Circonus UI. If the submission check is set to `"auto"` an HTTPTrap check named `telegraf-httptrap` will be used on the broker specified in the previous settings. This check will be created if it does not exist. This setting is required.|
|`tls_ca`|The certificate authority file to use when connecting to the Circonus broker. If this is not provided, the CA information will be retrieved from the Circonus API.|
|`tls_cert`|A TLS certificate file to use when connecting to the Circonus broker. This is optional.|
|`tls_key`|A TLS key file to use when connecting to the Circonus broker. This is optional.|
|`insecure_skip_verify`|This will skip TLS verification when connecting to the Circonus broker. This should only be set to `true` for testing purposes.|
|`api_token`|The authentication token to used when connecting to the Circonus API. It is recommended to create a token/application combination specifically for use with this Telegraf plugin. This is required.|
|`api_url`|The URL that can be used to connect to the Circonus API. This will default to the Circonus SaaS API URL if not provided.|
|`api_app`|The API token application to use when connecting to the Circonus API. This will default to `telegraf` if not provided.|
|`api_tls_ca`|The certificate authority file to use when connecting to the Circonus API, if needed.|
|`api_insecure_skip_verify`|This will skip TLS verification when conneciting to the Circonus API. This should only be set to `true` for testing purposes.|

[docs]: https://docs.circonus.com/circonus/checks/check-types/httptrap
