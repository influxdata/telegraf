# CGroup Input Plugin

This input plugin will capture specific statistics per cgroup.

Consider restricting paths to the set of cgroups you really
want to monitor if you have a large number of cgroups, to avoid
any cardinality issues.

Following file formats are supported:

* Single value

```text
VAL\n
```

* New line separated values

```text
VAL0\n
VAL1\n
```

* Space separated values

```text
VAL0 VAL1 ...\n
```

* Space separated keys and value, separated by new line

```text
KEY0 ... VAL0\n
KEY1 ... VAL1\n
```

## Tags

All measurements have the `path` tag.

## Configuration

```toml
# Read specific statistics per cgroup
[[inputs.cgroup]]
  ## Directories in which to look for files, globs are supported.
  ## Consider restricting paths to the set of cgroups you really
  ## want to monitor if you have a large number of cgroups, to avoid
  ## any cardinality issues.
  # paths = [
  #   "/sys/fs/cgroup/memory",
  #   "/sys/fs/cgroup/memory/child1",
  #   "/sys/fs/cgroup/memory/child2/*",
  # ]
  ## cgroup stat fields, as file names, globs are supported.
  ## these file names are appended to each path from above.
  # files = ["memory.*usage*", "memory.limit_in_bytes"]
```

## Example

```toml
# [[inputs.cgroup]]
  # paths = [
  #   "/sys/fs/cgroup/cpu",              # root cgroup
  #   "/sys/fs/cgroup/cpu/*",            # all container cgroups
  #   "/sys/fs/cgroup/cpu/*/*",          # all children cgroups under each container cgroup
  # ]
  # files = ["cpuacct.usage", "cpu.cfs_period_us", "cpu.cfs_quota_us"]

# [[inputs.cgroup]]
  # paths = [
  #   "/sys/fs/cgroup/unified/*",        # root cgroup
  # ]
  # files = ["*"]
```
