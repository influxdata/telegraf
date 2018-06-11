package libvirt

const xmlDump = `<domain type='kvm' id='787'>
  <name>r-2825-QA</name>
  <uuid>a68ff3a5-7c65-48e5-aa0d-8559c0b0afc0</uuid>
  <description>Debian GNU/Linux 5.0 (64-bit)</description>
  <memory unit='KiB'>262144</memory>
  <currentMemory unit='KiB'>262144</currentMemory>
  <vcpu placement='static'>1</vcpu>
  <cputune>
    <shares>166</shares>
  </cputune>
  <resource>
    <partition>/machine</partition>
  </resource>
  <sysinfo type='smbios'>
    <system>
      <entry name='manufacturer'>Apache Software Foundation</entry>
      <entry name='product'>CloudStack KVM Hypervisor</entry>
      <entry name='uuid'>a68ff3a5-7c65-48e5-aa0d-8559c0b0afc0</entry>
    </system>
  </sysinfo>
  <os>
    <type arch='x86_64' machine='pc-i440fx-rhel7.0.0'>hvm</type>
    <boot dev='cdrom'/>
    <boot dev='hd'/>
    <smbios mode='sysinfo'/>
  </os>
  <features>
    <acpi/>
    <apic/>
    <pae/>
  </features>
  <cpu mode='custom' match='exact' check='full'>
    <model fallback='forbid'>SandyBridge</model>
    <feature policy='require' name='hypervisor'/>
    <feature policy='require' name='xsaveopt'/>
  </cpu>
  <clock offset='utc'>
    <timer name='kvmclock'/>
  </clock>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>
  <devices>
    <emulator>/usr/libexec/qemu-kvm</emulator>
    <disk type='network' device='disk'>
      <driver name='qemu' type='raw' cache='none'/>
      <auth username='rbdnjcloudhost'>
        <secret type='ceph' uuid='282d5a88-b298-39b9-bd65-aafb088270c2'/>
      </auth>
      <source protocol='rbd' name='rbdnjcloudhost/bacfe2c6-8be7-4261-9558-1994fc9e83e9'>
        <host name='ceph-mon.labs.ena.net' port='6789'/>
      </source>
      <backingStore/>
      <target dev='vda' bus='virtio'/>
      <serial>bacfe2c68be742619558</serial>
      <alias name='virtio-disk0'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x05' function='0x0'/>
    </disk>
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw' cache='none'/>
      <source file='/usr/share/cloudstack-common/vms/systemvm.iso'/>
      <backingStore/>
      <target dev='hdc' bus='ide'/>
      <readonly/>
      <alias name='ide0-1-0'/>
      <address type='drive' controller='0' bus='1' target='0' unit='0'/>
    </disk>
    <controller type='usb' index='0' model='piix3-uhci'>
      <alias name='usb'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x01' function='0x2'/>
    </controller>
    <controller type='pci' index='0' model='pci-root'>
      <alias name='pci.0'/>
    </controller>
    <controller type='ide' index='0'>
      <alias name='ide'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x01' function='0x1'/>
    </controller>
    <controller type='virtio-serial' index='0'>
      <alias name='virtio-serial0'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x04' function='0x0'/>
    </controller>
    <interface type='bridge'>
      <mac address='0e:00:a9:fe:03:3c'/>
      <source bridge='cloud0'/>
      <target dev='vnet22'/>
      <model type='virtio'/>
      <alias name='net0'/>
      <rom bar='off' file=''/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x0'/>
    </interface>
    <interface type='bridge'>
      <mac address='06:a3:cc:00:32:e7'/>
      <source bridge='brbond0-101'/>
      <bandwidth>
        <inbound average='6400' peak='6400'/>
        <outbound average='6400' peak='6400'/>
      </bandwidth>
      <target dev='vnet23'/>
      <model type='virtio'/>
      <alias name='net1'/>
      <rom bar='off' file=''/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x06' function='0x0'/>
    </interface>
    <interface type='bridge'>
      <mac address='06:6a:b2:00:32:f3'/>
      <source bridge='brbond0-452'/>
      <bandwidth>
        <inbound average='6400' peak='6400'/>
        <outbound average='6400' peak='6400'/>
      </bandwidth>
      <target dev='vnet32'/>
      <model type='virtio'/>
      <alias name='net2'/>
      <rom bar='off' file=''/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x07' function='0x0'/>
    </interface>
    <serial type='pty'>
      <source path='/dev/pts/13'/>
      <target port='0'/>
      <alias name='serial0'/>
    </serial>
    <console type='pty' tty='/dev/pts/13'>
      <source path='/dev/pts/13'/>
      <target type='serial' port='0'/>
      <alias name='serial0'/>
    </console>
    <channel type='unix'>
      <source mode='bind' path='/var/lib/libvirt/qemu/r-2825-QA.agent'/>
      <target type='virtio' name='r-2825-QA.vport' state='disconnected'/>
      <alias name='channel0'/>
      <address type='virtio-serial' controller='0' bus='0' port='1'/>
    </channel>
    <input type='tablet' bus='usb'>
      <alias name='input0'/>
      <address type='usb' bus='0' port='1'/>
    </input>
    <input type='mouse' bus='ps2'>
      <alias name='input1'/>
    </input>
    <input type='keyboard' bus='ps2'>
      <alias name='input2'/>
    </input>
    <graphics type='vnc' port='5912' autoport='yes' listen='10.103.0.214'>
      <listen type='address' address='10.103.0.214'/>
    </graphics>
    <video>
      <model type='cirrus' vram='16384' heads='1' primary='yes'/>
      <alias name='video0'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x02' function='0x0'/>
    </video>
    <memballoon model='none'>
      <alias name='balloon0'/>
    </memballoon>
  </devices>
  <seclabel type='none' model='none'/>
  <seclabel type='dynamic' model='dac' relabel='yes'>
    <label>+0:+0</label>
    <imagelabel>+0:+0</imagelabel>
  </seclabel>
</domain>`
