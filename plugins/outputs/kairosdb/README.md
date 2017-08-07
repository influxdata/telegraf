# KairosDB Output Plugin

This plugin writes to [KairosDB](https://kairosdb.github.io/) using 2 protocols:
* [Telnet API](https://kairosdb.github.io/docs/build/html/PushingData.html#submitting-data-via-telnet)
* [REST API](https://kairosdb.github.io/docs/build/html/restapi/index.html)
  * supports either http or https
  * supports basic auth

Only int and float metric values are supported. String values are ignored, others will return an error.

### Configuration
```
[[outputs.kairosdb]]
  # method can be tcp, http, or https
  method = "http"
  host = "kairosdbhost"
  port = "4242"
  # user/password only supported by the REST api
  user = "username"
  password = "pwd"
```
