# Dovecot Input Plugin

The dovecot plugin uses the dovecot Stats protocol to gather metrics on configured
domains. You can read Dovecot's documentation
[here](http://wiki2.dovecot.org/Statistics)


### Configuration:

```
# Read metrics about dovecot servers
[[inputs.dovecot]]
  # Dovecot servers
  #  specify dovecot servers via an address:port list
  #  e.g.
  #    localhost:24242
  #
  # If no servers are specified, then localhost is used as the host.
  servers = ["localhost:24242"]
  # Only collect metrics for these domains, collect all if empty
  domains = []
```


### Tags:
	server: hostname
	domain: domain name


### Fields:

	reset_timestamp        time.Time
	last_update            time.Time
	num_logins             int64
	num_cmds               int64
	num_connected_sessions int64
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
telegraf -config telegraf.cfg -input-filter dovecot -test
* Plugin: dovecot, Collection 1
> dovecot,domain=xxxxx.it,server=dovecot--1.mail.sys clock_time=12105746411632.5,disk_input=115285225472i,disk_output=4885067755520i,invol_cs=169701886i,last_update="2016-02-09 08:49:47.000014113 +0100 CET",mail_cache_hits=441828i,mail_lookup_attr=0i,mail_lookup_path=25323i,mail_read_bytes=241188145i,mail_read_count=11719i,maj_faults=3168i,min_faults=321438988i,num_cmds=51635i,num_connected_sessions=2i,num_logins=17149i,read_bytes=7939026951110i,read_count=3716991752i,reset_timestamp="2016-01-28 09:34:36 +0100 CET",sys_cpu=222595.288,user_cpu=267468.08,vol_cs=3288715920i,write_bytes=4483648967059i,write_count=1640646952i 1455004219924838345
> dovecot,domain=yyyyy.com,server=dovecot-1.mail.sys clock_time=6650794455331782,disk_input=61957695569920i,disk_output=2638244004487168i,invol_cs=2004805041i,last_update="2016-02-09 08:49:49.000251296 +0100 CET",mail_cache_hits=2499112513i,mail_lookup_attr=506730i,mail_lookup_path=39128227i,mail_read_bytes=1076496874501i,mail_read_count=32615262i,maj_faults=1643304i,min_faults=4216116325i,num_cmds=85785559i,num_connected_sessions=1177i,num_logins=11658255i,read_bytes=4289150974554145i,read_count=1112000703i,reset_timestamp="2016-01-28 09:31:26 +0100 CET",sys_cpu=121125923.032,user_cpu=145561336.428,vol_cs=205451885i,write_bytes=2420130526835796i,write_count=2991367252i 1455004219925152529
> dovecot,domain=xxxxx.it,server=dovecot-2.mail.sys clock_time=10710826586999.143,disk_input=79792410624i,disk_output=4496066158592i,invol_cs=150426876i,last_update="2016-02-09 08:48:19.000209134 +0100 CET",mail_cache_hits=5480869i,mail_lookup_attr=0i,mail_lookup_path=122563i,mail_read_bytes=340746273i,mail_read_count=44275i,maj_faults=1722i,min_faults=288071875i,num_cmds=50098i,num_connected_sessions=0i,num_logins=16389i,read_bytes=7259551999517i,read_count=3396625369i,reset_timestamp="2016-01-28 09:31:29 +0100 CET",sys_cpu=200762.792,user_cpu=242477.664,vol_cs=2996657358i,write_bytes=4133381575263i,write_count=1497242759i 1455004219924888283
> dovecot,domain=yyyyy.com,server=dovecot-2.mail.sys clock_time=6522131245483702,disk_input=48259150004224i,disk_output=2754333359087616i,invol_cs=2294595260i,last_update="2016-02-09 08:49:49.000251919 +0100 CET",mail_cache_hits=2139113611i,mail_lookup_attr=520276i,mail_lookup_path=37940318i,mail_read_bytes=1088002215022i,mail_read_count=31350271i,maj_faults=994420i,min_faults=1486260543i,num_cmds=40414997i,num_connected_sessions=978i,num_logins=11259672i,read_bytes=4445546612487315i,read_count=1763534543i,reset_timestamp="2016-01-28 09:31:24 +0100 CET",sys_cpu=123655962.668,user_cpu=149259327.032,vol_cs=4215130546i,write_bytes=2531186030222761i,write_count=2186579650i 1455004219925398372
```

