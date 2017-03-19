# libvirt plugin

The libvirt plugin collects libvirt domain statistics for the given hypervisor URI.

This plugin requires that the `libvirt-bin` package is installed. 

It uses `virsh` to gather the metrics instead of using the `libvirt-go` bindings.
The reasoning behind this is that this way no runtime dependencies on the libvirt C libaries are added to Telegraf.

### Configuration:
```toml
[[inputs.libvirt]]
  ## specify a libvirt connection uri, see https://libvirt.org/uri.html
  uri = "qemu:///system"
```

### Measurements & Fields:
- libvirt
  - cpu_time (float, seconds)
  - max_memory (uint, KiB)
  - used_memory (uint, KiB)
  - n_vcpu (uint, #)

### Tags:
- libvirt
  - domain
  - state

### Example Output:
```
$ ./telegraf -config telegraf.conf -input-filter libvirt -test
libvirt,domain=test,state=running,host=kvm_host cpu_time=1489951430,max_memory=8388608i,used_memory=2097152i,n_vcpu=2i 1489951430000000000
```
