# Converter Processor

The converter processor is used to change the type of tag or field values.  In
addition to changing field types it can convert between fields and tags.

Values that cannot be converted are dropped.

**Note:** When converting tags to fields, take care to ensure the series is still
uniquely identifiable.  Fields with the same series key (measurement + tags)
will overwrite one another.

### Configuration:
```toml
# Convert values to another metric value type
[[processors.converter]]
  ## Tags to convert
  ##
  ## The table key determines the target type, and the array of key-values
  ## select the keys to convert.  The array may contain globs.
  ##   <target-type> = [<tag-key>...]
  [processors.converter.tags]
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
    tag = []
    string = []
    integer = []
    unsigned = []
    boolean = []
    float = []
```

### Examples:

```toml
[[processors.converter]]
  [processors.converter.tags]
    string = ["port"]

  [processors.converter.fields]
    integer = ["scboard_*"]
    tag = ["ParseServerConfigGeneration"]
```

```diff
- apache,port=80,server=debian-stretch-apache BusyWorkers=1,BytesPerReq=0,BytesPerSec=0,CPUChildrenSystem=0,CPUChildrenUser=0,CPULoad=0.00995025,CPUSystem=0.01,CPUUser=0.01,ConnsAsyncClosing=0,ConnsAsyncKeepAlive=0,ConnsAsyncWriting=0,ConnsTotal=0,IdleWorkers=49,Load1=0.01,Load15=0,Load5=0,ParentServerConfigGeneration=3,ParentServerMPMGeneration=2,ReqPerSec=0.00497512,ServerUptimeSeconds=201,TotalAccesses=1,TotalkBytes=0,Uptime=201,scboard_closing=0,scboard_dnslookup=0,scboard_finishing=0,scboard_idle_cleanup=0,scboard_keepalive=0,scboard_logging=0,scboard_open=100,scboard_reading=0,scboard_sending=1,scboard_starting=0,scboard_waiting=49 1502489900000000000
+ apache,server=debian-stretch-apache,ParentServerConfigGeneration=3 port="80",BusyWorkers=1,BytesPerReq=0,BytesPerSec=0,CPUChildrenSystem=0,CPUChildrenUser=0,CPULoad=0.00995025,CPUSystem=0.01,CPUUser=0.01,ConnsAsyncClosing=0,ConnsAsyncKeepAlive=0,ConnsAsyncWriting=0,ConnsTotal=0,IdleWorkers=49,Load1=0.01,Load15=0,Load5=0,ParentServerMPMGeneration=2,ReqPerSec=0.00497512,ServerUptimeSeconds=201,TotalAccesses=1,TotalkBytes=0,Uptime=201,scboard_closing=0i,scboard_dnslookup=0i,scboard_finishing=0i,scboard_idle_cleanup=0i,scboard_keepalive=0i,scboard_logging=0i,scboard_open=100i,scboard_reading=0i,scboard_sending=1i,scboard_starting=0i,scboard_waiting=49i 1502489900000000000
```
