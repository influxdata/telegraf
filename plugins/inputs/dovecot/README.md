# Dovecot Input Plugin

The dovecot plugin uses the dovecot Stats protocol to gather metrics on configured
domains. You can read Dovecot's documentation
[here](http://wiki2.dovecot.org/Statistics)


### Configuration:

```
# Read metrics about dovecot servers
[[inputs.dovecot]]
  ## specify dovecot servers via an address:port list
  ##  e.g.
  ##    localhost:24242
  ##
  ## If no servers are specified, then localhost is used as the host.
  servers = ["localhost:24242"]
  ## Type is one of "user", "domain", "ip", or "global"
  type = "global"
  ## Wildcard matches like "*.com". An empty string "" is same as "*"
  ## If type = "ip" filters should be <IP/network>
  filters = [""]
```


### Tags:
	server: hostname
	type: query type
	ip: ip addr
	user: username
	domain: domain name


### Fields:

	reset_timestamp        time.Time
	last_update            time.Time
	num_logins             int64
	num_cmds               int64
	num_connected_sessions int64				## not in <user> type
	user_cpu               float32
	sys_cpu                float32
	clock_time             float64
	min_faults             int64
	maj_faults             int64
	vol_cs                 int64
	invol_cs               int64
	disk_input             int64
	disk_output            int64
	read_count             int64
	read_bytes             int64
	write_count            int64
	write_bytes            int64
	mail_lookup_path       int64
	mail_lookup_attr       int64
	mail_read_count        int64
	mail_read_bytes        int64
	mail_cache_hits        int64


### Example Output:

```
telegraf --config t.cfg --input-filter dovecot --test
* Plugin: dovecot, Collection 1
> dovecot,ip=192.168.0.1,server=dovecot-1.domain.test,type=ip clock_time=0,disk_input=0i,disk_output=0i,invol_cs=0i,last_update="2016-04-08 10:59:47.000208479 +0200 CEST",mail_cache_hits=0i,mail_lookup_attr=0i,mail_lookup_path=0i,mail_read_bytes=0i,mail_read_count=0i,maj_faults=0i,min_faults=0i,num_cmds=12i,num_connected_sessions=0i,num_logins=6i,read_bytes=0i,read_count=0i,reset_timestamp="2016-04-08 10:33:34 +0200 CEST",sys_cpu=0,user_cpu=0,vol_cs=0i,write_bytes=0i,write_count=0i 1460106251633824223
* Plugin: dovecot, Collection 1
> dovecot,server=dovecot-1.domain.test,type=user,user=user-1@domain.test clock_time=0.00006,disk_input=405504i,disk_output=77824i,invol_cs=67i,last_update="2016-04-08 11:02:55.000111634 +0200 CEST",mail_cache_hits=26i,mail_lookup_attr=0i,mail_lookup_path=6i,mail_read_bytes=86233i,mail_read_count=5i,maj_faults=0i,min_faults=975i,num_cmds=41i,num_logins=3i,read_bytes=368833i,read_count=394i,reset_timestamp="2016-04-08 11:01:32 +0200 CEST",sys_cpu=0.008,user_cpu=0.004,vol_cs=323i,write_bytes=105086i,write_count=176i 1460106256637049167
* Plugin: dovecot, Collection 1
> dovecot,domain=domain.test,server=dovecot-1.domain.test,type=domain clock_time=100896189179847.7,disk_input=6467588263936i,disk_output=17933680439296i,invol_cs=1194808498i,last_update="2016-04-08 11:04:08.000377367 +0200 CEST",mail_cache_hits=46455781i,mail_lookup_attr=0i,mail_lookup_path=571490i,mail_read_bytes=79287033067i,mail_read_count=491243i,maj_faults=16992i,min_faults=1278442541i,num_cmds=606005i,num_connected_sessions=6597i,num_logins=166381i,read_bytes=30231409780721i,read_count=1624912080i,reset_timestamp="2016-04-08 10:28:45 +0200 CEST",sys_cpu=156440.372,user_cpu=216676.476,vol_cs=2749291157i,write_bytes=17097106707594i,write_count=944448998i 1460106261639672622
* Plugin: dovecot, Collection 1
> dovecot,server=dovecot-1.domain.test,type=global clock_time=101196971074203.94,disk_input=6493168218112i,disk_output=17978638815232i,invol_cs=1198855447i,last_update="2016-04-08 11:04:13.000379245 +0200 CEST",mail_cache_hits=68192209i,mail_lookup_attr=0i,mail_lookup_path=653861i,mail_read_bytes=86705151847i,mail_read_count=566125i,maj_faults=17208i,min_faults=1286179702i,num_cmds=917469i,num_connected_sessions=8896i,num_logins=174827i,read_bytes=30327690466186i,read_count=1772396430i,reset_timestamp="2016-04-08 10:28:45 +0200 CEST",sys_cpu=157965.692,user_cpu=219337.48,vol_cs=2827615787i,write_bytes=17150837661940i,write_count=992653220i 1460106266642153907
```
