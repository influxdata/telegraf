# Cosmic Input Plugin

This plugin will collect statistics about virtualmachines, volumes and public ip adresses available in Cosmic.
It uses [Cosmic API](https://apidoc.cosmiccloud.io/) to query the data.

See [Cosmic](https://github.com/MissionCriticalCloud/cosmic/) for more information on the infrastructure orchestrator.

**Series Cardinality Warning**

Depending on the work load of your DC/OS cluster, this plugin can quickly
create a high number of series which, when unchecked, can cause high load on
your database.

- Use the
  [measurement filtering](https://docs.influxdata.com/telegraf/latest/administration/configuration/#measurement-filtering)
  options to exclude unneeded tags.
- Write to a database with an appropriate
  [retention policy](https://docs.influxdata.com/influxdb/latest/guides/downsampling_and_retention/).
- Limit series cardinality in your database using the
  [`max-series-per-database`](https://docs.influxdata.com/influxdb/latest/administration/config/#max-series-per-database-1000000) and
  [`max-values-per-tag`](https://docs.influxdata.com/influxdb/latest/administration/config/#max-values-per-tag-100000) settings.
- Consider using the
  [Time Series Index](https://docs.influxdata.com/influxdb/latest/concepts/time-series-index/).
- Monitor your databases
  [series cardinality](https://docs.influxdata.com/influxdb/latest/query_language/spec/#show-cardinality).


### Configuration:

```toml
# Gather metrics from virtualmachines available in Cosmic
[[inputs.cosmic]]
  ##
  ## Connection parameters
  ##

  ## Cosmic API endpoint
  url = "https://localhost/client/api"
  ## The API key to use for metrics collection
  apikey = "xxx"
  ## The corresponding secret key
  secretkey = "xxx"
  ## Timeout in seconds per http request to the Cosmic API endpoint
  timeout = 60

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ##
  ## Metric collection parameters
  ##

  ## The domain for which to collect metrics
  # domainid = "00000000-0000-0000-0000-000000000000"
```

### Metrics:

- cosmic_virtualmachine_metrics
  - tags:
    - id
    - name
    - account
    - created
    - displayname
    - domain
    - domainid
    - hypervisor
    - instancename
    - templatenid
    - templatename
    - templatedisplaytext
    - userid
    - username
    - zoneid
    - zonename
  - fields:
    - cpunumber
    - memory
    - state
    - hostid
    - hostname
    - serviceofferingid
    - serviceofferingname
- cosmic_volume_metrics
  - tags:
    - account
    - created
    - domain
    - domainid
    - host
    - hypervisor
    - id
    - name
    - zoneid
    - zonename
  - fields:
    - attached
    - destroyed
    - deviceid
    - diskofferingdisplaytext
    - diskofferingid
    - diskofferingname
    - path
    - size
    - state
    - storage
    - storageid
    - virtualmachineid
    - vmdisplayname
    - vmname
    - vmstate
- cosmic_publicipaddress_metrics
  - tags:
    - account
    - domain
    - domainid
    - host
    - id
    - ipaddress
    - zoneid
    - zonename
  - fields:
    - aclid
    - state
    - vpcid

### Example output

```
cosmic_virtualmachine_metrics,account=testaccount,created=2018-01-01T00:00:00+0000,displayname=testvm,domain=testdomain,domainid=testdomain,host=localhost,hypervisor=KVM,id=00000000-0000-0000-0000-000000000002,instancename=i-1-12345-VM,name=testvm,templatedisplaytext=testtemplatetext,templatename=testtemplatename,templatenid=00000000-0000-0000-0000-000000000004,userid=00000000-0000-0000-0000-000000000005,username=testuser,zoneid=00000000-0000-0000-0000-000000000006,zonename=testzone cpunumber=1i,hostid="",hostname="",memory=1024i,serviceofferingid="00000000-0000-0000-0000-000000000003",serviceofferingname="testserviceofferingname",state="Stopped" 1523046473000000000
```
```
cosmic_volume_metrics,account=testaccount,created=2018-01-01T00:00:00+0000,domain=testdomain,domainid=00000000-0000-0000-0000-000000000001,host=localhost,hypervisor=KVM,id=00000000-0000-0000-0000-000000000002,name=testname,zoneid=00000000-0000-0000-0000-000000000003,zonename=testzone attached="",destroyed=false,deviceid=0i,diskofferingdisplaytext="testdiskofferingtext",diskofferingid="00000000-0000-0000-0000-000000000004",diskofferingname="testdiskofferingname",path="00000000-0000-0000-0000-000000000005",size=42949672960i,state="Ready",storage="teststorage",storageid="00000000-0000-0000-0000-000000000006",virtualmachineid="00000000-0000-0000-0000-000000000007",vmdisplayname="testvmdisplayname",vmname="testvmname",vmstate="Running" 1523365319000000000
```
```
cosmic_publicipaddress_metrics,account=testaccount,domain=testdomain,domainid=00000000-0000-0000-0000-000000000001,host=localhost,id=00000000-0000-0000-0000-000000000002,ipaddress=1.2..3.4,zoneid=00000000-0000-0000-0000-000000000003,zonename=testzone aclid="00000000-0000-0000-0000-000000000004",state="Allocated",vpcid="00000000-0000-0000-0000-000000000005" 1523365325000000000
```
