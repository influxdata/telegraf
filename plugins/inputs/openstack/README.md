# OpenStack Input Plugin

Collects the metrics from following services of OpenStack:

* NOVA(Compute Service)
* CINDER(Block Storage)
* MANILA(Shared filesystems)
* NEUTRON(Networking)
* KEYSTONE(Identity service)
* PLACEMENT(Placement service)
* GLANCE(Image service)
* HEAT(Orchestration)

At present this plugin requires the following APIs:

* blockstorage  v2
* compute  v2
* identity  v3
* orchestration  v1
* sharedfilesystems  v2
* networking  v2




### Configuration

```
#  ## This is the recommended interval to poll.
#  interval = '30m'
#
#  ## The identity endpoint to authenticate against and get the
#  ## service catalog from
#  identity_endpoint = "https://my.openstack.cloud:5000"
#
#  ## The domain to authenticate against when using a V3
#  ## identity endpoint.  Defaults to 'default'
#  domain = "default"
#
#  ## The project to authenticate as
#  project = "admin"
#
#  ## The user to authenticate as, must have admin rights
#  username = "admin"
#
#  ## The user's password to authenticate with
#  password = "Passw0rd"
#
#  ## Services to be enabled
#  #enabled_services = ["stacks","services", "projects", "hypervisors", "flavors", "servers", "volumes", "storage" , "subnets", "ports", "networks", "aggregates", "shares", "nova_services", "agents"]
#  enabled_services = ["services", "projects", "hypervisors", "flavors", "networks", "volumes"]
# 
#  #Dependencies
#  # | Service | Depends on |
#  # | servers | projects, hypervisors, flavors |
#  # | volumes | projects |  
#
#  ## Collect Server Diagnostics
#  server_diagnotics = false
#
#  InsecureSkipVerify = false
```

_NB:_ Note that the recommended polling interval is 30 minutes.  This can be
reduced on smaller deployments with a handful of VMs, but will need to
be increased on estates with 100s or 1000s of VMs as it can have a
performance impact.

### Measurements & Fields


* openstack_flavor   
    * disk        [integer]
    * ephemeral   [integer]
    * id          [string]
    * is_public   [boolean]
    * name        [string]
    * ram         [integer]
    * rxtx_factor [float]
    * swap_mb     [integer]
    * vcpus       [integer]
* openstack_hypervisor              
    * cpu_arch                [string]
    * cpu_model               [string]
    * cpu_topology_cores      [integer]
    * cpu_topology_sockets    [integer]
    * cpu_topology_threads    [integer]
    * cpu_vendor              [string]
    * current_workload        [integer]
    * disk_available_least    [integer]
    * free_disk_gb            [integer]
    * free_ram_mb             [integer]
    * host_ip                 [string]
    * hypervisor_hostname     [string]
    * hypervisor_type         [string]
    * id                      [string]
    * local_gb                [integer]
    * local_gb_used           [integer]
    * memory_mb               [integer]
    * memory_mb_used          [integer]
    * running_vms             [integer]
    * service_disabled_reason [string]
    * service_host            [string]
    * service_id              [string]
    * state                   [string]
    * status                  [string]
    * vcpus                   [integer]
    * vcpus_used              [integer]
    * version                 [integer]
* openstack_identity
    * description [string]
    * domain_id   [string]
    * enabled     [boolean]
    * id          [string]
    * is_domain   [boolean]
    * name        [string]
    * parent_id   [string]
    * projects    [integer]
* openstack_network
    * admin_state_up          [boolean]
    * availability_zone_hints [integer]
    * created_at              [string]
    * description             [string]
    * id                      [string]
    * name                    [string]
    * project_id              [string]
    * shared                  [boolean]
    * status                  [string]
    * subnets                 [integer]
    * tenant_id               [string]
    * updated_at              [string]
* openstack_nova_service
    * binary          [string]
    * disabled_reason [string]
    * host_machine    [string]
    * id              [string]
    * state           [string]
    * status          [string]
    * updated_at      [string]
    * zone            [string]
* openstack_port
    * admin_state_up        [boolean]
    * allowed_address_pairs [integer]
    * description           [string]
    * device_id             [string]
    * device_owner          [string]
    * fixed_ips             [integer]
    * id                    [string]
    * mac_address           [string]
    * name                  [string]
    * network_id            [string]
    * project_id            [string]
    * security_groups       [integer]
    * status                [string]
    * tenant_id             [string]
* openstack_server
    * accessIPv4       [string]
    * accessIPv6       [string]
    * addresses        [integer]
    * adminPass        [string]
    * created          [string]
    * disk_gb          [integer]
    * fault_code       [integer]
    * fault_details    [string]
    * fault_message    [string]
    * flavor           [string]
    * host_id          [string]
    * host_name        [string]
    * id               [string]
    * image            [string]
    * key_name         [string]
    * name             [string]
    * progress         [integer]
    * ram_mb           [integer]
    * security_groups  [integer]
    * status           [string]
    * tenant_id        [string]
    * updated          [string]
    * user_id          [string]
    * vcpus            [integer]
    * volumes_attached [integer]
* openstack_server_diagnostics
    * cpu0_time          [float]
    * cpu1_time          [float]
    * cpu2_time          [float]
    * cpu3_time          [float]
    * cpu4_time          [float]
    * cpu5_time          [float]
    * cpu6_time          [float]
    * cpu7_time          [float]
    * hda_errors         [float]
    * hda_read           [float]
    * hda_read_req       [float]
    * hda_write          [float]
    * hda_write_req      [float]
    * memory             [float]
    * memory-actual      [float]
    * memory-available   [float]
    * memory-last_update [float]
    * memory-major_fault [float]
    * memory-minor_fault [float]
    * memory-rss         [float]
    * memory-swap_in     [float]
    * memory-swap_out    [float]
    * memory-unused      [float]
    * memory-usable      [float]
    * no_of_ports        [integer]
    * port_1_rx          [float]
    * port_1_rx_drop     [float]
    * port_1_rx_errors   [float]
    * port_1_rx_packets  [float]
    * port_1_tx          [float]
    * port_1_tx_drop     [float]
    * port_1_tx_errors   [float]
    * port_1_tx_packets  [float]
    * port_2_rx          [float]
    * port_2_rx_drop     [float]
    * port_2_rx_errors   [float]
    * port_2_rx_packets  [float]
    * port_2_tx          [float]
    * port_2_tx_drop     [float]
    * port_2_tx_errors   [float]
    * port_2_tx_packets  [float]
    * port_3_rx          [float]
    * port_3_rx_drop     [float]
    * port_3_rx_errors   [float]
    * port_3_rx_packets  [float]
    * port_3_tx          [float]
    * port_3_tx_drop     [float]
    * port_3_tx_errors   [float]
    * port_3_tx_packets  [float]
    * server_id          [string]
    * vda_errors         [float]
    * vda_read           [float]
    * vda_read_req       [float]
    * vda_write          [float]
    * vda_write_req      [float]
* openstack_service
    * name            [string]
    * service_enabled [boolean]
    * service_id      [string]
* openstack_subnet
    * cidr              [string]
    * dhcp_enabled      [boolean]
    * dns_nameservers   [integer]
    * gateway_ip        [string]
    * ip_version        [string]
    * ipv6_address_mode [string]
    * ipv6_ra_mode      [string]
    * name              [string]
    * network_id        [string]
    * project_id        [string]
    * subnet_id         [string]
    * subnet_pool_id    [string]
    * tenant_id         [string]
* openstack_volume
    * attachment_attached_at   [string]
    * attachment_attachment_id [string]
    * attachment_device        [string]
    * attachment_host_name     [string]
    * attachment_id            [string]
    * attachment_server_id     [string]
    * attachment_volume_id     [string]
    * availability_zone        [string]
    * bootable                 [string]
    * consistency_group_id     [string]
    * description              [string]
    * encrypted                [boolean]
    * id                       [string]
    * multiattach              [boolean]
    * name                     [string]
    * replication_status       [string]
    * size                     [integer]
    * size_gb                  [integer]
    * snapshot_id              [string]
    * source_volid             [string]
    * status                   [string]
    * total_attachments        [integer]
    * user_id                  [string]
    * volume_type              [string]


### Tags

* openstack_flavor
    * host
    * id
* openstack_hypervisor
    * host
    * id
* openstack_identity
    * host
    * id
* openstack_network
    * host
    * id
    * tenant_id
* openstack_nova_service
    * binary
    * host
    * id
* openstack_port
    * host
    * id
    * network_id
    * status
* openstack_server
    * host
    * host_id
    * id
    * name
    * project
    * tenant_id
* openstack_server_diagnostics
    * host
    * port_1
    * port_2
    * port_3
    * server_id
* openstack_service
    * host
    * service_id
* openstack_subnet
    * host
    * subnet_id
* openstack_volume
    * attachment_server_id
    * host
    * id
    * name
    * project
    * type

### Example Output

```
> openstack_port,host=ubuntu,id=65b564e8-b5d6-4f5e-88de-b8c628865a82,network_id=fa0d1093-3461-41c3-a770-955e08388d8e,status=ACTIVE admin_state_up=true,allowed_address_pairs=0i,description="",device_id="bdcaa2f0-647d-4da4-b701-197f3659610a",device_owner="compute:nova",fixed_ips=1i,id="65b564e8-b5d6-4f5e-88de-b8c628865a82",mac_address="fa:16:3e:12:31:06",name="",network_id="fa0d1093-3461-41c3-a770-955e08388d8e",project_id="6226db488eb446ea89564fa29d9340d0",security_groups=1i,status="ACTIVE",tenant_id="6226db488eb446ea89564fa29d9340d0" 1619543517000000000
> openstack_flavor,host=ubuntu,id=4 disk=10i,ephemeral=0i,id="4",is_public=true,name="large",ram=8192i,rxtx_factor=1,swap_mb=0i,vcpus=4i 1619543517000000000
> openstack_volume,attachment_server_id=b1753db6-d22c-4f9d-a336-3d309a9f8694,host=ubuntu,id=1ca235c4-e3fe-4f86-8eb3-bc15babc55a8,name=uc-mount,project=admin,type= attachment_attached_at="2021-01-25T16:17:54Z",attachment_attachment_id="b9eb2861-c416-423d-8f11-d1aca354945f",attachment_device="/dev/vdb",attachment_host_name="localhost",attachment_id="1ca235c4-e3fe-4f86-8eb3-bc15babc55a8",attachment_server_id="b1753db6-d22c-4f9d-a336-3d309a9f8694",attachment_volume_id="1ca235c4-e3fe-4f86-8eb3-bc15babc55a8",availability_zone="nova",bootable="false",consistency_group_id="",description="",encrypted=false,id="1ca235c4-e3fe-4f86-8eb3-bc15babc55a8",multiattach=false,name="uc-mount",replication_status="",size=955i,size_gb=955i,snapshot_id="",source_volid="",status="in-use",total_attachments=1i,user_id="529a2f3e43954995ba768f4c20802cbb",volume_type="" 1619543517000000000
> openstack_service,host=ubuntu,service_id=f2b5acb4564f4aababb519c76b9b41a0 name="volume",service_enabled=true,service_id="f2b5acb4564f4aababb519c76b9b41a0" 1619543517000000000
> openstack_subnet,host=ubuntu,subnet_id=7408e585-113f-4f27-b7ef-0be200180544 cidr="1.1.0.0/24",dhcp_enabled=true,dns_nameservers=0i,gateway_ip="1.1.0.1",ip_version="4",ipv6_address_mode="",ipv6_ra_mode="",name="sub",network_id="fa0d1093-3461-41c3-a770-955e08388d8e",project_id="6226db488eb446ea89564fa29d9340d0",subnet_id="7408e585-113f-4f27-b7ef-0be200180544",subnet_pool_id="",tenant_id="6226db488eb446ea89564fa29d9340d0" 1619543517000000000
> openstack_nova_service,binary=nova-consoleauth,host=ubuntu,id=41 binary="nova-consoleauth",disabled_reason="",host_machine="machine",id="41",state="up",status="enabled",updated_at="2099-04-27T17:11:52Z",zone="internal" 1619543517000000000
> openstack_server,host=ubuntu,host_id=c7ad246fa36d345debf793d00162ea6aa89764694eb9500a06c0fe2d,id=bdcaa2f0-647d-4da4-b701-197f3659610a,name=servername,project=admin,tenant_id=6226db488eb446ea89564fa29d9340d0 accessIPv4="",accessIPv6="",addresses=3i,adminPass="",created="2020-09-14T14:50:08Z",disk_gb=80i,fault_code=0i,fault_details="",fault_message="",flavor="9",host_id="c7ad246fa36d345debf793d00162ea6aa89764694eb9500a06c0fe2d",host_name="localhost",id="bdcaa2f0-647d-4da4-b701-197f3659610a",image="95f4ec5d-56ac-45c2-9c65-46301707c1a4",key_name="key",name="servername",progress=0i,ram_mb=32768i,security_groups=3i,status="shutoff",tenant_id="6226db488eb446ea89564fa29d9340d0",updated="2020-09-19T03:50:59Z",user_id="529a2f3e43954995ba768f4c20802cbb",vcpus=8i,volumes_attached=0i 1619543517000000000
> openstack_server_diagnostics,host=ubuntu,port_1=tap5546c369-29,port_2=tapd7febd6b-c1,port_3=tape68e2d08-09,server_id=316e0048-5d91-4d87-93d9-d21ee7940902 cpu0_time=74550000000,cpu1_time=9110000000,cpu2_time=11694310000000,cpu3_time=11326380000000,cpu4_time=11674290000000,cpu5_time=11348890000000,cpu6_time=11789130000000,cpu7_time=11018680000000,hda_errors=-1,hda_read=37068,hda_read_req=15,hda_write=0,hda_write_req=0,memory=33554432,memory-actual=33554432,memory-available=32779680,memory-last_update=1619543508,memory-major_fault=556,memory-minor_fault=926825,memory-rss=82420,memory-swap_in=0,memory-swap_out=0,memory-unused=30049484,memory-usable=29871836,no_of_ports=3i,port_1_rx=7397,port_1_rx_drop=0,port_1_rx_errors=0,port_1_rx_packets=94,port_1_tx=14394,port_1_tx_drop=0,port_1_tx_errors=0,port_1_tx_packets=189,port_2_rx=219402258960,port_2_rx_drop=2733983869,port_2_rx_errors=0,port_2_rx_packets=2551071664,port_2_tx=51394558670,port_2_tx_drop=0,port_2_tx_errors=0,port_2_tx_packets=597593221,port_3_rx=220065807256,port_3_rx_drop=2723979297,port_3_rx_errors=0,port_3_rx_packets=2558837778,port_3_tx=51427399260,port_3_tx_drop=0,port_3_tx_errors=0,port_3_tx_packets=597975510,server_id="316e0048-5d91-4d87-93d9-d21ee7940902",vda_errors=-1,vda_read=177386496,vda_read_req=6466,vda_write=9470976,vda_write_req=544 1619543517000000000
> openstack_hypervisor,host=ubuntu,id=5 cpu_arch="x86_64",cpu_model="model",cpu_topology_cores=20i,cpu_topology_sockets=1i,cpu_topology_threads=2i,cpu_vendor="cpu_vendor",current_workload=0i,disk_available_least=2345i,free_disk_gb=2567i,free_ram_mb=183528i,host_ip="12.12.12.12",hypervisor_hostname="compute1",hypervisor_type="QEMU",id="5",local_gb=3367i,local_gb_used=800i,memory_mb=515304i,memory_mb_used=331776i,running_vms=10i,service_disabled_reason="",service_host="compute",service_id="161",state="up",status="enabled",vcpus=80i,vcpus_used=80i,version=2012000i 1619543517000000000
> openstack_network,host=ubuntu,id=85d94590-e29d-4062-afca-8db475d02cae,tenant_id=6226db488eb446ea89564fa29d9340d0 admin_state_up=true,availability_zone_hints=0i,created_at="2020-07-27T18:16:59Z",description="",id="85d94590-e29d-4062-afca-8db475d02cae",name="test_net",project_id="6226db488eb446ea89564fa29d9340d0",shared=false,status="ACTIVE",subnets=0i,tenant_id="6226db488eb446ea89564fa29d9340d0",updated_at="2020-07-27T18:16:59Z" 1619543517000000000
> openstack_identity,host=ubuntu,id=92be484905a343c786fd94bb4112a207 description="",domain_id="default",enabled=true,id="92be484905a343c786fd94bb4112a207",is_domain=false,name="test_proj",parent_id="default",projects=6i 1619543517000000000
```
