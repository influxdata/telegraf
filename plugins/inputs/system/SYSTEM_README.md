# System Input Plugin

The system plugin gathers general stats on system load, uptime,
and number of users logged in. It is basically equivalent
to the unix `uptime` command.

### Configuration:

```toml
# Read metrics about system load & uptime
[[inputs.system]]
  # no configuration
```

### Measurements & Fields:

- system
    - load1 (float)
    - load15 (float)
    - load5 (float)
    - n_users (integer)
    - uptime (integer, seconds)
    - uptime_format (string)

### Tags:

None

### Example Output:

```
$ telegraf -config ~/ws/telegraf.conf -input-filter system -test
* Plugin: system, Collection 1
> system load1=2.05,load15=2.38,load5=2.03,n_users=4i,uptime=239043i,uptime_format="2 days, 18:24" 1457546165399253452
```
