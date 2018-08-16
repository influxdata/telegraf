# Parser Processor Plugin
This plugin parses defined fields containing the specified data format.

## Configuration
```toml
[[processors.parser]]
  ## specify the name of the field[s] whose value will be parsed
  parse_fields = ["message"]

  ## specify what to do with the original message. [merge|replace|keep] default=keep
  original = "merge"

  [processors.parser.config]
    data_format = "json"
    ## additional configurations for parser go here
    tag_keys = ["verb", "request"]
```

### Tags:

User specified tags may be added by this processor.

### Example Config:
```toml
[[inputs.exec]]
  data_format = "influx"
  commands = [
    "echo -en 'thing,host=\"mcfly\" message=\"{\\\"verb\\\":\\\"GET\\\",\\\"request\\\":\\\"/time/to/awesome\\\"}\" 1519652321000000000'"
  ]

[[processors.parser]]
  ## specify the name of the field[s] whose value will be parsed
  parse_fields = ["message"]

  ## specify what to do with the original message. [merge|replace|keep] default=keep
  original = "merge"

  [processors.parser.config]
    data_format = "json"
    ## additional configurations for parser go here
    tag_keys = ["verb", "request"]

[[outputs.file]]
  files = ["stdout"]
```

### Example Output [original=merge]:
```
# input = nginx_requests,host="mcfly" message="{\"verb\":\"GET\",\"request\":\"/time/to/awesome\"}" 1519652321000000000
nginx_requests,host="mcfly",verb="GET",request="/time/to/awesome" message="{\"verb\":\"GET\",\"request\":\"/time/to/awesome\"}" 1519652321000000000
```

### Caveats
While this plugin is functional, it may not work in *every* scenario. For the above example, "keep" and "replace" fail because the parsed field produces a metric with no fields. This leads to errors when serializing the output.
