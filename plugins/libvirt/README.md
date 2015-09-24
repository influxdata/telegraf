# Telegraf Plugin: libvirt

#### Description

The libvirt plugin collects libvirt statistics.

To test this plugin set the following configuration:

```toml
[libvirt]
  uri = "test:///default"
```

This mocks a libvirt deamon with one running domain. The URI for a connection to a local qemu would be
`qemu:///system`.

## Resources

* http://wiki.libvirt.org/page/UbuntuKVMWalkthrough
* http://godoc.org/github.com/alexzorin/libvirt-go

## Measurements:

Meta:
- units: int64
- tags: `domain=vader`

Measurements names:
- libvirt_cpu_time
- libvirt_max_mem
- libvirt_memory
- libvirt_nr_virt_cpu
