# uWSGI

The uWSGI gathers metrics about uWSGI using its [Stats Server][stats_server].

### Configuration

```toml
[[inputs.uwsgi]]
    ## List with urls of uWSGI Stats servers. Url must match pattern:
    ## scheme://address[:port]
    ##
    ## For example:
    ## servers = ["tcp://localhost:5050", "http://localhost:1717", "unix:///tmp/statsock"]
    servers = []
```

### Measurements and fields

- uwsgi_overview
    - listen_queue
    - listen_queue_errors
    - load
    - signal_queue

- uwsgi_workers
    - requests
    - accepting
    - delta_request
    - exceptions
    - harakiri_count
    - signals
    - signal_queue
    - status
    - rss
    - vsz
    - running_time
    - last_spawn
    - respawn_count
    - tx
    - avg_rt

- uwsgi_apps
    - modifier1
    - requests
    - startup_time
    - exceptions

- uwsgi_cores
    - requests
    - static_requests
    - routed_requests
    - offloaded_requests
    - write_errors
    - read_errors
    - in_request

###  Tags

- uwsgi_overview
    - url
    - pid
    - uid
    - gid
    - version
    - cwd

- uwsgi_workers
    - worker_id
    - url
    - pid

- uwsgi_apps
    - app_id
    - worker_id
    - mount_point
    - chdir

- uwsgi_cores
    - core_id
    - worker_id

### Example Output:

```
* Plugin: uwsgi, Collection 1
> uwsgi_overview,cwd=/tmp/uwsgi,gid=1000,pid=5347,uid=1000,url=tcp://localhost:5050,version=2.0.12 listen_queue=0i,listen_queue_errors=0i,load=0i,signal_queue=0i 1460660786056661209
> uwsgi_workers,pid=5350,url=tcp://localhost:5050,worker_id=1 accepting=1i,avg_rt=0i,delta_request=0i,exceptions=0i,harakiri_count=0i,last_spawn=1460660781i,requests=0i,respawn_count=1i,rss=0i,running_time=0i,signal_queue=0i,signals=0i,status="idle",tx=0i,vsz=0i 1460660786056861579
> uwsgi_apps,app_id=0,worker_id=1 exceptions=0i,modifier1=0i,requests=0i,startup_time=0i 1460660786057015120
> uwsgi_cores,core_id=0,worker_id=1 in_request=0i,offloaded_requests=0i,read_errors=0i,requests=0i,routed_requests=0i,static_requests=0i,write_errors=0i 1460660786057186673
```

[stats_server]: http://uwsgi-docs.readthedocs.org/en/latest/StatsServer.html
