
# OpenStack Input Plugin

Collects the metrics from following services of OpenStack:

* CINDER(Block Storage)
* GLANCE(Image service)
* HEAT(Orchestration)
* KEYSTONE(Identity service)
* NEUTRON(Networking)
* NOVA(Compute Service)

At present this plugin requires the following APIs:

* blockstorage  v2
* compute  v2
* identity  v3
* networking  v2
* orchestration  v1

## Configuration and Recommendations

### Recommendations

Due to the large number of unique tags that this plugin generates, in order to keep the cardinality down it is **highly recommended** to use [modifiers](https://github.com/influxdata/telegraf/blob/master/docs/CONFIGURATION.md#modifiers) like `tagexclude` to discard unwanted tags.

For deployments with only a small number of VMs and hosts, a small polling interval (e.g. seconds-minutes) is acceptable. For larger deployments, polling a large number of systems will impact performance. Use the `interval` option to change how often the plugin is run:

`interval`: How often a metric is gathered. Setting this value at the plugin level overrides the global agent interval setting.

Also, consider polling OpenStack services at different intervals depending on your requirements. This will help with load and cardinality as well.

```toml
[[inputs.openstack]]
  interval = 5m
  ....
  authentication_endpoint = "https://my.openstack.cloud:5000"
  ...
  enabled_services = ["nova_services"]
  ....

[[inputs.openstack]]
  interval = 30m
  ....
  authentication_endpoint = "https://my.openstack.cloud:5000"
  ...
  enabled_services = ["services", "projects", "hypervisors", "flavors", "networks", "volumes"]
  ....
```

### Configuration

```toml
# Collects performance metrics from OpenStack services
[[inputs.openstack]]
  ## The recommended interval to poll is '30m'

  ## The identity endpoint to authenticate against and get the service catalog from.
  authentication_endpoint = "https://my.openstack.cloud:5000"

  ## The domain to authenticate against when using a V3 identity endpoint.
  # domain = "default"

  ## The project to authenticate as.
  # project = "admin"

  ## User authentication credentials. Must have admin rights.
  username = "admin"
  password = "password"

  ## Available services are:
  ## "agents", "aggregates", "flavors", "hypervisors", "networks", "nova_services",
  ## "ports", "projects", "servers", "services", "stacks", "storage_pools", "subnets", "volumes"
  # enabled_services = ["services", "projects", "hypervisors", "flavors", "networks", "volumes"]

  ## Collect Server Diagnostics
  # server_diagnotics = false

  ## output secrets (such as adminPass(for server) and UserID(for volume)).
  # output_secrets = false

  ## Amount of time allowed to complete the HTTP(s) request.
  # timeout = "5s"

  ## HTTP Proxy support
  # http_proxy_url = ""

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Options for tags received from Openstack
  # tag_prefix = "openstack_tag_"
  # tag_value = "true"

  ## Timestamp format for timestamp data recieved from Openstack.
  ## If false format is unix nanoseconds.
  # human_readable_timestamps = false

  ## Measure Openstack call duration
  # measure_openstack_requests = false
```

### Measurements, Tags & Fields

* openstack_aggregate
  * name
  * aggregate_host  [string]
  * aggregate_hosts  [integer]
  * created_at  [string]
  * deleted  [boolean]
  * deleted_at  [string]
  * id  [integer]
  * updated_at  [string]
* openstack_flavor
  * is_public
  * name
  * disk  [integer]
  * ephemeral  [integer]
  * id  [string]
  * ram  [integer]
  * rxtx_factor  [float]
  * swap  [integer]
  * vcpus  [integer]
* openstack_hypervisor
  * cpu_arch
  * cpu_feature_tsc
  * cpu_feature_tsc-deadline
  * cpu_feature_tsc_adjust
  * cpu_feature_tsx-ctrl
  * cpu_feature_vme
  * cpu_feature_vmx
  * cpu_feature_x2apic
  * cpu_feature_xgetbv1
  * cpu_feature_xsave
  * cpu_model
  * cpu_vendor
  * hypervisor_hostname
  * hypervisor_type
  * hypervisor_version
  * service_host
  * service_id
  * state
  * status
  * cpu_topology_cores  [integer]
  * cpu_topology_sockets  [integer]
  * cpu_topology_threads  [integer]
  * current_workload  [integer]
  * disk_available_least  [integer]
  * free_disk_gb  [integer]
  * free_ram_mb  [integer]
  * host_ip  [string]
  * id  [string]
  * local_gb  [integer]
  * local_gb_used  [integer]
  * memory_mb  [integer]
  * memory_mb_used  [integer]
  * running_vms  [integer]
  * vcpus  [integer]
  * vcpus_used  [integer]
* openstack_identity
  * description
  * domain_id
  * name
  * parent_id
  * enabled   boolean
  * id        string
  * is_domain boolean
  * projects  integer
* openstack_network
  * name
  * openstack_tags_xyz
  * project_id
  * status
  * tenant_id
  * admin_state_up  [boolean]
  * availability_zone_hints  [string]
  * created_at  [string]
  * id  [string]
  * shared  [boolean]
  * subnet_id  [string]
  * subnets  [integer]
  * updated_at  [string]
* openstack_neutron_agent
  * agent_host
  * agent_type
  * availability_zone
  * binary
  * topic
  * admin_state_up  [boolean]
  * alive  [boolean]
  * created_at  [string]
  * heartbeat_timestamp  [string]
  * id  [string]
  * resources_synced  [boolean]
  * started_at  [string]
* openstack_nova_service
  * host_machine
  * name
  * state
  * status
  * zone
  * disabled_reason  [string]
  * forced_down  [boolean]
  * id  [string]
  * updated_at  [string]
* openstack_port
  * device_id
  * device_owner
  * name
  * network_id
  * project_id
  * status
  * tenant_id
  * admin_state_up  [boolean]
  * allowed_address_pairs  [integer]
  * fixed_ips  [integer]
  * id  [string]
  * ip_address  [string]
  * mac_address  [string]
  * security_groups  [string]
  * subnet_id  [string]
* openstack_request_duration
  * agents  [integer]
  * aggregates  [integer]
  * flavors  [integer]
  * hypervisors  [integer]
  * networks  [integer]
  * nova_services  [integer]
  * ports  [integer]
  * projects  [integer]
  * servers  [integer]
  * stacks  [integer]
  * storage_pools  [integer]
  * subnets  [integer]
  * volumes  [integer]
* openstack_server
  * flavor
  * host_id
  * host_name
  * image
  * key_name
  * name
  * project
  * status
  * tenant_id
  * user_id
  * accessIPv4  [string]
  * accessIPv6  [string]
  * addresses  [integer]
  * adminPass  [string]
  * created  [string]
  * disk_gb  [integer]
  * fault_code  [integer]
  * fault_created  [string]
  * fault_details  [string]
  * fault_message  [string]
  * id  [string]
  * progress  [integer]
  * ram_mb  [integer]
  * security_groups  [integer]
  * updated  [string]
  * vcpus  [integer]
  * volume_id  [string]
  * volumes_attached  [integer]
* openstack_server_diagnostics
  * disk_name
  * no_of_disks
  * no_of_ports
  * port_name
  * server_id
  * cpu0_time  [float]
  * cpu1_time  [float]
  * cpu2_time  [float]
  * cpu3_time  [float]
  * cpu4_time  [float]
  * cpu5_time  [float]
  * cpu6_time  [float]
  * cpu7_time  [float]
  * disk_errors  [float]
  * disk_read  [float]
  * disk_read_req  [float]
  * disk_write  [float]
  * disk_write_req  [float]
  * memory  [float]
  * memory-actual  [float]
  * memory-rss  [float]
  * memory-swap_in  [float]
  * port_rx  [float]
  * port_rx_drop  [float]
  * port_rx_errors  [float]
  * port_rx_packets  [float]
  * port_tx  [float]
  * port_tx_drop  [float]
  * port_tx_errors  [float]
  * port_tx_packets  [float]
* openstack_service
  * name
  * service_enabled  [boolean]
  * service_id  [string]
* openstack_storage_pool
  * driver_version
  * name
  * storage_protocol
  * vendor_name
  * volume_backend_name
  * free_capacity_gb  [float]
  * total_capacity_gb  [float]
* openstack_subnet
  * cidr
  * gateway_ip
  * ip_version
  * name
  * network_id
  * openstack_tags_subnet_type_PRV
  * project_id
  * tenant_id
  * allocation_pools  [string]
  * dhcp_enabled  [boolean]
  * dns_nameservers  [string]
  * id  [string]
* openstack_volume
  * attachment_attachment_id
  * attachment_device
  * attachment_host_name
  * availability_zone
  * bootable
  * description
  * name
  * status
  * user_id
  * volume_type
  * attachment_attached_at  [string]
  * attachment_server_id  [string]
  * created_at  [string]
  * encrypted  [boolean]
  * id  [string]
  * multiattach  [boolean]
  * size  [integer]
  * total_attachments  [integer]
  * updated_at  [string]

### Example Output

```text
> openstack_neutron_agent,agent_host=vim2,agent_type=DHCP\ agent,availability_zone=nova,binary=neutron-dhcp-agent,host=telegraf_host,topic=dhcp_agent admin_state_up=true,alive=true,created_at="2021-01-07T03:40:53Z",heartbeat_timestamp="2021-10-14T07:46:40Z",id="17e1e446-d7da-4656-9e32-67d3690a306f",resources_synced=false,started_at="2021-07-02T21:47:42Z" 1634197616000000000
> openstack_aggregate,host=telegraf_host,name=non-dpdk aggregate_host="vim3",aggregate_hosts=2i,created_at="2021-02-01T18:28:00Z",deleted=false,deleted_at="0001-01-01T00:00:00Z",id=3i,updated_at="0001-01-01T00:00:00Z" 1634197617000000000
> openstack_flavor,host=telegraf_host,is_public=true,name=hwflavor disk=20i,ephemeral=0i,id="f89785c0-6b9f-47f5-a02e-f0fcbb223163",ram=8192i,rxtx_factor=1,swap=0i,vcpus=8i 1634197617000000000
> openstack_hypervisor,cpu_arch=x86_64,cpu_feature_3dnowprefetch=true,cpu_feature_abm=true,cpu_feature_acpi=true,cpu_feature_adx=true,cpu_feature_aes=true,cpu_feature_apic=true,cpu_feature_xtpr=true,cpu_model=C-Server,cpu_vendor=xyz,host=telegraf_host,hypervisor_hostname=vim3,hypervisor_type=QEMU,hypervisor_version=4002000,service_host=vim3,service_id=192,state=up,status=enabled cpu_topology_cores=28i,cpu_topology_sockets=1i,cpu_topology_threads=2i,current_workload=0i,disk_available_least=2596i,free_disk_gb=2744i,free_ram_mb=374092i,host_ip="xx:xx:xx:x::xxx",id="12",local_gb=3366i,local_gb_used=622i,memory_mb=515404i,memory_mb_used=141312i,running_vms=15i,vcpus=0i,vcpus_used=72i 1634197618000000000
> openstack_network,host=telegraf_host,name=Network\ 2,project_id=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,status=active,tenant_id=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx admin_state_up=true,availability_zone_hints="",created_at="2021-07-29T15:58:25Z",id="f5af5e71-e890-4245-a377-d4d86273c319",shared=false,subnet_id="2f7341c6-074d-42aa-9abc-71c662d9b336",subnets=1i,updated_at="2021-09-02T16:46:48Z" 1634197618000000000
> openstack_nova_service,host=telegraf_host,host_machine=vim3,name=nova-compute,state=up,status=enabled,zone=nova disabled_reason="",forced_down=false,id="192",updated_at="2021-10-14T07:46:52Z" 1634197619000000000
> openstack_port,device_id=a043b8b3-2831-462a-bba8-19088f3db45a,device_owner=compute:nova,host=telegraf_host,name=offload-port1,network_id=6b40d744-9a48-43f2-a4c8-2e0ccb45ac96,project_id=71f9bc44621234f8af99a3949258fc7b,status=ACTIVE,tenant_id=71f9bc44621234f8af99a3949258fc7b admin_state_up=true,allowed_address_pairs=0i,fixed_ips=1i,id="fb64626a-07e1-4d78-a70d-900e989537cc",ip_address="1.1.1.5",mac_address="xx:xx:xx:xx:xx:xx",security_groups="",subnet_id="eafa1eca-b318-4746-a55a-682478466689" 1634197620000000000
> openstack_identity,domain_id=default,host=telegraf_host,name=service,parent_id=default enabled=true,id="a0877dd2ed1d4b5f952f5689bc04b0cb",is_domain=false,projects=7i 1634197621000000000
> openstack_server,flavor=0d438971-56cf-4f86-801f-7b04b29384cb,host=telegraf_host,host_id=c0fe05b14261d35cf8748a3f5aae1234b88c2fd62b69fe24ca4a27e9,host_name=vim1,image=b295f1f3-1w23-470c-8734-197676eedd16,name=test-VM7,project=admin,status=active,tenant_id=80ac889731f540498fb1dc78e4bcd5ed,user_id=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx accessIPv4="",accessIPv6="",addresses=1i,adminPass="",created="2021-09-07T14:40:11Z",disk_gb=8i,fault_code=0i,fault_created="0001-01-01T00:00:00Z",fault_details="",fault_message="",id="db92ee0d-459b-458e-9fe3-2be5ec7c87e1",progress=0i,ram_mb=16384i,security_groups=1i,updated="2021-09-07T14:40:19Z",vcpus=4i,volumes_attached=0i 1634197656000000000
> openstack_service,host=telegraf_host,name=identity service_enabled=true,service_id="ad605eff92444a158d0f78768f2c4668" 1634197656000000000
> openstack_storage_pool,driver_version=1.0.0,host=telegraf_host,name=storage_bloack_1,storage_protocol=nfs,vendor_name=xyz,volume_backend_name=abc free_capacity_gb=4847.54,total_capacity_gb=4864 1634197658000000000
> openstack_subnet,cidr=10.10.20.10/28,gateway_ip=10.10.20.17,host=telegraf_host,ip_version=4,name=IPv4_Subnet_2,network_id=73c6e1d3-f522-4a3f-8e3c-762a0c06d68b,openstack_tags_lab=True,project_id=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,tenant_id=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx allocation_pools="10.10.20.11-10.10.20.30",dhcp_enabled=true,dns_nameservers="",id="db69fbb2-9ca1-4370-8c78-82a27951c94b" 1634197660000000000
> openstack_volume,attachment_attachment_id=c83ca0d6-c467-44a0-ac1f-f87d769c0c65,attachment_device=/dev/vda,attachment_host_name=vim1,availability_zone=nova,bootable=true,host=telegraf_host,status=in-use,user_id=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx,volume_type=storage_bloack_1 attachment_attached_at="2021-01-12T21:02:04Z",attachment_server_id="c0c6b4af-0d26-4a0b-a6b4-4ea41fa3bb4a",created_at="2021-01-12T21:01:47Z",encrypted=false,id="d4204f1b-b1ae-1233-b25c-a57d91d2846e",multiattach=false,size=80i,total_attachments=1i,updated_at="2021-01-12T21:02:04Z" 1634197660000000000
> openstack_request_duration,host=telegraf_host networks=703214354i 1634197660000000000
> openstack_server_diagnostics,disk_name=vda,host=telegraf_host,no_of_disks=1,no_of_ports=2,port_name=vhu1234566c-9c,server_id=fdddb58c-bbb9-1234-894b-7ae140178909 cpu0_time=4924220000000,cpu1_time=218809610000000,cpu2_time=218624300000000,cpu3_time=220505700000000,disk_errors=-1,disk_read=619156992,disk_read_req=35423,disk_write=8432728064,disk_write_req=882445,memory=8388608,memory-actual=8388608,memory-rss=37276,memory-swap_in=0,port_rx=410516469288,port_rx_drop=13373626,port_rx_errors=-1,port_rx_packets=52140392,port_tx=417312195654,port_tx_drop=0,port_tx_errors=0,port_tx_packets=321385978 1634197660000000000
```
