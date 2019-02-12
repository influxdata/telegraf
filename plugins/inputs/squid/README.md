# Squid Input Plugin

This plugin gathers metrics from a Squid web proxy cache server (http://www.squid-cache.org/).

### Configuration:

To use this plugin you must [configure the manager api to allow access](https://wiki.squid-cache.org/Features/CacheManager) from the system on which telegraf is running

```toml
# Squid web proxy cache plugin
[[inputs.squid]]
  ## url of the squid proxy manager counters page
  url = http://localhost:3128/squid-internal-mgr/counters
  ## Maximum time to receive response.
  response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

### Metrics

Metrics currently gathered come from `/squid-internal-mgr/counters`, and are all 64-bit counters.

- squid
  - tags:
    - source (url used)      
  - fields:
    - aborted_requests
    - cd_kbytes_recv
    - cd_kbytes_sent
    - cd_local_memory
    - cd_memory
    - cd_msgs_recv
    - cd_msgs_sent
    - cd_times_used
    - client_http_errors
    - client_http_hit_kbytes_out
    - client_http_hits
    - client_http_kbytes_in
    - client_http_kbytes_out
    - client_http_requests
    - cpu_time
    - icp_kbytes_recv
    - icp_kbytes_sent
    - icp_pkts_recv
    - icp_pkts_sent
    - icp_q_kbytes_recv
    - icp_q_kbytes_sent
    - icp_queries_recv
    - icp_queries_sent
    - icp_query_timeouts
    - icp_r_kbytes_recv
    - icp_r_kbytes_sent
    - icp_replies_queued
    - icp_replies_recv
    - icp_replies_sent
    - icp_times_used
    - page_faults
    - select_loops
    - server_all_errors
    - server_all_kbytes_in
    - server_all_kbytes_out
    - server_all_requests
    - server_ftp_kbytes_in
    - server_ftp_kbytes_out
    - server_ftp_requests
    - server_http_errors
    - server_http_kbytes_in
    - server_http_kbytes_out
    - server_http_requests
    - server_other_errors
    - server_other_kbytes_in
    - server_other_kbytes_out
    - server_other_requests
    - swap_files_cleaned
    - swap_ins
    - swap_outs
    - unlink_requests
    - wall_time

### Example Output:

```
squid,host=d3541b1c1112,source=http://squid:3128/squid-internal-mgr/counters wall_time=0.884082,server_http_kbytes_in=0,server_other_kbytes_out=0,icp_kbytes_sent=0,icp_q_kbytes_recv=0,icp_r_kbytes_recv=0,cd_times_used=0,server_all_errors=0,server_other_requests=0,unlink_requests=0,swap_files_cleaned=0,cpu_time=0.056241,server_all_kbytes_out=0,server_http_requests=0,icp_replies_sent=0,icp_replies_queued=0,client_http_hit_kbytes_out=0,cd_msgs_recv=0,cd_kbytes_recv=0,server_other_errors=0,icp_r_kbytes_sent=0,icp_times_used=0,cd_memory=0,client_http_requests=0,client_http_kbytes_in=0,server_http_errors=0,server_ftp_kbytes_in=0,page_faults=0,client_http_errors=0,icp_pkts_sent=0,icp_queries_sent=0,icp_query_timeouts=0,aborted_requests=0,swap_outs=0,swap_ins=0,icp_queries_recv=0,server_ftp_kbytes_out=0,icp_pkts_recv=0,icp_kbytes_recv=0,cd_msgs_sent=0,cd_local_memory=0,client_http_kbytes_out=0,server_all_kbytes_in=0,cd_kbytes_sent=0,server_http_kbytes_out=0,select_loops=2,client_http_hits=0,server_other_kbytes_in=0,server_ftp_requests=0,server_ftp_errors=0,server_all_requests=0,icp_replies_recv=0,icp_q_kbytes_sent=0 1565150341000000000
```

### Development

Run short tests:
```
go test -short
```

Run full integration tests:
```
cd plugins/inputs/squid
docker-compose -f dev/docker-compose.yml up squid
go test
```
