# Windows Eventlog Input Plugin

Collects Windows events.

Supports Windows Vista and higher.

### Configuration:

```toml
  eventlog_name = "Application"
  xpath_query = "Event/System[EventID=999]"
```

Alternatively it is possible to use special XML query documented [here](https://docs.microsoft.com/en-us/windows/win32/wes/consuming-events), in this case eventlog_name field will be ignored (use triple quote for xml in telegraf config file).

### Measurements & Fields:

- win_eventlog
    - record_id : integer
    - event_id : integer
    - description : string
    - created : string
    - source : string

The `level` tag can have the following values:
- 1 - critical
- 2 - error
- 3 - warning
- 4 - information

### Tags:

- All measurements have the following tags:
    - level
    - eventlog_name
    
### Example Output:
```
win_eventlog,eventlog_name=Application,host=MYHOSTNAME,level=2 description="TEST777",source="Service777",created="2020-02-11 13:07:45.748337 +0000 UTC",record_id=58267i,event_id=999i 1581426470000000000
```