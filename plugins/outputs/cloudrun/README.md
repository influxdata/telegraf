### Cloud Run Output Plugin

This plugin can be used to send time series metrics into a metrics proxy that is located in the Google Cloud Run service, with OAuth2 enabled, and includes token refresh.         

There are two required fields, the URL where your Cloud Run application lives, and the JSON secret key file location, which is [generated](https://cloud.google.com/iam/docs/creating-managing-service-account-keys) by the user in their Google Cloud Run project. 

Below is an example of how to send metrics into a proxy located in Cloud Run:

```
# A plugin that can transmit metrics over OAuth2
[[outputs.cloudrun]]
  ## URL is the address to send metrics to
  url = 'https://metrics-proxy-abc123-uc.a.run.app'

  ## Timeout for HTTP message
  timeout = "30s"
  
  json_file_location = "C:/Path/To/Secrets/file.json"
  data_format = "influx"
```

Developed by Casey Flanigan, Zachary Mares, and John Farrington
