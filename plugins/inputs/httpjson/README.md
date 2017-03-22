# HTTP JSON Plugin

The httpjson plugin can collect data from remote URLs which respond with JSON. Then it flattens JSON and finds all numeric values, treating them as floats.

For example, if you have a service called _mycollector_, which has HTTP endpoint for gathering stats at http://my.service.com/_stats, you would configure the HTTP JSON plugin like this:

```
[[inputs.httpjson]]
  name = "mycollector"

  servers = [
    "http://my.service.com/_stats"
  ]

  # HTTP method to use (case-sensitive)
  method = "GET"

  # Set response_timeout (default 5 seconds)
  response_timeout = "5s"
```

`name` is used as a prefix for the measurements.

`method` specifies HTTP method to use for requests.

`response_timeout` specifies timeout to wait to get the response

You can also specify which keys from server response should be considered tags:

```
[[inputs.httpjson]]
  ...

  tag_keys = [
    "role",
    "version"
  ]
```

If the JSON response is an array of objects, then each object will be parsed with the same configuration.

You can also specify additional request parameters for the service:

```
[[inputs.httpjson]]
  ...

 [inputs.httpjson.parameters]
    event_type = "cpu_spike"
    threshold = "0.75"

```

You can also specify additional request header parameters for the service:

```
[[inputs.httpjson]]
  ...

 [inputs.httpjson.headers]
    X-Auth-Token = "my-xauth-token"
    apiVersion = "v1"
```

# Example:

Let's say that we have a service named "mycollector" configured like this:

```
[[inputs.httpjson]]
  name = "mycollector"
  servers = [
    "http://my.service.com/_stats"
  ]
  # HTTP method to use (case-sensitive)
  method = "GET"
  tag_keys = ["service"]
```

which responds with the following JSON:

```json
{
    "service": "service01",
    "a": 0.5,
    "b": {
        "c": "some text",
        "d": 0.1,
        "e": 5
    }
}
```

The collected metrics will be:
```
httpjson_mycollector_a,service='service01',server='http://my.service.com/_stats' value=0.5
httpjson_mycollector_b_d,service='service01',server='http://my.service.com/_stats' value=0.1
httpjson_mycollector_b_e,service='service01',server='http://my.service.com/_stats' value=5
```

# Example 2, Multiple Services:

There is also the option to collect JSON from multiple services, here is an example doing that.

```
[[inputs.httpjson]]
  name = "mycollector1"
  servers = [
    "http://my.service1.com/_stats"
  ]
  # HTTP method to use (case-sensitive)
  method = "GET"

[[inputs.httpjson]]
  name = "mycollector2"
  servers = [
    "http://service.net/json/stats"
  ]
  # HTTP method to use (case-sensitive)
  method = "POST"
```

The services respond with the following JSON:

mycollector1:
```json
{
    "a": 0.5,
    "b": {
        "c": "some text",
        "d": 0.1,
        "e": 5
    }
}
```

mycollector2:
```json
{
    "load": 100,
    "users": 1335
}
```

The collected metrics will be:

```
httpjson_mycollector1_a,server='http://my.service.com/_stats' value=0.5
httpjson_mycollector1_b_d,server='http://my.service.com/_stats' value=0.1
httpjson_mycollector1_b_e,server='http://my.service.com/_stats' value=5

httpjson_mycollector2_load,server='http://service.net/json/stats' value=100
httpjson_mycollector2_users,server='http://service.net/json/stats' value=1335
```

# Example 3, Multiple Metrics in Response:

The response JSON can be treated as an array of data points that are all parsed with the same configuration.

```
[[inputs.httpjson]]
  name = "mycollector"
  servers = [
    "http://my.service.com/_stats"
  ]
  # HTTP method to use (case-sensitive)
  method = "GET"
  tag_keys = ["service"]
```

which responds with the following JSON:

```json
[
    {
        "service": "service01",
        "a": 0.5,
        "b": {
            "c": "some text",
            "d": 0.1,
            "e": 5
        }
    },
    {
        "service": "service02",
        "a": 0.6,
        "b": {
            "c": "some text",
            "d": 0.2,
            "e": 6
        }
    }
]
```

The collected metrics will be:
```
httpjson_mycollector_a,service='service01',server='http://my.service.com/_stats' value=0.5
httpjson_mycollector_b_d,service='service01',server='http://my.service.com/_stats' value=0.1
httpjson_mycollector_b_e,service='service01',server='http://my.service.com/_stats' value=5
httpjson_mycollector_a,service='service02',server='http://my.service.com/_stats' value=0.6
httpjson_mycollector_b_d,service='service02',server='http://my.service.com/_stats' value=0.2
httpjson_mycollector_b_e,service='service02',server='http://my.service.com/_stats' value=6
```
