# Proc Values Plugin

Read integer values from proc files.

### Configuration:

```
# Sample Config:
[[inputs.procval]]
  [inputs.procval.files]
    ## specify list of proc files to read
	# fieldName = /proc/path/to/procfile
	#
	# for example if you want to measure the available
	# entropy on the system:
	entropy = "/proc/sys/kernel/random/entropy_avail"
```

### Example output:

```
telegraf -config telegraf.conf -test -input-filter procval -test
* Plugin: procval, Collection 1
> procval entropy=867i 1458074832696755386
```
