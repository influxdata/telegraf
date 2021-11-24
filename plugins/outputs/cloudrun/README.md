### CloudRun


CloudRun Conf file syntax:
```
# A plugin that can transmit metrics over HTTP
[[outputs.cloudrun]]
  ## URL is the address to send metrics to
  url = 'https://metrics-proxy-abc123-uc.a.run.app'

  ## Timeout for HTTP message
  timeout = "30s"
  
  json_file_location = "C:/Path/To/Secrets/file.json"
  data_format = "wavefront"
  convert_paths = false
```

Developed by Casey Flanigan, Zachary Mares, and John Farrington
