### Cloud Run Output Plugin

 This plugin is used to send time series metrics into a proxy, located behind the Google Cloud Run service, with OAuth2 enabled.         

There are two required fields, the URL where your Cloud Run application lives, and the JSON secret key file location, which is [generated](https://cloud.google.com/iam/docs/creating-managing-service-account-keys) by the user in their Google Cloud Run project. 

Below is an example of how to send metrics into a proxy located in Cloud Run:

```
  ## A plugin that can transmit metrics over OAuth2
  ## URL is the Cloud Run Wavefront proxy address to send metrics to
  # url = "http://127.0.0.1:8080/telegraf"

  ## Timeout for Cloud Run message, suggested as 30s to account for handshaking
  # timeout = "30s"

  ## Cloud Run JSON file location
  ## This is the location of the JSON file generated from your GCP project that's authorized to send
  ## metrics into CloudRun.
  ## Windows users, note that you need to use forward slashes.
  # credentials_file = "/etc/telegraf/example_secret.json"

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "wavefront"

  ## NOTE: The default headers have already been set to the following by default:
  ## defaultContentType   = "application/octet-stream"
  ## defaultAccept        = "application/json"
  ## defaultMethod        = http.MethodPost
```

Developed by Casey Flanigan, Zachary Mares, and John Farrington
