# HTTP Json Output Plugin

This output plugin allow you to write metrics to custom storage through JSON API

### Request format example:

Your API will expect to get request body with this format

```json
{
  "metrics": [
    {
      "name": "measurement", // Measurement
      "fields": "value=0.64", // Fields value
      "tags": "tag1=tag1, tag2=tag2", // Tag keys
      "time": "10000020" // Time will be UNIX timestampt format
    },
    ...
  ],
  "data": {
    "data": "your additional data"
  }
}

```


### Configuration:

```toml
# Configuration for sending metrics with HTTP Json Output Plugin
[[outputs.httpjson]]
  ## Setup your HTTP Json service name
  # name = "your_httpjson_service_name"

  ## Set the target server. The URL must be a valid HTTP(s) URL
  # server = "http://localhost:3000"

  ## Setup additional data you want to sent along with the metrics data
  ## All value must be string
  # [outputs.httpjson.data]
  #   authToken = "12345"

  ## Setup additional headers for the HTTP(s) request
  ## All value must be string
  # [outputs.httpjson.headers]
  #   Content-Type = "application/json;charset=UTF-8"
```


