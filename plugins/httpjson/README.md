# HTTP JSON Plugin

The httpjson plugin can collect data from remote URLs which respond with JSON. Then it flattens JSON and finds all numeric values, treating them as floats.

For example, if you have a service called _mycollector_, which has HTTP endpoint for gathering stats http://my.service.com/_stats:

```
[[httpjson.services]]
  name = "mycollector"

  servers = [
    "http://my.service.com/_stats"
  ]

  # HTTP method to use (case-sensitive)
  method = "GET"
```

The name is used as a prefix for the measurements.

The `method` specifies HTTP method to use for requests.

You can specify which keys from server response should be considered as tags:

```
[[httpjson.services]]
  ...

  tag_keys = [
    "role",
    "version"
  ]
```

**NOTE**: tag values should be strings.

You can also specify additional request parameters for the service:

```
[[httpjson.services]]
  ...

 [httpjson.services.parameters]
    event_type = "cpu_spike"
    threshold = "0.75"

```


# Sample

Let's say that we have a service named "mycollector", which responds with:
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

The collected metrics will be:
```
httpjson_mycollector_a value=0.5
httpjson_mycollector_b_d value=0.1
httpjson_mycollector_b_e value=5
```
