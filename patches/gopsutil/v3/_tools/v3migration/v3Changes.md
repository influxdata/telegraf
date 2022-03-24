# v2 to v3 changes

- v3 is in the `v3` directory

- [process] RLimit is now uint64 ([#364](https://github.com/shirou/gopsutil/issues/364))
- [process] Remove process.NetIOCounters ([#429](https://github.com/shirou/gopsutil/issues/429))
- [docker] fix typo of memoryLimitInBbytes  ([#464](https://github.com/shirou/gopsutil/issues/464))
- [mem] VirtualMemoryStat JSON fields capitalization ([#545](https://github.com/shirou/gopsutil/issues/545))
  - various JSON field name and some of Variable name have been changed. see v3migration.sh
- [all] various kind of platform dependent values/constants such as process.GetWin32Proc is now private. see v3migration.sh
- [process] process.Status() now returns []string. and status string is "Running", not just "R". defined in process.go. ([#596](https://github.com/shirou/gopsutil/issues/596))
- [docker] `CgroupCPU()` now returns `*CgroupCPUStat` with Usage  ([#590](https://github.com/shirou/gopsutil/issues/590) and [#581](https://github.com/shirou/gopsutil/issues/581))
- [disk] `disk.Opts` is now string[], not string. (related to [#955](https://github.com/shirou/gopsutil/issues/955))
- [host] Fixed temperature sensors detection in Linux ([#905](https://github.com/shirou/gopsutil/issues/905))
- [disk] `GetDiskSerialNumber()` is now `SerialNumber()` and spread to all platforms
- [disk] `GetLabel ()` is now `Label()` and spread to all platform
- [net] Change net.InterfaceStat.Addrs to InterfaceAddrList ([#226](https://github.com/shirou/gopsutil/issues/226))
- [cpu] Removed windows-specific `ProcInfo()`
