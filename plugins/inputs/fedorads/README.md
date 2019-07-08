# fedorads
## Input plugin to monitor your 389 (Red Hat) Directory Server

This plugin gathers metrics from 389 Directory Server's cn=Monitor backend. It is based on work made for Openldap Input Plugin.
### Configuration:
```
# 389 Directory Server cn=Monitor plugin
[[inputs.fedorads]]
  host = "localhost"
  port = 389

  # ldaps, starttls, or no encryption. default is an empty string, disabling all encryption.
  # note that port will likely need to be changed to 636 for ldaps
  # valid options: "" | "starttls" | "ldaps"
  tls = ""

  # skip peer certificate verification. Default is false.
  insecure_skip_verify = false

  # Path to PEM-encoded Root certificate to use to verify server certificate
  tls_ca = "/etc/ssl/certs.pem"

  # dn/password to bind with. If bind_dn is empty, an anonymous bind is performed.
  bind_dn = ""
  bind_password = ""

  # Connections status
  status = false
```
If you enable the Connection status (`status = true`) a full connection status detail will be added to the metrics. The idea is to monitor all metrics provided by the 389 Console.

### Measurements & Fields:
All **cn=monitor** attributes are gathered based on this LDAP query:
`(objectClass=extensibleObject)`

A Red Hat Directory Server 10.x will provide these metrics:
  - threads
  - connection (multivalue, ignored if `status = false` in your conf)
  - currentconnections
  - totalconnections
  - currentconnectionsatmaxthreads
  - maxthreadsperconnhits
  - dtablesize
  - readwaiters
  - opsinitiated
  - opscompleted
  - entriessent
  - bytessent
  - currenttime
  - starttime
  - nbackends
  - anonymousbinds
  - unauthbinds
  - simpleauthbinds
  - strongauthbinds
  - bindsecurityerrors
  - inops
  - readops
  - compareops
  - addentryops
  - removeentryops
  - modifyentryops
  - modifyrdnops
  - listops
  - searchops
  - onelevelsearchops
  - wholesubtreesearchops
  - referrals
  - chainings
  - securityerrors
  - errors
  - connections __(the same of currentconnections)__
  - connectionseq
  - connectionsinmaxthreads
  - connectionsmaxthreadscount
  - bytesrecv
  - bytessent
  - entriesreturned
  - referralsreturned
  - masterentries
  - copyentries
  - cacheentries
  - cachehits
  - slavehits

#### Connection status measurement
If `status = true` the connection attribute will be parsed in this metric:
- the metric prefix is "conn."
- the file dscriptor is then added to the metric name
- the name of the monitored parameter is then concatenated with a dot

Format:

`conn.<fd>.<param> = <value>`

### Tags:

    server= # value from config
    port= # value from config
    version= # value from cn=monitor version attribute
    
### Example Output:

```
$ telegraf -config etc/telegraf.conf --config-directory etc/ -input-filter fedorads -test -debug
2019-07-05T10:09:47Z I! Starting Telegraf
> fedorads,host=telegraf.example.com,port=489,server=ldap.example.com,version=389-Directory/1.3.8.4\ B2019.037.1535 addentryops=9i,anonymousbinds=1094i,bindsecurityerrors=82i,bytesrecv=0i,bytessent=646535300i,cacheentries=0i,cachehits=0i,chainings=0i,compareops=0i,conn.64.binddn="NULLDN",conn.64.ip="10.10.15.17",conn.64.opentime="20190704223412Z",conn.64.opscompleted=1i,conn.64.opsinitiated=1i,conn.64.rw="-",conn.65.binddn="NULLDN",conn.65.ip="10.10.15.17",conn.65.opentime="20190704223412Z",conn.65.opscompleted=1i,conn.65.opsinitiated=1i,conn.65.rw="-",conn.66.binddn="NULLDN",conn.66.ip="10.10.15.17",conn.66.opentime="20190704223416Z",conn.66.opscompleted=1i,conn.66.opsinitiated=1i,conn.66.rw="-",connections=25i,connectionseq=55854i,connectionsinmaxthreads=0i,connectionsmaxthreadscount=0i,copyentries=0i,currentconnections=25i,currentconnectionsatmaxthreads=0i,dtablesize=512i,entriesreturned=1372697i,entriessent=1372696i,errors=307i,inops=92043i,listops=0i,masterentries=0i,maxthreadsperconnhits=0i,modifyentryops=1045i,modifyrdnops=8i,nbackends=3i,onelevelsearchops=586i,opscompleted=92042i,opsinitiated=92043i,readops=0i,readwaiters=0i,referrals=0i,referralsreturned=0i,removeentryops=12i,searchops=19822i,securityerrors=83i,simpleauthbinds=5217i,slavehits=0i,strongauthbinds=0i,threads=24i,totalconnections=55854i,unauthbinds=1094i,wholesubtreesearchops=14463i 1562321388000000000
```

### To do:
- monitor replication?
- monitor all DB backends performance, as 389 Console does.

It would provide a new ton of metrics, with big resource requirements to manage all data.
