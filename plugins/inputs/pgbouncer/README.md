# Pgbouncer plugin

This pgbouncer plugin provides metrics for your pgbouncer connection information.

### Configuration:

```toml
# Description
[[inputs.pgbouncer]]
  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost:port[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqotest port=... password=... sslmode=... dbname=...
  ##
  ## All connection parameters are optional, except for dbname,
  ## you need to set it always as pgbouncer.
  address = "host=localhost user=postgres port=6432 sslmode=disable dbname=pgbouncer"

  ## A list of databases to pull metrics about. If not specified, metrics for all
  ## databases are gathered.
  # databases = ["app_production", "testing"]
`
```

### Measurements & Fields:

Pgbouncer provides two measurement named "pgbouncer_pools" and "pgbouncer_stats", each have the fields as below:

#### pgbouncer_pools

- cl_active
- cl_waiting
- maxwait
- pool_mode
- sv_active
- sv_idle
- sv_login
- sv_tested
- sv_used

### pgbouncer_stats

- avg_query
- avg_recv
- avg_req
- avg_sent
- total_query_time
- total_received
- total_requests
- total_sent

More information about the meaning of these metrics can be found in the [PgBouncer usage](https://pgbouncer.github.io/usage.html)

### Example Output:

```
$ ./telegraf -config telegraf.conf -input-filter pgbouncer -test
> pgbouncer_pools,db=pgbouncer,host=localhost,pool_mode=transaction,server=host\=localhost\ user\=elena\ port\=6432\ dbname\=pgbouncer\ sslmode\=disable,user=elena cl_active=1500i,cl_waiting=0i,maxwait=0i,sv_active=0i,sv_idle=5i,sv_login=0i,sv_tested=0i,sv_used=5i 1466594520564518897
> pgbouncer_stats,db=pgbouncer,host=localhost,server=host\=localhost\ user\=elena\ port\=6432\ dbname\=pgbouncer\ sslmode\=disable avg_query=1157i,avg_recv=36727i,avg_req=131i,avg_sent=23359i,total_query_time=252173878876i,total_received=55956189078i,total_requests=193601888i,total_sent=36703848280i 1466594520564825345
```

