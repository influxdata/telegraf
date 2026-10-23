# SeaweedFS Input Plugin

The SeaweedFS plugin gathers metrics about volumes, disk usage and collections
of a [SeaweedFS][seaweedfs] cluster. It uses the `status` HTTP APIs of the
master and of the volume servers and therefore does not require the Prometheus
endpoints to be enabled.

⭐ Telegraf v1.41.0
🏷️ server
💻 all

[seaweedfs]: https://github.com/seaweedfs/seaweedfs

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read volume, disk and collection metrics from a SeaweedFS cluster
[[inputs.seaweedfs]]
  ## Master servers of the cluster.
  ## Every server is queried for its leader state, the topology is read from
  ## the leader. At least one master is required.
  servers = ["http://localhost:9333"]

  ## Volume servers to query for disk usage and volume details.
  ## If left empty, the volume servers are discovered via the master topology.
  ## Discovery requires the "Url" values reported by the master to be
  ## resolvable from the host running Telegraf.
  ##   example: volume_servers = ["http://localhost:8080"]
  # volume_servers = []

  ## Emit one metric per volume instead of only per-collection aggregates.
  ## This creates a series per volume and can be a lot on large clusters.
  # detailed_volumes = false

  ## HTTP response timeout (default: 5s)
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Master and volume servers

The plugin reads the leader state from `/cluster/status` of every configured
master and takes the cluster topology from `/dir/status` of the leader. The
volume servers reported by the topology are queried at `/status` for the disk
usage and the volume inventory. The volume servers reported by the master are
plain `host:port` addresses, so they must be resolvable from the host running
Telegraf. Set `volume_servers` to override the discovery, for example when
Telegraf runs outside of the cluster and accesses the servers via a port
forward.

All metrics are derived from read-only endpoints. If the cluster runs with
[JWT authentication][security] the status endpoints are only available to
clients on the master's `white_list`, so add the Telegraf host there.

[security]: https://github.com/seaweedfs/seaweedfs/wiki/Security

## Metrics

- seaweedfs_cluster
  - tags:
    - server - Address of the master the topology was read from
  - fields:
    - volume_count (integer) - Number of volumes across all volume servers
    - ec_shard_count (integer) - Number of erasure coded shards
    - max_volume_slots (integer) - Maximum number of volume slots of the cluster
    - free_volume_slots (integer) - Unused volume slots
    - volume_server_count (integer) - Number of volume servers
    - data_center_count (integer) - Number of data centers
    - rack_count (integer) - Number of racks
    - volume_layout_count (integer) - Number of volume layouts

- seaweedfs_master
  - tags:
    - server - Address of the master
  - fields:
    - is_leader (boolean) - Whether this master is the leader
    - max_volume_id (integer) - Highest volume ID assigned by the master

- seaweedfs_volume_server
  - tags:
    - server - Address of the volume server
    - datacenter - Data center the volume server belongs to
    - rack - Rack the volume server belongs to
  - fields:
    - volume_count (integer) - Number of volumes stored on the server
    - ec_shard_count (integer) - Number of erasure coded shards stored on the server
    - max_volume_slots (integer) - Maximum number of volume slots of the server
    - free_volume_slots (integer) - Unused volume slots of the server
    - disk_count (integer) - Number of disks reported by the server, `0` if the
      server could not be reached

- seaweedfs_disk
  - tags:
    - server - Address of the volume server
    - path - Directory of the disk
    - disk_type - Type of the disk, if reported by the server
  - fields:
    - total_bytes (integer) - Total size of the disk
    - used_bytes (integer) - Used size of the disk
    - free_bytes (integer) - Free size of the disk
    - used_percent (float) - Used percentage of the disk
    - free_percent (float) - Free percentage of the disk
    - volume_count (integer) - Number of volumes stored on the disk

- seaweedfs_collection
  - tags:
    - collection - Collection name, volumes without a collection are reported
      as `default`
  - fields:
    - volume_count (integer) - Number of volumes of the collection
    - readonly_volume_count (integer) - Number of read-only volumes
    - writable_volume_count (integer) - Number of volumes the master can write to
    - size_bytes (integer) - Size of all volumes of the collection
    - file_count (integer) - Number of files in the collection
    - delete_count (integer) - Number of deleted files that are not vacuumed yet
    - deleted_bytes (integer) - Size of the deleted files not vacuumed yet

- seaweedfs_volume
  - tags:
    - server - Address of the volume server
    - collection - Collection name
    - volume_id - ID of the volume
  - fields:
    - size_bytes (integer) - Size of the volume
    - file_count (integer) - Number of files in the volume
    - delete_count (integer) - Number of deleted files not vacuumed yet
    - deleted_bytes (integer) - Size of the deleted files not vacuumed yet
    - read_only (boolean) - Whether the volume is read-only
    - version (integer) - Version of the volume
    - disk_id - ID of the disk the volume is stored on

The `seaweedfs_volume` metric is only emitted if `detailed_volumes` is enabled.
The `volume_count` of a collection counts every replica, so with replication it
is a multiple of the number of logical volumes.

The SeaweedFS components also expose Prometheus metrics on their metrics port,
covering for example filer and S3 request counts. Those can be collected with
the [prometheus input plugin][prometheus] if required.

[prometheus]: ../../../plugins/inputs/prometheus/README.md

## Example Output

```text
seaweedfs_master,host=telegraf,server=http://seaweedfs-master.s3.svc.cluster.local:9333 is_leader=true,max_volume_id=14i 1792782535000000000
seaweedfs_cluster,host=telegraf,server=http://seaweedfs-master.s3.svc.cluster.local:9333 data_center_count=1i,ec_shard_count=0i,free_volume_slots=21i,max_volume_slots=35i,rack_count=1i,volume_count=14i,volume_layout_count=2i,volume_server_count=1i 1792782535000000000
seaweedfs_volume_server,datacenter=DefaultDataCenter,host=telegraf,rack=DefaultRack,server=http://seaweedfs-volume-0.seaweedfs-volume.s3:8080 disk_count=1i,ec_shard_count=0i,free_volume_slots=21i,max_volume_slots=35i,volume_count=14i 1792782535000000000
seaweedfs_disk,host=telegraf,path=/data1,server=http://seaweedfs-volume-0.seaweedfs-volume.s3:8080 free_bytes=408165601280i,free_percent=83.27107,total_bytes=490164953088i,used_bytes=81999351808i,used_percent=16.72893,volume_count=14i 1792782535000000000
seaweedfs_collection,collection=default,host=telegraf delete_count=0i,deleted_bytes=0i,file_count=73i,readonly_volume_count=0i,size_bytes=44356832i,volume_count=7i,writable_volume_count=7i 1792782535000000000
seaweedfs_collection,collection=lab,host=telegraf delete_count=0i,deleted_bytes=0i,file_count=13645i,readonly_volume_count=0i,size_bytes=3102462751i,volume_count=7i,writable_volume_count=7i 1792782535000000000
```
