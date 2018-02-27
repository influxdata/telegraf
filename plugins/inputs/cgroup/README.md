# CGroup Input Plugin For Telegraf Agent

This input plugin will capture specific statistics per cgroup.

Consider restricting paths to the set of cgroups you really
want to monitor if you have a large number of cgroups, to avoid
any cardinality issues.

Following file formats are supported:

* Single value

```
VAL\n
```

* New line separated values

```
VAL0\n
VAL1\n
```

* Space separated values

```
VAL0 VAL1 ...\n
```

* New line separated key-space-value's

```
KEY0 VAL0\n
KEY1 VAL1\n
```


### Tags:

All measurements have the following tags:
  - path


### Configuration:

```
# [[inputs.cgroup]]
  # paths = [
  #   "/cgroup/memory",           # root cgroup
  #   "/cgroup/memory/child1",    # container cgroup
  #   "/cgroup/memory/child2/*",  # all children cgroups under child2, but not child2 itself
  # ]
  # files = ["memory.*usage*", "memory.limit_in_bytes"]

# [[inputs.cgroup]]
  # paths = [
  #   "/cgroup/cpu",              # root cgroup
  #   "/cgroup/cpu/*",            # all container cgroups
  #   "/cgroup/cpu/*/*",          # all children cgroups under each container cgroup
  # ]
  # files = ["cpuacct.usage", "cpu.cfs_period_us", "cpu.cfs_quota_us"]
```
