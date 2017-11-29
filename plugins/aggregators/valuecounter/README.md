# ValueCounter Aggregator Plugin

The valuecounter plugin counts the occurance of values in fields and emmits the
counter once every 'preriod' seconds.

A usecase for the valuecounter plugin is when you are processing a HTTP access
log (with the logparser input) and want to count the HTTP status codes.

The fields which will be counted must be configured with the `field_names`
configuration directive. When no `field_names` is provided the plugin will not
count any fields. The results are emitted in fields in the format:
`originalfieldname_fieldvalue = count`.

### Configuration:

```toml
[[aggregators.valuecounter]]
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false
  ## The fields for which the values will be counted
  field_names = ["status"]
```

### Measurements & Fields:

- measurement1
    - field_value1
    - field_value2

### Tags:

No tags are applied by this aggregator.

### Example Output:

Example for parsing a HTTP access log.

telegraf.conf:
```
[[inputs.logparser]]
  files = ["/tmp/tst.log"]
  [inputs.logparser.grok]
    patterns = ['%{DATA:url:tag} %{NUMBER:response:string}']
    measurement = "access"

[[aggregators.valuecounter]]
  namepass = ["access"]
  field_names = ["response"]
```

/tmp/tst.log
```
/some/path 200
/some/path 401
/some/path 200
```

```
$ telegraf --config telegraf.conf --quiet

access,url=/some/path,path=/tmp/tst.log,host=localhost.localdomain response="200" 1511948755991487011
access,url=/some/path,path=/tmp/tst.log,host=localhost.localdomain response="401" 1511948755991522282
access,url=/some/path,path=/tmp/tst.log,host=localhost.localdomain response="200" 1511948755991531697

access,path=/tmp/tst.log,host=localhost.localdomain,url=/some/path response_200=2i,response_401=1i 1511948761000000000
```
