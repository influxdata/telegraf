# gopsutil: psutil for golang

[![Test](https://github.com/shirou/gopsutil/actions/workflows/test.yml/badge.svg)](https://github.com/shirou/gopsutil/actions/workflows/test.yml) [![Coverage Status](https://coveralls.io/repos/github/shirou/gopsutil/badge.svg?branch=master)](https://coveralls.io/github/shirou/gopsutil?branch=master) [![Go Reference](https://pkg.go.dev/badge/github.com/shirou/gopsutil.svg)](https://pkg.go.dev/github.com/shirou/gopsutil)

This is a port of psutil (https://github.com/giampaolo/psutil). The
challenge is porting all psutil functions on some architectures.

## v3 migration

from v3.20.10, gopsutil becomes v3 which breaks backwards compatibility.
See [v3Changes.md](_tools/v3migration/v3Changes.md) more detail changes.

## Tag semantics

gopsutil tag policy is almost same as Semantic Versioning, but
automatically increase like Ubuntu versioning.

for example, v2.17.04 means

- v2: major version
- 17: release year, 2017
- 04: release month

gopsutil aims to keep backwards compatibility until major version change.

Tagged at every end of month, but if there are only a few commits, it
can be skipped.

## Available Architectures

- FreeBSD i386/amd64/arm
- Linux i386/amd64/arm(raspberry pi)
- Windows i386/amd64/arm/arm64
- Darwin i386/amd64
- OpenBSD amd64 (Thank you @mpfz0r!)
- Solaris amd64 (developed and tested on SmartOS/Illumos, Thank you
  @jen20!)

These have partial support:

- CPU on DragonFly BSD (#893, Thank you @gballet!)
- host on Linux RISC-V (#896, Thank you @tklauser!)

All works are implemented without cgo by porting C structs to golang
structs.

## Usage

```go
package main

import (
    "fmt"

    "github.com/shirou/gopsutil/v3/mem"
    // "github.com/shirou/gopsutil/mem"  // to use v2
)

func main() {
    v, _ := mem.VirtualMemory()

    // almost every return value is a struct
    fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

    // convert to JSON. String() is also implemented
    fmt.Println(v)
}
```

The output is below.

    Total: 3179569152, Free:284233728, UsedPercent:84.508194%
    {"total":3179569152,"available":492572672,"used":2895335424,"usedPercent":84.50819439828305, (snip...)}

You can set an alternative location to `/proc` by setting the `HOST_PROC`
environment variable.

You can set an alternative location to `/sys` by setting the `HOST_SYS`
environment variable.

You can set an alternative location to `/etc` by setting the `HOST_ETC`
environment variable.

You can set an alternative location to `/var` by setting the `HOST_VAR`
environment variable.

You can set an alternative location to `/run` by setting the `HOST_RUN`
environment variable.

You can set an alternative location to `/dev` by setting the `HOST_DEV`
environment variable.

## Documentation

see http://godoc.org/github.com/shirou/gopsutil

## Requirements

- go1.16 or above is required.

## More Info

Several methods have been added which are not present in psutil, but
will provide useful information.

- host/HostInfo() (linux)
  - Hostname
  - Uptime
  - Procs
  - OS (ex: "linux")
  - Platform (ex: "ubuntu", "arch")
  - PlatformFamily (ex: "debian")
  - PlatformVersion (ex: "Ubuntu 13.10")
  - VirtualizationSystem (ex: "LXC")
  - VirtualizationRole (ex: "guest"/"host")
- IOCounters
  - Label (linux only) The registered [device mapper
    name](https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-block-dm)
- cpu/CPUInfo() (linux, freebsd)
  - CPU (ex: 0, 1, ...)
  - VendorID (ex: "GenuineIntel")
  - Family
  - Model
  - Stepping
  - PhysicalID
  - CoreID
  - Cores (ex: 2)
  - ModelName (ex: "Intel(R) Core(TM) i7-2640M CPU @ 2.80GHz")
  - Mhz
  - CacheSize
  - Flags (ex: "fpu vme de pse tsc msr pae mce cx8 ...")
  - Microcode
- load/Avg() (linux, freebsd, solaris)
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
- net_protocols (linux only)
  - system wide stats on network protocols (i.e IP, TCP, UDP, etc.)
  - sourced from /proc/net/snmp
- iptables nf_conntrack (linux only)
  - system wide stats on netfilter conntrack module
  - sourced from /proc/sys/net/netfilter/nf_conntrack_count

Some code is ported from Ohai. many thanks.

## Current Status

- x: works
- b: almost works, but something is broken

|name                  |Linux  |FreeBSD  |OpenBSD  |macOS   |Windows  |Solaris  |Plan 9   |
|----------------------|-------|---------|---------|--------|---------|---------|---------|
|cpu\_times            |x      |x        |x        |x       |x        |         |b        |
|cpu\_count            |x      |x        |x        |x       |x        |         |x        |
|cpu\_percent          |x      |x        |x        |x       |x        |         |         |
|cpu\_times\_percent   |x      |x        |x        |x       |x        |         |         |
|virtual\_memory       |x      |x        |x        |x       |x        | b       |x        |
|swap\_memory          |x      |x        |x        |x       |         |         |x        |
|disk\_partitions      |x      |x        |x        |x       |x        |         |         |
|disk\_io\_counters    |x      |x        |x        |        |         |         |         |
|disk\_usage           |x      |x        |x        |x       |x        |         |         |
|net\_io\_counters     |x      |x        |x        |b       |x        |         |         |
|boot\_time            |x      |x        |x        |x       |x        |         |         |
|users                 |x      |x        |x        |x       |x        |         |         |
|pids                  |x      |x        |x        |x       |x        |         |         |
|pid\_exists           |x      |x        |x        |x       |x        |         |         |
|net\_connections      |x      |         |x        |x       |         |         |         |
|net\_protocols        |x      |         |         |        |         |         |         |
|net\_if\_addrs        |       |         |         |        |         |         |         |
|net\_if\_stats        |       |         |         |        |         |         |         |
|netfilter\_conntrack  |x      |         |         |        |         |         |         |


### Process class

|name                |Linux  |FreeBSD  |OpenBSD  |macOS  |Windows  |
|--------------------|-------|---------|---------|-------|---------|
|pid                 |x      |x        |x        |x      |x        |
|ppid                |x      |x        |x        |x      |x        |
|name                |x      |x        |x        |x      |x        |
|cmdline             |x      |x        |         |x      |x        |
|create\_time        |x      |         |         |x      |x        |
|status              |x      |x        |x        |x      |         |
|cwd                 |x      |         |         |x      |         |
|exe                 |x      |x        |x        |       |x        |
|uids                |x      |x        |x        |x      |         |
|gids                |x      |x        |x        |x      |         |
|terminal            |x      |x        |x        |       |         |
|io\_counters        |x      |x        |x        |       |x        |
|nice                |x      |x        |x        |x      |x        |
|num\_fds            |x      |         |         |       |         |
|num\_ctx\_switches  |x      |         |         |       |         |
|num\_threads        |x      |x        |x        |x      |x        |
|cpu\_times          |x      |         |         |       |x        |
|memory\_info        |x      |x        |x        |x      |x        |
|memory\_info\_ex    |x      |         |         |       |         |
|memory\_maps        |x      |         |         |       |         |
|open\_files         |x      |         |         |       |         |
|send\_signal        |x      |x        |x        |x      |         |
|suspend             |x      |x        |x        |x      |         |
|resume              |x      |x        |x        |x      |         |
|terminate           |x      |x        |x        |x      |x        |
|kill                |x      |x        |x        |x      |         |
|username            |x      |x        |x        |x      |x        |
|ionice              |       |         |         |       |         |
|rlimit              |x      |         |         |       |         |
|num\_handlers       |       |         |         |       |         |
|threads             |x      |         |         |       |         |
|cpu\_percent        |x      |         |x        |x      |         |
|cpu\_affinity       |       |         |         |       |         |
|memory\_percent     |       |         |         |       |         |
|parent              |x      |         |x        |x      |x        |
|children            |x      |x        |x        |x      |x        |
|connections         |x      |         |x        |x      |         |
|is\_running         |       |         |         |       |         |
|page\_faults        |x      |         |         |       |         |

### Original Metrics

|item             |Linux  |FreeBSD  |OpenBSD  |macOS   |Windows |Solaris  |
|-----------------|-------|---------|---------|--------|--------|---------|
|**HostInfo**     |       |         |         |        |        |         |
|hostname         |x      |x        |x        |x       |x       |x        |
|uptime           |x      |x        |x        |x       |        |x        |
|process          |x      |x        |x        |        |        |x        |
|os               |x      |x        |x        |x       |x       |x        |
|platform         |x      |x        |x        |x       |        |x        |
|platformfamily   |x      |x        |x        |x       |        |x        |
|virtualization   |x      |         |         |        |        |         |
|**CPU**          |       |         |         |        |        |         |
|VendorID         |x      |x        |x        |x       |x       |x        |
|Family           |x      |x        |x        |x       |x       |x        |
|Model            |x      |x        |x        |x       |x       |x        |
|Stepping         |x      |x        |x        |x       |x       |x        |
|PhysicalID       |x      |         |         |        |        |x        |
|CoreID           |x      |         |         |        |        |x        |
|Cores            |x      |         |         |        |x       |x        |
|ModelName        |x      |x        |x        |x       |x       |x        |
|Microcode        |x      |         |         |        |        |x        |
|**LoadAvg**      |       |         |         |        |        |         |
|Load1            |x      |x        |x        |x       |        |         |
|Load5            |x      |x        |x        |x       |        |         |
|Load15           |x      |x        |x        |x       |        |         |
|**GetDockerID**  |       |         |         |        |        |         |
|container id     |x      |no       |no       |no      |no      |         |
|**CgroupsCPU**   |       |         |         |        |        |         |
|user             |x      |no       |no       |no      |no      |         |
|system           |x      |no       |no       |no      |no      |         |
|**CgroupsMem**   |       |         |         |        |        |         |
|various          |x      |no       |no       |no      |no      |         |

- future work
  - process_iter
  - wait_procs
  - Process class
    - as_dict
    - wait

## License

New BSD License (same as psutil)

## Related Works

I have been influenced by the following great works:

- psutil: https://github.com/giampaolo/psutil
- dstat: https://github.com/dagwieers/dstat
- gosigar: https://github.com/cloudfoundry/gosigar/
- goprocinfo: https://github.com/c9s/goprocinfo
- go-ps: https://github.com/mitchellh/go-ps
- ohai: https://github.com/opscode/ohai/
- bosun:
  https://github.com/bosun-monitor/bosun/tree/master/cmd/scollector/collectors
- mackerel:
  https://github.com/mackerelio/mackerel-agent/tree/master/metrics

## How to Contribute

1.  Fork it
2.  Create your feature branch (git checkout -b my-new-feature)
3.  Commit your changes (git commit -am 'Add some feature')
4.  Push to the branch (git push origin my-new-feature)
5.  Create new Pull Request

English is not my native language, so PRs correcting grammar or spelling
are welcome and appreciated.
