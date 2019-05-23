# Docker application HTTP JSON Input Plugin

The containerapp plugin collects data from container HTTP URL which respond with JSON.
It flattens the JSON and finds all numeric values, treating them as floats.
Settings are taken from the environment variables of the container.


### Configuration:

```toml
[[inputs.containerapp]]
  ## NOTE This plugin only reads numerical measurements, strings and booleans
  ## will be ignored.

  ## Config source type
  ##   Types: docker_env, kubernetes_label
  config_type = "docker_env"

  ## Interval with which new docker container collectors start(default 10ms)
  start_interval = "10ms"

  ## Rescan containers list interval(if there is a heavy load on the server, not always the messages come through the event api)
  sync_interval = "10m"

  ## Which config source names should be use as tag
  tags_name = ["MON_INTERVAL", "MON_PATH"]

  ## Prefix of config source names should be use as tag
  tags_prefix = "MON_TAG_"

  ## Mandatory config values, skip if it does not exist
  tags_mandatory = ["mon.db"]

  # docker_env conf
  [inputs.containerapp.docker_env]
     ## Docker Endpoint
     ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
     ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
     endpoint = "unix:///var/run/docker.sock"

  # kubernetes conf
  ## if this section declared plugin will search labels or annotations in k8s pods
  [inputs.containerapp.kubernetes]
     # kubernetes nodename, to be served by this instance 
     nodename = "testnode" 
     # full path to kubernetes config file
     kubeconfig = "/home/user/.kube/config"

  ## Mapping config source -> http
  [inputs.containerapp.http]
     ## Interval with which docker container application metrics gather
     interval  = "MON_INTERVAL"

     name_override  = "MON_NAME_OVERRIDE"

     ## Variable in which additional tags for the container are stored.
     ## Tags for different containers can be very different, therefore some of them 
     ## are stored in a json dump and can be simply populated by external container 
     ## deployment systems
     ##   Example value: MON_CUSTOM_TAGS={"my_tag_1":"my_tag_1"}
     ##   (stored in an config source variable as json dump)
     custom_tags = "MON_CUSTOM_TAGS"

     ## Metrics URL port
     ##   Example value: MON_PORT=8888
     http_port  = "MON_PORT"
	 
     ## Metrics URL path
     ##   Example value: MON_PATH=/mon/
     http_path  = "MON_PATH"

     ## Set response_timeout
     http_response_timeout  = "MON_RESPONSE_TIMEOUT"

     ## HTTP method to use: GET or POST (case-sensitive)
     http_method  = "MON_METHOD"
  
     ## HTTP parameters (all values must be strings).  For "GET" requests, data
     ## will be included in the query.  For "POST" requests, data will be included
     ## in the request body as "x-www-form-urlencoded".
     ##   Example value: MON_PARAMETERS={"my_parameter": "my_parameter"}
     ##   (stored in an config source variable as json dump)
     http_parameters   = "MON_PARAMETERS"
  
     ## HTTP Headers (all values must be strings)
     ##   Example value: MON_HEADERS={"my_header": "my_header"}  
     ##   (stored in an config source variable as json dump)
     http_headers  = "MON_HEADERS"

     ## List of tag names to extract from top-level of JSON server response
     tag_keys_json  = "MON_TAG_KEYS"

  ## HTTP default configuration
  [inputs.containerapp.http_defaults]
     ## Examples:
     interval  = "10s"
     name_override  = "test"
     http_port  = "8080"
     http_path  = "metrics"
     http_response_timeout = "5s"
     http_method  = "metrics"
```

### Measurements & Fields:

- containerapp
	- response_time (float): Response time in seconds

Additional fields are dependant on the response of the remote service being polled.

### Tags:

- All measurements have the following tags:
	- server: HTTP origin as defined in configuration as `http_port`, `http_path` and automatically calculated container ip address.
    - containerid: docker container ID

Any top level keys listed under `tag_keys_json` in the configuration are added as tags.  Top level keys are defined as keys in the root level of the object in a single object response, or in the root level of each object within an array of objects.


### Examples Output:

This plugin understands responses containing a single JSON object, or a JSON Array of Objects.

**Object Output:**

Given the following response body:
```json
{
    "a": 0.5,
    "b": {
        "c": "some text",
        "d": 0.1,
        "e": 5
    },
    "service": "service01"
}
```
The following metric is produced:

`httpjson,server=http://172.18.0.2:9999/stats/ b_d=0.1,a=0.5,b_e=5,response_time=0.001`

Note that only numerical values are extracted and the type is float.

If `tag_keys_json` is included in the configuration:

```toml
[[inputs.containerapp]]
 [inputs.containerapp.http]
  ##   MON_TAG_KEYS=["service"] 
  tag_keys_json  = "MON_TAG_KEYS"
```

Then the `service` tag will also be added:

`httpjson,server=http://172.18.0.2:9999/stats/,service=service01 b_d=0.1,a=0.5,b_e=5,response_time=0.001`

**Array Output:**

If the service returns an array of objects, one metric is be created for each object:

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

`httpjson,server=http://localhost:9999/stats/,service=service01 a=0.5,b_d=0.1,b_e=5,response_time=0.003`
`httpjson,server=http://localhost:9999/stats/,service=service02 a=0.6,b_d=0.2,b_e=6,response_time=0.003`
