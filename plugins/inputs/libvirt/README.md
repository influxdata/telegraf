# Telegraf Plugin: libvirt

#### Description

The libvirt plugin collects libvirt statistics.

Metrics this collector generates:
libvirt.vm.count - number of running VMs
libvirt.vm.cpu.load - VM's current CPU load (can be higher than 100%)
libvirt.vm.cpu.time - CPU time spent by VM
libvirt.vm.disk.read.requests - number of total VM's read requests
libvirt.vm.disk.read.bytes - number of total VM's read bytes
libvirt.vm.disk.write.requests - number of total VM's write requests
libvirt.vm.disk.write.bytes - number of total VM's write bytes
libvirt.vm.disk.total.requests - number of total VM's read + write requests
libvirt.vm.disk.total.bytes - number of total VM's read + write bytes
libvirt.vm.disk.current.read.requests - number of VM's current read requests
libvirt.vm.disk.current.read.bytes - number of VM's current read bytes
libvirt.vm.disk.current.write.requests - number of VM's current write requests
libvirt.vm.disk.current.write.bytes - number of VM's current write bytes
libvirt.vm.disk.current.total.requests - number of VM's current read + write requests
libvirt.vm.disk.current.total.bytes - number of VM's current read + write bytes
libvirt.vm.memory - memory used by VM in kB
libvirt.vm.max.memory - memory requested in VM's template in kB
libvirt.vm.max.vcpus - number of CPU requested in VM's template
libvirt.vm.network.rx - number of VM's received bytes via network
libvirt.vm.network.tx - number of VM's transmitted bytes via network
libvirt.vm.network.current.rx - VM's current network incoming bandwidth
libvirt.vm.network.current.tx - VM's current network outcoming bandwidth
libvirt.vm.cpustat.count - number of CPUs
libvirt.vm.cpustat.cpu.mhz - CPU's MHz
libvirt.vm.cpustat.cpu.cores - CPU's number of cores


#### To test this plugin set the following configuration:

This mocks a libvirt deamon with one running domain. The URI for a connection to a local qemu would be
`qemu:///system`.

The libvirt Test driver is a per-process fake hypervisor driver,
with a driver name of 'test'. The driver maintains all its state in memory.
`test:///default`

Metrics from test file in docker owlet123/libvirt
`test+tcp://127.0.0.1/root/test.xml`