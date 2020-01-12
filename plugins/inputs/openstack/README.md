# OpenStack Input Plugin

Collects the following metrics from OpenStack:

* Identity
    * Number of projects
* Compute
    * Per-hypervisor VCPUs (used/available), memory (used/avaialable) & running VMs
    * Per-server server status (e.g. running, suspended), VCPUs, memory & disk
* Block Storage
    * Per-volume size and type
    * Per-storage pool utilization

At present this plugin requires the following APIs:

* Keystone V3
* Nova V2
* Cinder V2

### Configuration

```
# Read metrics about an OpenStack cloud
# [[inputs.openstack]]
#   ## This is the recommended interval to poll so as not to overwhelm APIs
#   interval = '30m'
#
#   ## The identity endpoint to authenticate against and get the
#   ## service catalog from
#   identity_endpoint = "https://my.openstack.cloud:5000"
#
#   ## The domain to authenticate against when using a V3
#   ## identity endpoint.  Defaults to 'default'
#   domain = "default"
#
#   ## The project to authenticate as
#   project = "admin"
#
#   ## The user to authenticate as, must have admin rights
#   username = "admin"
#
#   ## The user's password to authenticate with
#   password = "Passw0rd"
```

_NB:_ Note that the recommended polling interval is 30 minutes.  This can be
reduced on smaller deployments with a handful of VMs, but will need to
be increased on estates with 100s or 1000s of VMs as it can have a
performance impact.

### Measurements & Fields

* openstack_identity
    * projects - Total number of projects [int]
* openstack_hypervisor
    * memory_mb - Memory available [int, megabytes]
    * memory_used_mb - Memory used [int, megabytes]
    * running_vms - Running VMs [int]
    * vcpus - VCPUs available [int]
    * vcpus_used - VCPUs used [int]
* openstack_server
    * status - VM status [string]
    * vcpus - VCPUs used [int]
    * ram_mb - RAM used [int, megabytes]
    * disk_gb - Disk used [int, gigabytes]
* openstack_volume
    * size_gb - Disk used [int, gigabytes]
* openstack_storage_pool
    * total_capacity_gb - Total size of storage pool [float64, bytes]
    * free_capacity_gb - Remaining size of storage pool [float64, bytes]

### Tags

* openstack_hypervisor
    * name - The hypervisor name for which the measurement is taken
* openstack_server, openstack_volume
    * name = The name of the resource
    * project - The project that a resource belongs to
    * type - The volume type
* openstack_storage_pool
    * name - The pool being refered to

### Example Output

```
simon@influxdb:~$ ./go/bin/telegraf -test -config telegraf.conf -input-filter openstack
> openstack_identity,host=symphony projects=16i 1534251270000000000
> openstack_hypervisor,host=symphony,name=compute0 memory_mb=128921i,memory_mb_used=37376i,running_vms=3i,vcpus=16i,vcpus_used=10i 1534251270000000000
> openstack_server,host=symphony,name=kubernetes,project=kube disk_gb=100i,ram_mb=32768i,status="active",vcpus=8i 1534251270000000000
> openstack_volume,host=symphony,name=volume0,project=test,type=ssd size_gb=12i 1478616110000000000
> openstack_storage_pool,host=symphony,name=cinder.volumes.flash total_capacity=86,free_capacity=45.64 1497012342000000000
```
