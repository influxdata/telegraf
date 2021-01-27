# RavenDB Input Plugin

Reads metrics from RavenDB servers via monitoring endpoints APIs.

Requires RavenDB Server 5.2+.

### Configuration

The following is an example config for RavenDB. **Note:** The client certificate used should have `Operator` permissions on the cluster.

```toml
## Global tags useful to group other metrics with the RavenDB Node metrics
[global_tags]
 cluster = "ECommerce Stage" 

[[inputs.ravendb]]
  ## Node URL and port that RavenDB is listening on.
  url = "https://localhost:8080"

  ## RavenDB X509 client certificate setup
  tls_cert = "/etc/telegraf/raven.crt"
  tls_key = "/etc/telegraf/raven.key"

  ## Optional request timeouts
  ##
  ## ResponseHeaderTimeout, if non-zero, specifies the amount of time to wait
  ## for a server's response headers after fully writing the request.
  # header_timeout = "3s"
  ##
  ## client_timeout specifies a time limit for requests made by this client.
  ## Includes connection time, any redirects, and reading the response body.
  # client_timeout = "4s"

  ## When true, collect server stats
  # gather_server_stats = true

  ## When true, collect per database stats
  # gather_db_stats = true

  ## When true, collect per index stats
  # gather_index_stats = true
  
  ## When true, collect per collection stats
  # gather_collection_stats = true

  ## List of db where database stats are collected
  ## If empty, all db are concerned
  # db_stats_dbs = []

  ## List of db where index status are collected
  ## If empty, all indexes from all db are concerned
  # index_stats_dbs = []
  
  ## List of db where collection status are collected
  ## If empty, all collections from all db are concerned
  # collection_stats_dbs = []
```

### Metrics

- ravendb_server
  - tags:
    - url
    - node_tag
    - cluster_id
    - public_server_url (optional)  
  - fields:
    - backup_current_number_of_running_backups
    - backup_max_number_of_concurrent_backups
    - certificate_server_certificate_expiration_left_in_sec (optional)
    - certificate_well_known_admin_certificates (optional, separated by ';')
    - cluster_current_term
    - cluster_index      
    - cluster_node_state
      - 0 -> Passive
      - 1 -> Candidate
      - 2 -> Follower
      - 3 -> LeaderElect
      - 4 -> Leader
    - config_public_tcp_server_urls (optional, separated by ';')
    - config_server_urls
    - config_tcp_server_urls (optional, separated by ';')
    - cpu_assigned_processor_count
    - cpu_machine_usage
    - cpu_machine_io_wait (optional)
    - cpu_process_usage
    - cpu_processor_count
    - cpu_thread_pool_available_worker_threads
    - cpu_thread_pool_available_completion_port_threads
    - databases_loaded_count
    - databases_total_count
    - disk_remaining_storage_space_percentage
    - disk_system_store_used_data_file_size_in_mb
    - disk_system_store_total_data_file_size_in_mb
    - disk_total_free_space_in_mb
    - license_expiration_left_in_sec (optional)
    - license_max_cores
    - license_type
    - license_utilized_cpu_cores
    - memory_allocated_in_mb  
    - memory_installed_in_mb
    - memory_low_memory_severity
      - 0 -> None
      - 1 -> Low
      - 2 -> Extremely Low
    - memory_physical_in_mb
    - memory_total_dirty_in_mb
    - memory_total_swap_size_in_mb
    - memory_total_swap_usage_in_mb
    - memory_working_set_swap_usage_in_mb
    - network_concurrent_requests_count
    - network_last_authorized_non_cluster_admin_request_time_in_sec (optional)
    - network_last_request_time_in_sec (optional)
    - network_requests_per_sec
    - network_tcp_active_connections
    - network_total_requests
    - server_full_version
    - server_process_id
    - server_version
    - uptime_in_sec
  
- ravendb_databases
  - tags:
    - url
    - database_name
    - database_id
    - node_tag
    - public_server_url (optional)
  - fields:
    - counts_alerts
    - counts_attachments
    - counts_documents
    - counts_performance_hints
    - counts_rehabs
    - counts_replication_factor
    - counts_revisions
    - counts_unique_attachments
    - statistics_doc_puts_per_sec
    - statistics_map_index_indexes_per_sec
    - statistics_map_reduce_index_mapped_per_sec
    - statistics_map_reduce_index_reduced_per_sec
    - statistics_request_average_duration
    - statistics_requests_count
    - statistics_requests_per_sec
    - indexes_auto_count
    - indexes_count
    - indexes_disabled_count
    - indexes_errors_count
    - indexes_errored_count
    - indexes_idle_count
    - indexes_stale_count
    - indexes_static_count
    - storage_documents_allocated_data_file_in_mb
    - storage_documents_used_data_file_in_mb
    - storage_indexes_allocated_data_file_in_mb
    - storage_indexes_used_data_file_in_mb
    - storage_total_allocated_storage_file_in_mb
    - storage_total_free_space_in_mb
    - time_since_last_backup_in_sec (optional)
    - uptime_in_sec

- ravendb_indexes
  - tags: 
    - database_name
    - index_name
    - node_tag
    - public_server_url (optional)
    - url
  - fields
    - errors
    - is_invalid
    - lock_mode
      - Unlock
      - LockedIgnore
      - LockedError
    - mapped_per_sec
    - priority
      - Low
      - Normal
      - High
    - reduced_per_sec
    - state
      - Normal
      - Disabled
      - Idle
      - Error
    - status
      - Running
      - Paused
      - Disabled
    - time_since_last_indexing_in_sec (optional)
    - time_since_last_query_in_sec (optional)
    - type
      - None
      - AutoMap
      - AutoMapReduce
      - Map
      - MapReduce
      - Faulty
      - JavaScriptMap
      - JavaScriptMapReduce

- ravendb_collections
  - tags:
    - collection_name
    - database_name
    - node_tag
    - public_server_url (optional)
    - url
  - fields
    - documents_count
    - documents_size_in_bytes
    - revisions_size_in_bytes
    - tombstones_size_in_bytes
    - total_size_in_bytes

### Contributors

- Marcin Lewandowski (https://github.com/ml054/)
- Casey Barton (https://github.com/bartoncasey)