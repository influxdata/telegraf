gopsutil: psutil for golang
==============================

.. image:: https://drone.io/github.com/shirou/gopsutil/status.png
        :target: https://drone.io/github.com/shirou/gopsutil

.. image:: https://coveralls.io/repos/shirou/gopsutil/badge.png?branch=master
        :target: https://coveralls.io/r/shirou/gopsutil?branch=master


This is a port of psutil (http://pythonhosted.org/psutil/). The challenge is porting all 
psutil functions on some architectures...

.. highlights:: Package Structure Changed!

   Package (a.k.a. directory) structure has been changed!! see `issue 24 <https://github.com/shirou/gopsutil/issues/24>`_

.. highlights:: golang 1.4 will become REQUIRED!

   Since syscall package becomes frozen, we should use golang/x/sys of golang 1.4 as soon as possible.


Available Architectures
------------------------------------

- FreeBSD i386/amd64
- Linux i386/amd64/arm(raspberry pi)
- Windows/amd64
- Darwin/amd64

All works are implemented without cgo by porting c struct to golang struct.


Usage
---------

.. code:: go

   import (
   	"fmt"

   	"github.com/shirou/gopsutil/mem"
   )

   func main() {
   	v, _ := mem.VirtualMemory()

   	// almost every return value is a struct
   	fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

   	// convert to JSON. String() is also implemented
   	fmt.Println(v)
   }

The output is below.

::

  Total: 3179569152, Free:284233728, UsedPercent:84.508194%
  {"total":3179569152,"available":492572672,"used":2895335424,"usedPercent":84.50819439828305, (snip)}


Documentation
------------------------

see http://godoc.org/github.com/shirou/gopsutil


More Info
--------------------

Several methods have been added which are not present in psutil, but will provide useful information.

- host/HostInfo()  (linux)

  - Hostname
  - Uptime
  - Procs
  - OS                    (ex: "linux")
  - Platform              (ex: "ubuntu", "arch")
  - PlatformFamily        (ex: "debian")
  - PlatformVersion       (ex: "Ubuntu 13.10")
  - VirtualizationSystem  (ex: "LXC")
  - VirtualizationRole    (ex: "guest"/"host")

- cpu/CPUInfo()  (linux, freebsd)

  - CPU          (ex: 0, 1, ...)
  - VendorID     (ex: "GenuineIntel")
  - Family
  - Model
  - Stepping
  - PhysicalID
  - CoreID
  - Cores        (ex: 2)
  - ModelName    (ex: "Intel(R) Core(TM) i7-2640M CPU @ 2.80GHz")
  - Mhz
  - CacheSize
  - Flags        (ex: "fpu vme de pse tsc msr pae mce cx8 ...")

- load/LoadAvg()  (linux, freebsd)

  - Load1
  - Load5
  - Load15

- docker/GetDockerIDList() (linux only)

  - container id list ([]string)

- docker/CgroupCPU() (linux only)

  - user
  - system

- docker/CgroupMem() (linux only)

  - various status

Some codes are ported from Ohai. many thanks.


Current Status
------------------

- x: work
- b: almost work but something broken

================= ====== ======= ====== =======
name              Linux  FreeBSD MacOSX Windows
cpu_times            x      x              x
cpu_count            x      x      x       x
cpu_percent          x      x              x
cpu_times_percent    x      x              x
virtual_memory       x      x      x       x
swap_memory          x      x      x
disk_partitions      x      x      x       x
disk_io_counters     x      x
disk_usage           x      x      x       x
net_io_counters      x      x      b       x
boot_time            x      x      x       x
users                x      x      x       x
pids                 x      x      x       x
pid_exists           x      x      x       x
net_connections
================= ====== ======= ====== =======

Process class
^^^^^^^^^^^^^^^

================ ===== ======= ====== =======
name             Linux FreeBSD MacOSX Windows
pid                 x     x      x       x
ppid                x     x      x       x
name                x     x      x       x
cmdline             x            x       x
create_time         x
status              x     x      x
cwd                 x
exe                 x     x              x
uids                x     x      x
gids                x     x      x
terminal            x     x      x
io_counters         x
nice                x            x       x
num_fds             x
num_ctx_switches    x
num_threads         x     x      x       x
cpu_times           x
memory_info         x     x      x
memory_info_ex      x
memory_maps         x
open_files          x
send_signal         x     x      x
suspend             x     x      x
resume              x     x      x
terminate           x     x      x
kill                x     x      x
username            x           
ionice
rlimit
num_handlres
threads
cpu_percent
cpu_affinity
memory_percent
children
connections
is_running
================ ===== ======= ====== =======

Original Metrics
^^^^^^^^^^^^^^^^^^^
================== ===== ======= ====== =======
item               Linux FreeBSD MacOSX Windows
**HostInfo**
hostname              x     x      x       x
  uptime              x     x      x
  proces              x     x
  os                  x     x      x       x
  platform            x     x      x
  platformfamiliy     x     x      x
  virtualization      x
**CPU**       
  VendorID            x     x      x       x
  Family              x     x      x       x
  Model               x     x      x       x
  Stepping            x     x      x       x
  PhysicalID          x
  CoreID              x
  Cores               x                    x
  ModelName           x     x      x       x
**LoadAvg**              
  Load1               x     x      x
  Load5               x     x      x
  Load15              x     x      x
**GetDockerID**          
  container id        x     no    no      no
**CgroupsCPU**           
  user                x     no    no      no
  system              x     no    no      no
**CgroupsMem**           
  various             x     no    no      no
================== ===== ======= ====== =======

- future work

  - process_iter
  - wait_procs
  - Process class

    - parent (use ppid instead)
    - as_dict
    - wait


License
------------

New BSD License (same as psutil)


Related Works
-----------------------

I have been influenced by the following great works:

- psutil: http://pythonhosted.org/psutil/
- dstat: https://github.com/dagwieers/dstat
- gosiger: https://github.com/cloudfoundry/gosigar/
- goprocinfo: https://github.com/c9s/goprocinfo
- go-ps: https://github.com/mitchellh/go-ps
- ohai: https://github.com/opscode/ohai/
- bosun: https://github.com/bosun-monitor/bosun/tree/master/cmd/scollector/collectors
- mackerel: https://github.com/mackerelio/mackerel-agent/tree/master/metrics

How to Contribute
---------------------------

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
4. Push to the branch (git push origin my-new-feature)
5. Create new Pull Request

My English is terrible, so documentation or correcting comments are also
welcome.
