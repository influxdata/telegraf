# pidstat Input Plugin

Collect [pidstat](https://github.com/sysstat/sysstat) metrics - requires the
pidstat package installed.

This plugin collects per process system metrics with the pidstat utility. It
provides resource usage data 1) per process id 2) summed over command names or
both

### Configuration:
```toml
#Pidstat metrics collector
[[inputs.pidstat]]
  ## Gather metrics per pid or per command name
  ## metrics pidstat_pid and pidstat_command respectively
  # per_pid = true
  # per_command = true
  #
  ## Which command names to track. None = all
  # programs = []
  ## programs = [ "ngnix", "sh" ]
```

### Metrics:
- pidstat_pid
	- fields:

		# memory
		- minflt_per_s: (float64)
		- majflt_per_s: (float64)
		- RSS: (float64)
		- VSZ: (float64)
		- pct_MEM: (float64)

		# stack 
		- StkRef: (float64),
		- StkSize: (float64),

		# io
		- kB_rd_per_s: (float64),
		- kB_wr_per_s: (float64),
		- kB_ccwr_per_s: (float64),
		- iodelay: (float64),


		# threading
		- threads: (float64)
		- fd-nr: (float64)

		# context switching
		- cswch_per_s: (float64)
		- nvcswch_per_s: (float64)

		# processor
		- pct_usr: (float64)
		- pct_system: (float64)
		- pct_guest: (float64)
		- pct_CPU: (float64)
		- CPU: (float64)
	- tags:
		"arch"
		"cores"
		"UID"
		"PID" # pidstat_pid only
		"os"
		"os_ver"
		"sys_name"
		"command" # command name

Another metric - pidstat_command is the same except for "PID" tag and represents
the values summed over command names.

for meanings of field and value names please refer to pidstat documentation


With the configuration below:
```toml
# Pidstat metrics collector
[[inputs.pidstat]]
  ## Gather metrics per pid or per command name
  ## metrics pidstat_pid and pidstat_command respectively
   #per_pid = true
   per_command = true 
  #
  ## Which command names to track. None = all
   programs = ["sys*"]
```

you get the following output:
```
./telegraf --input-filter pidstat --test
2017/10/23 13:27:55 I! Using config file: /etc/telegraf/telegraf.conf
* Plugin: inputs.pidstat, Collection 1
> pidstat_command,UID=0,host=tyler-GL753VD,os=Linux,os_ver=4.10.0-37-generic,sys_name=(tyler-GL753VD),Command=systemd,arch=_x86_64_,cores=(8\
> CPU)
> pct_usr=0,StkSize=132,kB_rd_per_s=-1,RSS=10812,cswch_per_s=0.04,threads=1,pct_system=0,pct_guest=0,nvcswch_per_s=0.01,kB_ccwr_per_s=-1,iodelay=14,minflt_per_s=0.25,pct_MEM=0.14,majflt_per_s=0,fd-nr=15,kB_wr_per_s=-1,VSZ=222840,CPU=5,pct_CPU=0,StkRef=16
> 1508758076000000000
> pidstat_command,os_ver=4.10.0-37-generic,sys_name=(tyler-GL753VD),host=tyler-GL753VD,UID=0,Command=systemd-journal,arch=_x86_64_,cores=(8\
> CPU),os=Linux
> pct_guest=0,VSZ=35492,pct_MEM=0.08,pct_system=0,CPU=7,nvcswch_per_s=0,kB_rd_per_s=-1,kB_wr_per_s=-1,RSS=6092,cswch_per_s=0.05,iodelay=3,kB_ccwr_per_s=-1,majflt_per_s=0,minflt_per_s=0.02,pct_usr=0,pct_CPU=0
> 1508758076000000000
> pidstat_command,arch=_x86_64_,UID=0,Command=systemd-udevd,cores=(8\
> CPU),os=Linux,os_ver=4.10.0-37-generic,sys_name=(tyler-GL753VD),host=tyler-GL753VD
> RSS=4492,minflt_per_s=0.53,pct_usr=0,pct_CPU=0,pct_system=0,kB_rd_per_s=-1,iodelay=1,VSZ=45752,CPU=5,pct_guest=0,cswch_per_s=0.05,nvcswch_per_s=0.01,kB_wr_per_s=-1,kB_ccwr_per_s=-1,majflt_per_s=0,pct_MEM=0.06
> 1508758076000000000
> pidstat_command,sys_name=(tyler-GL753VD),UID=1000,host=tyler-GL753VD,Command=syndaemon,arch=_x86_64_,cores=(8\
> CPU),os=Linux,os_ver=4.10.0-37-generic
> minflt_per_s=0,majflt_per_s=0,pct_guest=0,fd-nr=5,VSZ=22504,RSS=1244,pct_CPU=0.01,CPU=2,nvcswch_per_s=0.01,cswch_per_s=1.6,iodelay=0,kB_wr_per_s=0,kB_ccwr_per_s=0,pct_MEM=0.02,pct_usr=0,pct_system=0,kB_rd_per_s=0,threads=1,StkSize=132,StkRef=12
> 1508758076000000000
> pidstat_command,os_ver=4.10.0-37-generic,cores=(8\
> CPU),host=tyler-GL753VD,sys_name=(tyler-GL753VD),arch=_x86_64_,UID=100,Command=systemd-timesyn,os=Linux
> VSZ=102464,CPU=3,pct_CPU=0,cswch_per_s=0.01,minflt_per_s=0,majflt_per_s=0,pct_system=0,pct_usr=0,pct_guest=0,nvcswch_per_s=0,RSS=2512,pct_MEM=0.03
> 1508758076000000000
> pidstat_command,UID=0,os=Linux,arch=_x86_64_,host=tyler-GL753VD,Command=systemd-logind,os_ver=4.10.0-37-generic,sys_name=(tyler-GL753VD),cores=(8\
> CPU)
> pct_MEM=0.04,majflt_per_s=0,pct_system=0,pct_guest=0,VSZ=20408,RSS=2836,minflt_per_s=0,pct_usr=0,pct_CPU=0,CPU=0,cswch_per_s=0.03,nvcswch_per_s=0
> 1508758076000000000
> pidstat_command,UID=104,os_ver=4.10.0-37-generic,cores=(8\
> CPU),Command=rsyslogd,host=tyler-GL753VD,os=Linux,sys_name=(tyler-GL753VD),arch=_x86_64_
> minflt_per_s=0.01,majflt_per_s=0,CPU=6,pct_usr=0,nvcswch_per_s=0,cswch_per_s=0,pct_MEM=0.04,VSZ=262776,RSS=3384,pct_CPU=0,pct_system=0,pct_guest=0
> 1508758076000000000
```
