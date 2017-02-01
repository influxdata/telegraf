# hsperfdata Plugin

The plugin gathers data from Hotspot JVMs via the hsperfdata files they expose. This plugin won't work if you've disabled their creation using `-XX:-UsePerfData` or `-XX:+PerfDisableSharedMem`!

### Configuration:

```toml
[[inputs.hsperfdata]]
  # Optional: gather data from processes belonging to a different user. By
  # default, the username in the USER environment variable is used to generate
  # the hsperfdata directory name (usually "/tmp/hsperfdata_username")
  user: "root"

  # use the named keys in the hsperfdata file as tags, not fields. By default,
  # every key is exposed as a field. This example shows how to tag by JVM major
  # version:
  tags: ["java.property.java.vm.specification.version"]
```

### Measurements & Fields:

All metrics are gathered as the "java" measurement.

All keys in the hsperfdata file are exposed as fields; there's no comprehensive list as they vary by Hotspot version.

### Tags:

- All measurements have the following tags:
    - pid (the process id of the monitored process)
    - procname (the class name containing the `main` function being run)

### Example Output:

Most fields abbreviated; there's usually 200-300 of them:

```
$ ./telegraf -config telegraf.conf -input-filter example -test
java,host=nwhite91-mac,pid=17427,procname=com.sun.javaws.Main java.ci.totalTime="49874911809",...,sun.zip.zipFiles="28" 1479466710000000000
```
