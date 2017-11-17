# VMware vSphere Input Plugin

The VMware vSphere plugin uses the vSphere API to gather metrics from multiple vCenters.
 
* Clusters
* Hosts
* VMs
* Data stores

## Configuration

```
# Read metrics from one or many vCenters
[[inputs.vsphere]]
  ## List of vCenter URLs, including credentials.
  ## <protocol>://<user>:<password>@<vcenter>/sdk
  ## Note the "@" characted must be escaped as %40
  ## e.g.
  ##     https://administrator%40vsphere.local:password@vcenter.local/sdk
  #
  vcenters = [ "https://administrator%40vsphere.local:password@vcenter.local/sdk" ]
  
  ## gather metrics from Clusters (default: true)
  gather_clusters = true
  
  ## gather metrics from Hosts (default: true)
  gather_hosts = true
  
  ## gather metrics from VMs (default: true)
  gather_vms = true
  
  ## gather metrics from Data stores (default: true)
  gather_datastores = true
  
  ## Number of objects to retreive per query (default: 500)
  objects_per_query = 500
  
  ## Object Discover duration is the interval to wait between collection cycles to (re)discover objects subject to gathering metrics (default: 300s)
  object_discover_interval = 300s
    
  ## Optional: if you have customized the default sampling period in vcenter you can specify them below
  cluster_sampling_period = 300s
  host_sampling_period = 20s
  vm_sampling_period = 20s
  datastore_sampling_period = 300s
```
