# Telegraf Plugin: Beanstalkd

### Configuration:

```
# # Read metrics from one or many beanstalkd servers
[[inputs.beanstalkd]]
#   ## An array of address to gather stats about. Specify an ip on hostname
#   ## with optional port. ie localhost, 10.0.0.1:11300, etc.
   servers = ["localhost:11300"]
```

### Measurements & Fields:

- Measurement
    - current-jobs-urgent
    - current-jobs-ready
    - current-jobs-reserved
    - current-jobs-delayed
    - current-jobs-buried
    - cmd-put
    - cmd-peek
    - cmd-peek-ready
    - cmd-peek-delayed
    - cmd-peek-buried
    - cmd-reserve
    - cmd-reserve-with-timeout
    - cmd-delete
    - cmd-release
    - cmd-use
    - cmd-watch
    - cmd-ignore
    - cmd-bury
    - cmd-kick
    - cmd-touch
    - cmd-stats
    - cmd-stats-job
    - cmd-stats-tube
    - cmd-list-tubes
    - cmd-list-tube-used
    - cmd-list-tubes-watched
    - cmd-pause-tube
    - job-timeouts
    - total-jobs
    - current-tubes
    - current-connections
    - current-producers
    - current-workers
    - current-waiting
    - total-connections
    - uptime
    - binlog-oldest-index
    - binlog-current-index
    - binlog-records-migrated
    - binlog-records-written
    - binlog-max-size"

### Example Output:

Using this configuration:
```
# # Read metrics from one or many beanstalkd servers
[[inputs.beanstalkd]]
#   ## An array of address to gather stats about. Specify an ip on hostname
#   ## with optional port. ie localhost, 10.0.0.1:11300, etc.
   servers = ["localhost:11300"]
```

When run with:
```
./telegraf -config telegraf.conf -input-filter beanstalkd -test
```

It produces:
```
* Plugin: inputs.beanstalkd, Collection 1
> beanstalkd,host=irrlab,server=localhost:11300 binlog-current-index=10i,binlog-max-size=10485760i,binlog-oldest-index=10i,binlog-records-migrated=0i,binlog-records-written=1838i,cmd-bury=0i,cmd-delete=919i,cmd-ignore=1255i,cmd-kick=0i,cmd-list-tube-used=0i,cmd-list-tubes=0i,cmd-list-tubes-watched=0i,cmd-pause-tube=0i,cmd-peek=0i,cmd-peek-buried=0i,cmd-peek-delayed=0i,cmd-peek-ready=0i,cmd-put=919i,cmd-release=0i,cmd-reserve=1255i,cmd-reserve-with-timeout=0i,cmd-stats=341i,cmd-stats-job=0i,cmd-stats-tube=0i,cmd-touch=0i,cmd-use=919i,cmd-watch=1255i,current-connections=1i,current-jobs-buried=0i,current-jobs-delayed=0i,current-jobs-ready=0i,current-jobs-reserved=0i,current-jobs-urgent=0i,current-producers=0i,current-tubes=1i,current-waiting=0i,current-workers=0i,job-timeouts=0i,total-connections=1595i,total-jobs=919i,uptime=15736i 1476809911000000000
```
