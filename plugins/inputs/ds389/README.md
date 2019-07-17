# 389 Directory Server Input Plugin

This plugin gathers metrics from 389 Directory Servers's cn=Monitor backend and cn=monitor,cn=${database},cn=ldbm database,cn=plugins,cn=config for each indexes ans database files.

### Configuration:

```toml
[[inputs.ds389]]
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
  
  # reverse metric names so they sort more naturally
  # Defaults to false if unset, but is set to true when generating a new config
  dbtomonitor = ["db1","db2"]
```

### Measurements & Fields:

All attributes are gathered based on this LDAP query:

```(objectClass=*)```

Metric names are attributes name. 

If dbtomonitor array is provided , it can gather metrics for each dbfilename like uniquemember, memberof, givename indexes. 

A 389DS 1.3.7 server will provide these metrics:

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
- connections
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



### Tags:

- server= # value from config
- port= # value from config

### Example Output:

```
$ telegraf -config telegraf.conf -input-filter ds389 -test 
> ds389,host=dod-X750JB,port=389,server=ldap01 addentryops=0i,anonymousbinds=0i,bindsecurityerrors=3i,bytesrecv=0i,bytessent=190256225i,cacheentries=0i,cachehits=0i,chainings=0i,compareops=0i,connections=6i,connectionseq=86840i,connectionsinmaxthreads=0i,connectionsmaxthreadscount=0i,copyentries=0i,currentconnections=6i,currentconnectionsatmaxthreads=0i,dtablesize=1024i,entriesreturned=259120i,entriessent=259120i,errors=255i,inops=306715i,listops=0i,masterentries=0i,maxthreadsperconnhits=0i,modifyentryops=11i,modifyrdnops=0i,onelevelsearchops=118i,opscompleted=306714i,opsinitiated=306715i,readops=0i,readwaiters=0i,referrals=0i,referralsreturned=0i,removeentryops=0i,searchops=117848i,securityerrors=0i,simpleauthbinds=86815i,slavehits=0i,strongauthbinds=0i,totalconnections=86840i,unauthbinds=3i,wholesubtreesearchops=113152i 1554566915000000000
```
