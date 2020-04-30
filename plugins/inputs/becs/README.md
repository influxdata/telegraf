# BECS Input Plugin

The [becs][becs] plugin uses BECS's API to gather metrics from Packetfront's BECS 

Tested BECS versions: 

3.17.1

3.18.0

### Configuration:
```toml
[[inputs.becs]]
  ## BECS server.
  ## Default = "localhost:4490".
  server = "localhost:4490"

  ## BECS login credentials.
  username = "becs"
  password = "becs"
  # namespace = ""

  ## Resources to collect clients from.
  ## Example = ["10.0.0.0/8"].
  # resources = []

  ## Include memory pools from applications.
  ## Default = false.
  include_pools = false
```

### Metrics

- becs_applications:
  - tags:
    - application
    - memorypool (optional)
    - server
  - fields:
    - cpuusage
    - cpuaverage60
    - emptypages (optional)
    - out (optional)
    - pages (optional)
    - size (optional)
    - uptime

- becs_metrics:
  - tags:
    - active
    - application
    - applicationid
    - cell
    - emtype
    - metric
    - server
  - fields:
    - elements

- becs_clients:
  - tags:
   - resource
   - server
  - fields:
    - clients

### Example Output:

```
becs_applications,application=EMIBOS-labb-primary,host=labbbecs,server=labbbecs uptime=13950i,cpuusage=0i,cpuaverage60=0i 1587031700000000000
becs_metrics,active=1,application=CRE,applicationid=CRE-labb-primary,cell=labb,emtype=ctshes,host=labbbecs,metric=em_elements,server=labbbecs elements=3889i 1588105280000000000
becs_clients,host=labbbecs,resource=10.0.0.0/8,server=labbbecs clients=3889i 1588268190000000000
```

[becs]: https://pfsw.com/becs/