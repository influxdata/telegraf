# CouchDB Input Plugin
---

The CouchDB plugin gathers metrics of CouchDB using [_stats](http://docs.couchdb.org/en/1.6.1/api/server/common.html?highlight=stats#get--_stats) endpoint.

### Configuration:

```
# Sample Config:
[[inputs.couchdb]]
  hosts = ["http://localhost:5984/_stats"]
```

### Measurements & Fields:

Statistics specific to the internals of CouchDB:

- couchdb_auth_cache_misses
- couchdb_database_writes
- couchdb_open_databases
- couchdb_auth_cache_hits
- couchdb_request_time
- couchdb_database_reads
- couchdb_open_os_files

Statistics of HTTP requests by method:

- httpd_request_methods_put
- httpd_request_methods_get
- httpd_request_methods_copy
- httpd_request_methods_delete
- httpd_request_methods_post
- httpd_request_methods_head

Statistics of HTTP requests by response code:

- httpd_status_codes_200
- httpd_status_codes_201
- httpd_status_codes_202
- httpd_status_codes_301
- httpd_status_codes_304
- httpd_status_codes_400
- httpd_status_codes_401
- httpd_status_codes_403
- httpd_status_codes_404
- httpd_status_codes_405
- httpd_status_codes_409
- httpd_status_codes_412
- httpd_status_codes_500

httpd statistics:

- httpd_clients_requesting_changes
- httpd_temporary_view_reads
- httpd_requests
- httpd_bulk_requests
- httpd_view_reads

### Tags:

- server (url of the couchdb _stats endpoint)

### Example output:

```
➜  telegraf git:(master) ✗ ./telegraf --config ./config.conf  --input-filter couchdb --test
* Plugin: couchdb,
 Collection 1
> couchdb,server=http://localhost:5984/_stats couchdb_auth_cache_hits_current=0,
couchdb_auth_cache_hits_max=0,
couchdb_auth_cache_hits_mean=0,
couchdb_auth_cache_hits_min=0,
couchdb_auth_cache_hits_stddev=0,
couchdb_auth_cache_hits_sum=0,
couchdb_auth_cache_misses_current=0,
couchdb_auth_cache_misses_max=0,
couchdb_auth_cache_misses_mean=0,
couchdb_auth_cache_misses_min=0,
couchdb_auth_cache_misses_stddev=0,
couchdb_auth_cache_misses_sum=0,
couchdb_database_reads_current=0,
couchdb_database_reads_max=0,
couchdb_database_reads_mean=0,
couchdb_database_reads_min=0,
couchdb_database_reads_stddev=0,
couchdb_database_reads_sum=0,
couchdb_database_writes_current=1102,
couchdb_database_writes_max=131,
couchdb_database_writes_mean=0.116,
couchdb_database_writes_min=0,
couchdb_database_writes_stddev=3.536,
couchdb_database_writes_sum=1102,
couchdb_open_databases_current=1,
couchdb_open_databases_max=1,
couchdb_open_databases_mean=0,
couchdb_open_databases_min=0,
couchdb_open_databases_stddev=0.01,
couchdb_open_databases_sum=1,
couchdb_open_os_files_current=2,
couchdb_open_os_files_max=2,
couchdb_open_os_files_mean=0,
couchdb_open_os_files_min=0,
couchdb_open_os_files_stddev=0.02,
couchdb_open_os_files_sum=2,
couchdb_request_time_current=242.21,
couchdb_request_time_max=102,
couchdb_request_time_mean=5.767,
couchdb_request_time_min=1,
couchdb_request_time_stddev=17.369,
couchdb_request_time_sum=242.21,
httpd_bulk_requests_current=0,
httpd_bulk_requests_max=0,
httpd_bulk_requests_mean=0,
httpd_bulk_requests_min=0,
httpd_bulk_requests_stddev=0,
httpd_bulk_requests_sum=0,
httpd_clients_requesting_changes_current=0,
httpd_clients_requesting_changes_max=0,
httpd_clients_requesting_changes_mean=0,
httpd_clients_requesting_changes_min=0,
httpd_clients_requesting_changes_stddev=0,
httpd_clients_requesting_changes_sum=0,
httpd_request_methods_copy_current=0,
httpd_request_methods_copy_max=0,
httpd_request_methods_copy_mean=0,
httpd_request_methods_copy_min=0,
httpd_request_methods_copy_stddev=0,
httpd_request_methods_copy_sum=0,
httpd_request_methods_delete_current=0,
httpd_request_methods_delete_max=0,
httpd_request_methods_delete_mean=0,
httpd_request_methods_delete_min=0,
httpd_request_methods_delete_stddev=0,
httpd_request_methods_delete_sum=0,
httpd_request_methods_get_current=31,
httpd_request_methods_get_max=1,
httpd_request_methods_get_mean=0.003,
httpd_request_methods_get_min=0,
httpd_request_methods_get_stddev=0.057,
httpd_request_methods_get_sum=31,
httpd_request_methods_head_current=0,
httpd_request_methods_head_max=0,
httpd_request_methods_head_mean=0,
httpd_request_methods_head_min=0,
httpd_request_methods_head_stddev=0,
httpd_request_methods_head_sum=0,
httpd_request_methods_post_current=1102,
httpd_request_methods_post_max=131,
httpd_request_methods_post_mean=0.116,
httpd_request_methods_post_min=0,
httpd_request_methods_post_stddev=3.536,
httpd_request_methods_post_sum=1102,
httpd_request_methods_put_current=1,
httpd_request_methods_put_max=1,
httpd_request_methods_put_mean=0,
httpd_request_methods_put_min=0,
httpd_request_methods_put_stddev=0.01,
httpd_request_methods_put_sum=1,
httpd_requests_current=1133,
httpd_requests_max=130,
httpd_requests_mean=0.118,
httpd_requests_min=0,
httpd_requests_stddev=3.512,
httpd_requests_sum=1133,
httpd_status_codes_200_current=31,
httpd_status_codes_200_max=1,
httpd_status_codes_200_mean=0.003,
httpd_status_codes_200_min=0,
httpd_status_codes_200_stddev=0.057,
httpd_status_codes_200_sum=31,
httpd_status_codes_201_current=1103,
httpd_status_codes_201_max=130,
httpd_status_codes_201_mean=0.116,
httpd_status_codes_201_min=0,
httpd_status_codes_201_stddev=3.532,
httpd_status_codes_201_sum=1103,
httpd_status_codes_202_current=0,
httpd_status_codes_202_max=0,
httpd_status_codes_202_mean=0,
httpd_status_codes_202_min=0,
httpd_status_codes_202_stddev=0,
httpd_status_codes_202_sum=0,
httpd_status_codes_301_current=0,
httpd_status_codes_301_max=0,
httpd_status_codes_301_mean=0,
httpd_status_codes_301_min=0,
httpd_status_codes_301_stddev=0,
httpd_status_codes_301_sum=0,
httpd_status_codes_304_current=0,
httpd_status_codes_304_max=0,
httpd_status_codes_304_mean=0,
httpd_status_codes_304_min=0,
httpd_status_codes_304_stddev=0,
httpd_status_codes_304_sum=0,
httpd_status_codes_400_current=0,
httpd_status_codes_400_max=0,
httpd_status_codes_400_mean=0,
httpd_status_codes_400_min=0,
httpd_status_codes_400_stddev=0,
httpd_status_codes_400_sum=0,
httpd_status_codes_401_current=0,
httpd_status_codes_401_max=0,
httpd_status_codes_401_mean=0,
httpd_status_codes_401_min=0,
httpd_status_codes_401_stddev=0,
httpd_status_codes_401_sum=0,
httpd_status_codes_403_current=0,
httpd_status_codes_403_max=0,
httpd_status_codes_403_mean=0,
httpd_status_codes_403_min=0,
httpd_status_codes_403_stddev=0,
httpd_status_codes_403_sum=0,
httpd_status_codes_404_current=0,
httpd_status_codes_404_max=0,
httpd_status_codes_404_mean=0,
httpd_status_codes_404_min=0,
httpd_status_codes_404_stddev=0,
httpd_status_codes_404_sum=0,
httpd_status_codes_405_current=0,
httpd_status_codes_405_max=0,
httpd_status_codes_405_mean=0,
httpd_status_codes_405_min=0,
httpd_status_codes_405_stddev=0,
httpd_status_codes_405_sum=0,
httpd_status_codes_409_current=0,
httpd_status_codes_409_max=0,
httpd_status_codes_409_mean=0,
httpd_status_codes_409_min=0,
httpd_status_codes_409_stddev=0,
httpd_status_codes_409_sum=0,
httpd_status_codes_412_current=0,
httpd_status_codes_412_max=0,
httpd_status_codes_412_mean=0,
httpd_status_codes_412_min=0,
httpd_status_codes_412_stddev=0,
httpd_status_codes_412_sum=0,
httpd_status_codes_500_current=0,
httpd_status_codes_500_max=0,
httpd_status_codes_500_mean=0,
httpd_status_codes_500_min=0,
httpd_status_codes_500_stddev=0,
httpd_status_codes_500_sum=0,
httpd_temporary_view_reads_current=0,
httpd_temporary_view_reads_max=0,
httpd_temporary_view_reads_mean=0,
httpd_temporary_view_reads_min=0,
httpd_temporary_view_reads_stddev=0,
httpd_temporary_view_reads_sum=0,
httpd_view_reads_current=0,
httpd_view_reads_max=0,
httpd_view_reads_mean=0,
httpd_view_reads_min=0,
httpd_view_reads_stddev=0,
httpd_view_reads_sum=0 1454692257621938169
```
