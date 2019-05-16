# Logz.io Output Plugin

This plugin sends metrics to Logz.io over HTTPs.

### Configuration:

```toml
[[outputs.logzio]]
  ## Logz.io account token
  token = "your Logz.io token" # required

  ## Use your listener URL for your Logz.io account region.
  # url = "https://listener.logz.io:8071"
  
  ## Timeout for HTTP requests
  # timeout = "5s"
```

### Required parameters:

* `token`: Your Logz.io token, which can be found under "settings" in your account.

### Optional parameters:
* `url`: Logz.io listener URL.
* `timeout` : Time limit for requests made by Logz.io client