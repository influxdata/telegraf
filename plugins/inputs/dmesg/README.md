# dmesg regex counter Plugin

The dmesg plugin counts occurances of particular regexes in the kernel log (dmesg)

### Configuration:

```toml
# Get message counts from dmesg
[[inputs.dmesg]]
  ## some basic dmesg regexes
  filters = [{"filter": ".*oom_reaper.*|.*Out of memory.*", "field": "oom.count"},
  		 	 {"filter": ".*Power-on or device reset occurred.*", "field": "device.reset"},
			 {"filter": ".*I/O error.*", "field": "io.error"},
			 {"filter": ".*MCE MEMORY.*", "field": "mce.memory.errors"}]
  dmesg_binary = "/usr/bin/dmesg"
  ## CLI options for the dmesg binary (-T, -H, etc.)
  options = []
```

### Measurements & Fields:

Measurements are generated based on the filters defined. Our default collection would show the following:

- dmesg
  - oom.count
  - device.reset
  - io.error
  - mce.memory.errors

### Tags:

No additional tags applied

### Example Output:

Assuming our default regex list, output would appear like this:

```
$ telegraf --config ~/ws/telegraf.conf --input-filter dmesg --test
* Plugin: dmesg, Collection 1
> dmesg oom.count=0,device.reset=0,io.error=0,mce.memory.errors=0 1617814276000000000
```
