# S.M.A.R.T metrics collection

A linux-only collector to query metrics from smartmontools for S.M.A.R.T-capable
hard drives (HDDs), solid-state drives (SSDs) and others. 

Read more on smartmontools at the [official](https://www.smartmontools.org/) link here.
Read more on S.M.A.R.T and its origin [here](https://en.wikipedia.org/wiki/S.M.A.R.T.)

# Sample Config 
```
  ## smartctl requires installation of the smartmontools for your distro (linux only)
  ## along with root permission to run. In this collector we presume sudo access to the 
  ## binary.
  ##
  ## Users have the ability to specify an list of disk name to include, to exclude, 
  ## or both. In this iteration of the collectors, you must specify the full smartctl
  ## path for the disk, we are not currently supporting regex. For example, to include/exclude
  ## /dev/sda from your list, you would specify:
  ## include = ["/dev/sda -d scsi"]
  ## exclude = ['/dev/sda -d scsi"]
  ## 
  ## NOTE: If you specify an include list, this will skip the smartctl --scan function
  ## and only collect for those you've requested (minus any exclusions).
  [[inputs.smartctl]]
    include = ["/dev/bus/0 -d megaraid,24"]
    exclude = ["/dev/sda -d scsi"]
```

# Points to note
1. The smartmontools are required to run with sudo permissions. This means two things:
you will have to enable your telegraf user with the proper sudo perms to run
smartmontools; and you should make sure the smartmontools pkg is installed :D

2. This collector is set to run a smartctl scan to find all disks on a system. If
this is not what you'd like to do, you can specify the disks you'd like to collect 
by itemizing them in the `include` option in the telegraf config. This will skip 
the scan command and ONLY run smartctl on those disks in the list.

3. A corollary to the `include` set of disks is the `exclude` set. In this case,
the scan will still be run unless `include` is defined but will limit those to 
query for based on the set in the `exclude`.
