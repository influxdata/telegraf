# SeaweedFS Input Plugin

The SeaweedFS plugin gathers the part of a [SeaweedFS][seaweedfs] cluster's
inventory that the SeaweedFS Prometheus endpoint does not expose: the number of
files (needles) per collection and, optionally, the state of every single
volume — files, deleted bytes, read-only flag.

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
# Read the SeaweedFS inventory that its Prometheus endpoint does not expose
[[inputs.seaweedfs]]
  ## Master servers of the cluster.
  ## Every server is queried for its leader state, the topology is read from
  ## the leader. At least one master is required.
  servers = ["http://localhost:9333"]

  ## Volume servers to query for the volume inventory.
  ## If left empty, the volume servers are discovered via the master topology.
  ## Discovery requires the "Url" values reported by the master to be
  ## resolvable from the host running Telegraf.
  ##   example: volume_servers = ["http://localhost:8080"]
  # volume_servers = []

  ## Emit one metric per volume instead of only per-collection counts.
  ## This creates a series per volume and can be a lot on large clusters.
  # detailed_volumes = false

  ## HTTP client timeout, applies to every request (default: 5s)
  # timeout = "5s"

  ## Use the proxy settings of the environment (`HTTP_PROXY`, `HTTPS_PROXY` and
  ## `NO_PROXY`) for all requests
  # use_system_proxy = false
  ## Optional HTTP proxy URL, takes precedence over the environment
  # http_proxy_url = ""

  ## Optional local address to bind the client to
  # local_address = ""

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
volume servers reported by the topology are queried at `/status` for the volume
inventory. The volume servers reported by the master are plain `host:port`
addresses, so they must be resolvable from the host running Telegraf. Set
`volume_servers` to override the discovery, for example when Telegraf runs
outside of the cluster and accesses the servers via a port forward.

All metrics are derived from read-only endpoints, the same ones the admin UI
uses, so no metrics port has to be enabled. If the cluster runs with
[JWT authentication][security] the status endpoints are only available to
clients on the master's `white_list`, so add the Telegraf host there.

[security]: https://github.com/seaweedfs/seaweedfs/wiki/Security

### Relation to the Prometheus endpoint

SeaweedFS exposes a lot of cluster state on its metrics port, and that is the
cheapest way to collect it: the [prometheus input plugin][prometheus] covers the
capacity side with no additional code. This plugin therefore only emits what the
exposition does not have.

| Field | Prometheus metric (SeaweedFS 4.x, `SeaweedFS_` prefix) |
| --- | --- |
| disk total, used, free, percent | `volumeServer_resource{name,type}` |
| volumes per collection and server | `volumeServer_volumes{collection}` |
| maximum volume slots | `volumeServer_max_volumes` |
| per-collection size and trash | `volumeServer_total_disk_size{collection,type="normal"\|"deleted_bytes"}` |
| read-only volumes | `volumeServer_read_only_volumes{collection,type}` |
| writable volumes | `master_volume_layout_writable{collection}` |
| master leader state | `master_is_leader` |

Not exposed, and the reason for this plugin:

- **File (needle) counts.** No per-collection or per-volume file count exists;
  the only `file_*` series are read/write failure counters.
- **Per-volume state** (`detailed_volumes`). No series carries a volume ID, so
  "which volume is full, how much of it is trash, which are read-only" cannot be
  answered from the exposition.
- **Delete counts** per collection and per volume.
- **Topology structure** — number of servers, racks, data centers and layouts.
  The layout series are per collection and carry no rack or data-center labels.
- **Free volume slots on erasure-coded clusters.** They look derivable as
  `max_volumes - volumes`, but the master charges every EC shard
  `ceil(shards / 10)` of a slot, so that subtraction over-reports free slots.

To have the capacity metrics in the same database, add a second input:

```toml
[[inputs.prometheus]]
  urls = ["http://seaweedfs-master:9327/metrics", "http://seaweedfs-volume-0:9327/metrics"]
```

[prometheus]: ../../../plugins/inputs/prometheus/README.md

## Metrics

- seaweedfs_cluster
  - tags:
    - server - Address of the master the topology was read from
  - fields:
    - max_volume_id (integer) - Highest volume ID assigned by the master
    - ec_shard_count (integer) - Number of erasure coded shards
    - free_volume_slots (integer) - Unused volume slots, charging every EC shard
      as a fraction of a slot the way the master does
    - volume_server_count (integer) - Number of volume servers
    - data_center_count (integer) - Number of data centers
    - rack_count (integer) - Number of racks
    - volume_layout_count (integer) - Number of volume layouts

- seaweedfs_volume_server
  - tags:
    - server - Address of the volume server
    - datacenter - Data center the volume server belongs to
    - rack - Rack the volume server belongs to
  - fields:
    - ec_shard_count (integer) - Number of erasure coded shards stored on the
      server
    - free_volume_slots (integer) - Unused volume slots of the server, only
      reported when the master knows the server
    - disk_count (integer) - Number of disks reported by the server, `0` if the
      server could not be reached

- seaweedfs_disk
  - tags:
    - server - Address of the volume server
    - path - Directory of the disk
    - disk_type - Type of the disk, if reported by the server
  - fields:
    - volume_count (integer) - Number of volumes stored on the disk

- seaweedfs_collection
  - tags:
    - server - Address of the master the topology was read from
    - collection - Collection name, volumes without a collection are reported
      as `default`
  - fields:
    - file_count (integer) - Number of files in the collection
    - delete_count (integer) - Number of deleted files that are not vacuumed yet

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

The collection counts are logical: a volume ID is unique within a cluster and
shared by all replicas of that volume, so the replicas are counted once. The
per-volume metric is physical — every volume server reports the volumes it
holds, so with replication there is one `seaweedfs_volume` series per replica,
distinguished by the `server` tag.

## Example Output

```text
seaweedfs_cluster,host=telegraf,server=http://seaweedfs-master.s3.svc.cluster.local:9333 data_center_count=1i,ec_shard_count=0i,free_volume_slots=1265i,max_volume_id=518i,rack_count=1i,volume_layout_count=2i,volume_server_count=1i 1792782535000000000
seaweedfs_volume_server,datacenter=DefaultDataCenter,host=telegraf,rack=DefaultRack,server=http://seaweedfs-volume-0.seaweedfs-volume.s3:8080 disk_count=2i,ec_shard_count=0i,free_volume_slots=1265i 1792782535000000000
seaweedfs_disk,host=telegraf,path=/data1,server=http://seaweedfs-volume-0.seaweedfs-volume.s3:8080 volume_count=220i 1792782535000000000
seaweedfs_disk,host=telegraf,path=/misc,server=http://seaweedfs-volume-0.seaweedfs-volume.s3:8080 volume_count=235i 1792782535000000000
seaweedfs_collection,collection=default,host=telegraf,server=http://seaweedfs-master.s3.svc.cluster.local:9333 delete_count=0i,file_count=73i 1792782535000000000
seaweedfs_collection,collection=lab,host=telegraf,server=http://seaweedfs-master.s3.svc.cluster.local:9333 delete_count=0i,file_count=911938i 1792782535000000000
```
