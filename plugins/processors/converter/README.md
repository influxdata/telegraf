# Converter Processor

The converter processor is used to change the type of tag or field values.  In
addition to changing field types it can convert between fields and tags.

Values that cannot be converted are dropped.

**Note:** When converting tags to fields, take care to ensure the series is still
uniquely identifiable.  Fields with the same series key (measurement + tags)
will overwrite one another.

### Configuration
```toml
# Convert values to another metric value type
[[processors.converter]]
  ## Tags to convert
  ##
  ## The table key determines the target type, and the array of key-values
  ## select the keys to convert.  The array may contain globs.
  ##   <target-type> = [<tag-key>...]
  [processors.converter.tags]
    measurement = []
    string = []
    integer = []
    unsigned = []
    boolean = []
    float = []

  ## Fields to convert
  ##
  ## The table key determines the target type, and the array of key-values
  ## select the keys to convert.  The array may contain globs.
  ##   <target-type> = [<field-key>...]
  [processors.converter.fields]
    measurement = []
    tag = []
    string = []
    integer = []
    unsigned = []
    boolean = []
    float = []
```

### Example

Convert `port` tag to a string field:
```toml
[[processors.converter]]
  [processors.converter.tags]
    string = ["port"]
```

```diff
- apache,port=80,server=debian-stretch-apache BusyWorkers=1,BytesPerReq=0
+ apache,server=debian-stretch-apache port="80",BusyWorkers=1,BytesPerReq=0
```

Convert all `scboard_*` fields to an integer:
```toml
[[processors.converter]]
  [processors.converter.fields]
    integer = ["scboard_*"]
```

```diff
- apache scboard_closing=0,scboard_dnslookup=0,scboard_finishing=0,scboard_idle_cleanup=0,scboard_keepalive=0,scboard_logging=0,scboard_open=100,scboard_reading=0,scboard_sending=1,scboard_starting=0,scboard_waiting=49
+ apache scboard_closing=0i,scboard_dnslookup=0i,scboard_finishing=0i,scboard_idle_cleanup=0i,scboard_keepalive=0i,scboard_logging=0i,scboard_open=100i,scboard_reading=0i,scboard_sending=1i,scboard_starting=0i,scboard_waiting=49i
```

Rename the measurement from a tag value:
```toml
[[processors.converter]]
  [processors.converter.tags]
    measurement = ["topic"]
```

```diff
- mqtt_consumer,topic=sensor temp=42
+ sensor temp=42
```
