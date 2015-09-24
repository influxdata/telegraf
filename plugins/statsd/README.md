# Telegraf Service Plugin: statsd

#### Plugin arguments:

- **service_address** string: Address to listen for statsd UDP packets on
- **delete_gauges** boolean: Delete gauges on every collection interval
- **delete_counters** boolean: Delete counters on every collection interval
- **delete_sets** boolean: Delete set counters on every collection interval
- **allowed_pending_messages** integer: Number of messages allowed to queue up
on the UDP listener before the next flush. NOTE: gauge, counter, and set
measurements are aggregated as they arrive, so this is not a straight counter of
the number of total messages that the listener can handle between flushes.

#### Statsd bucket -> InfluxDB Mapping

By default, statsd buckets are converted to measurement names with the rules:
- "." -> "_"
- "-" -> "__"

This plugin also accepts a list of config tables to describe a mapping of a statsd
bucket to an InfluxDB measurement name and tags.

Each mapping must specify a match glob pattern. It can optionally take a name
for the measurement and a map of bucket indices to tag names.

For example, the following configuration:

```
    [[statsd.mappings]]
    match = "users.current.*.*"
    name = "current_users"
    [statsd.mappings.tagmap]
    unit = 0
    server = 2
    service = 3

    [[statsd.mappings]]
    match = "deploys.*.*"
    name = "service_deploys"
    [statsd.mappings.tagmap]
    service_type = 1
    service_name = 2
```

Will map statsd -> influx like so:
```
users.current.den001.myapp:32|g
=> [server="den001" service="myapp" unit="users"] statsd_current_users_gauge value=32

deploys.test.myservice:1|c
=> [service_name="myservice" service_type="test"] statsd_service_deploys_counter value=1

random.jumping-sheep:10|c
=> [] statsd_random_jumping__sheep_counter value=10
```

#### Description

The statsd plugin is a special type of plugin which runs a backgrounded statsd
listener service while telegraf is running.

The format of the statsd messages was based on the format described in the
original [etsy statsd](https://github.com/etsy/statsd/blob/master/docs/metric_types.md)
implementation. In short, the telegraf statsd listener will accept:

- Gauges
    - `users.current.den001.myapp:32|g` <- standard
    - `users.current.den001.myapp:+10|g` <- additive
    - `users.current.den001.myapp:-10|g`
- Counters
    - `deploys.test.myservice:1|c` <- increments by 1
    - `deploys.test.myservice:101|c` <- increments by 101
    - `deploys.test.myservice:1|c|@0.1` <- sample rate, increments by 10
- Sets
    - `users.unique:101|s`
    - `users.unique:101|s`
    - `users.unique:102|s` <- would result in a count of 2 for `users.unique`
- Timers
    - TODO
