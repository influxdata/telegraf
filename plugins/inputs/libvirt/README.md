# libvirt plugin

#### Description

The libvirt plugin collects libvirt domain statistics.

To test this plugin set the following configuration:

```toml
[libvirt]
  uri = "test:///default"
```

This mocks a libvirt deamon with one running domain. The URI for a connection to a local qemu would be
`qemu:///system`.

## Resources

* http://wiki.libvirt.org/page/UbuntuKVMWalkthrough
* https://godoc.org/github.com/libvirt/libvirt-go

## Measurements:
- libvirt
  - cpu_time
  - max_mem
  - memory
  - nr_virt_cpu

## Tags:
- domain

## Example Output:
```
libvirt,domain=test,host=dev-vm cpu_time=1489860811102172000i,max_mem=8388608i,memory=2097152i,nr_virt_cpu=2i 1489860811000000000
```
