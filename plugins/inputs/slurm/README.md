# SLURM Input Plugin

This plugin gather diag, jobs, nodes, partitions and reservation metrics by
leveraging SLURM's REST API as provided by the `slurmrestd` daemon.

This plugin targets the `openapi/v0.0.38` OpenAPI plugin as defined in SLURM's
documentation. That particular plugin should be configured when starting the
`slurmrestd` daemon up. For more information, be sure to check SLURM's
documentation [here][SLURM Doc].

A great wealth of information can also be found on the repository of the
Go module implementing the API client, [pcolladosoto/goslurm][].

[SLURM Doc]: https://slurm.schedmd.com/rest.html
[pcolladosoto/goslurm]: https://github.com/pcolladosoto/goslurm

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Gather SLURM metrics
[[inputs.slurm]]
  ## Slurmrestd URL. Both http and https can be used as schemas.
  url = "http://127.0.0.1:6820"

  ## Credentials for JWT-based authentication.
  # username = "foo"
  # token = "topSecret"

  ## Ignored endpoints
  ## List of endpoints a user can ignore, choose from: diag, jobs,
  ##   nodes, partitions, reservations
  ## Please note incorrect endpoints will be silently ignore and
  ## that endpoint names are case insensitive.
  # ignored_endpoints = []

  ## Maximum time to receive response.
  # response_timeout = "5s"

  ## Optional TLS Config. Note these options will only
  ## be taken into account when the scheme specififed on
  ## the URL parameter is https. They will be silently
  ## ignored otherwise.
  ## Set to true/false to enforce TLS being enabled/disabled. If not set,
  ## enable TLS only if any of the other options are specified.
  # tls_enable =
  ## Trusted root certificates for server
  # tls_ca = "/path/to/cafile"
  ## Used for TLS client certificate authentication
  # tls_cert = "/path/to/certfile"
  ## Used for TLS client certificate authentication
  # tls_key = "/path/to/keyfile"
  ## Password for the key file if it is encrypted
  # tls_key_pwd = ""
  ## Send the specified TLS server name via SNI
  # tls_server_name = "kubernetes.example.com"
  ## Minimal TLS version to accept by the client
  # tls_min_version = "TLS12"
  ## List of ciphers to accept, by default all secure ciphers will be accepted
  ## See https://pkg.go.dev/crypto/tls#pkg-constants for supported values.
  ## Use "all", "secure" and "insecure" to add all support ciphers, secure
  ## suites or insecure suites respectively.
  # tls_cipher_suites = ["secure"]
  ## Renegotiation method, "never", "once" or "freely"
  # tls_renegotiation_method = "never"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Metrics

Given the great deal of metrics offered by SLURM's API, an attempt has been
done to strike a balance between verbosity and usefulness in terms of the
gathered information.

- slurm_diag
  - tags:
    - source
  - fields:
    - server_thread_count
    - jobs_canceled
    - jobs_submitted
    - jobs_started
    - jobs_completed
    - jobs_failed
    - jobs_pending
    - jobs_running
    - schedule_cycle_last
    - schedule_cycle_mean
    - bf_queue_len
    - bf_queue_len_mean
    - bf_active
- slurm_jobs
  - tags:
    - source
    - name
    - job_id
  - fields:
    - state
    - state_reason
    - partition
    - nodes
    - node_count
    - priority
    - nice
    - group_id
    - command
    - standard_output
    - standard_error
    - standard_input
    - current_working_directory
    - submit_time
    - start_time
    - cpus
    - tasks
    - time_limit
    - tres_cpu
    - tres_mem
    - tres_node
    - tres_billing
- slurm_nodes
  - tags:
    - source
    - name
  - fields:
    - state
    - cores
    - cpus
    - cpu_load
    - alloc_cpu
    - real_memory
    - free_memory
    - alloc_memory
    - tres_cpu
    - tres_mem
    - tres_billing
    - tres_used_cpu
    - tres_used_mem
    - weight
    - slurmd_version
    - architecture
- slurm_partitions
  - tags:
    - source
    - name
  - fields:
    - state
    - total_cpu
    - total_nodes
    - nodes
    - tres_cpu
    - tres_mem
    - tres_node
    - tres_billing
- slurm_reservations
  - tags:
    - source
    - name
  - fields:
    - core_count
    - core_spec_count
    - groups
    - users
    - start_time
    - partition
    - accounts
    - node_count
    - node_list

## Example Output

```text
slurm_diag,host=hoth,source=slurm_primary.example.net bf_active=false,bf_queue_len=10i,bf_queue_len_mean=6i,jobs_canceled=0i,jobs_completed=222i,jobs_failed=0i,jobs_pending=10i,jobs_running=90i,jobs_started=212i,jobs_submitted=222i,schedule_cycle_last=234i,schedule_cycle_mean=111i,server_thread_count=3i 1723039486000000000
slurm_jobs,host=hoth,job_id=19711,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.9UaD7w",cpus=2i,current_working_directory="/home/sessiondir/pCNKDmta1u5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmoGJKDmgGh5ym",group_id=2005i,nice=50i,node_count=1i,nodes="naboo225",partition="atlas",priority=4294882018i,standard_error="/home/sessiondir/pCNKDmta1u5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmoGJKDmgGh5ym.comment",standard_input="/dev/null",standard_output="/home/sessiondir/pCNKDmta1u5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmoGJKDmgGh5ym.comment",start_time=1722845496i,state="RUNNING",state_reason="None",submit_time=1722845495i,tasks=1i,time_limit=3600i,tres_billing=1i,tres_cpu=1i,tres_mem=2000i,tres_node=1i 1723039486000000000
slurm_jobs,host=hoth,job_id=19716,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.YTCl3d",cpus=2i,current_working_directory="/home/sessiondir/JjMKDmQp1u5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmuGJKDm6k5srn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo228",partition="atlas",priority=4294882013i,standard_error="/home/sessiondir/JjMKDmQp1u5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmuGJKDm6k5srn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/JjMKDmQp1u5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmuGJKDm6k5srn.comment",start_time=1722846057i,state="RUNNING",state_reason="None",submit_time=1722846056i,tasks=1i,time_limit=3600i,tres_billing=1i,tres_cpu=1i,tres_mem=2000i,tres_node=1i 1723039486000000000
slurm_jobs,host=hoth,job_id=20026,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.KtR10Q",cpus=8i,current_working_directory="/home/sessiondir/9K8MDmtLBv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm8LJKDm6NIrPm",group_id=2005i,nice=50i,node_count=1i,nodes="naboo229",partition="atlas",priority=4294881703i,standard_error="/home/sessiondir/9K8MDmtLBv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm8LJKDm6NIrPm.comment",standard_input="/dev/null",standard_output="/home/sessiondir/9K8MDmtLBv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm8LJKDm6NIrPm.comment",start_time=1722911483i,state="RUNNING",state_reason="None",submit_time=1722911483i,tasks=8i,time_limit=3600i,tres_billing=8i,tres_cpu=8i,tres_mem=8000i,tres_node=1i 1723039486000000000
slurm_jobs,host=hoth,job_id=20129,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.k3ro05",cpus=8i,current_working_directory="/home/sessiondir/jH2LDm0FLv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmPNJKDmLay4Wn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo225",partition="atlas",priority=4294881600i,standard_error="/home/sessiondir/jH2LDm0FLv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmPNJKDmLay4Wn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/jH2LDm0FLv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmPNJKDmLay4Wn.comment",start_time=1722929192i,state="RUNNING",state_reason="None",submit_time=1722929192i,tasks=8i,time_limit=3600i,tres_billing=8i,tres_cpu=8i,tres_mem=16000i,tres_node=1i 1723039486000000000
slurm_jobs,host=hoth,job_id=20157,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.KyHNKg",cpus=8i,current_working_directory="/home/sessiondir/A0uLDm2DNv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmlNJKDm4B1ZWn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo219",partition="atlas",priority=4294881572i,standard_error="/home/sessiondir/A0uLDm2DNv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmlNJKDm4B1ZWn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/A0uLDm2DNv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmlNJKDm4B1ZWn.comment",start_time=1722931476i,state="RUNNING",state_reason="None",submit_time=1722931476i,tasks=8i,time_limit=3600i,tres_billing=8i,tres_cpu=8i,tres_mem=8000i,tres_node=1i 1723039486000000000
slurm_jobs,host=hoth,job_id=20217,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.VDyJ7D",cpus=8i,current_working_directory="/home/sessiondir/61rLDmLISv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmdOJKDmsnP3hn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo224",partition="atlas",priority=4294881512i,standard_error="/home/sessiondir/61rLDmLISv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmdOJKDmsnP3hn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/61rLDmLISv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmdOJKDmsnP3hn.comment",start_time=1722944978i,state="RUNNING",state_reason="None",submit_time=1722944978i,tasks=8i,time_limit=3600i,tres_billing=8i,tres_cpu=8i,tres_mem=8000i,tres_node=1i 1723039486000000000
slurm_jobs,host=hoth,job_id=20347,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.M0uebV",cpus=1i,current_working_directory="/home/sessiondir/RzvLDm09Xv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmfQJKDmPhjxWm",group_id=2005i,nice=50i,node_count=1i,nodes="naboo145",partition="atlas",priority=4294881382i,standard_error="/home/sessiondir/RzvLDm09Xv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmfQJKDmPhjxWm.comment",standard_input="/dev/null",standard_output="/home/sessiondir/RzvLDm09Xv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmfQJKDmPhjxWm.comment",start_time=1722966333i,state="RUNNING",state_reason="None",submit_time=1722966333i,tasks=1i,time_limit=3600i,tres_billing=1i,tres_cpu=1i,tres_mem=2000i,tres_node=1i 1723039486000000000
slurm_jobs,host=hoth,job_id=20356,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.kbtpdR",cpus=1i,current_working_directory="/home/sessiondir/YaVLDmOUYv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmmQJKDmKQXP7m",group_id=2005i,nice=50i,node_count=1i,nodes="naboo145",partition="atlas",priority=4294881373i,standard_error="/home/sessiondir/YaVLDmOUYv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmmQJKDmKQXP7m.comment",standard_input="/dev/null",standard_output="/home/sessiondir/YaVLDmOUYv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmmQJKDmKQXP7m.comment",start_time=1722967294i,state="RUNNING",state_reason="None",submit_time=1722967294i,tasks=1i,time_limit=3600i,tres_billing=1i,tres_cpu=1i,tres_mem=2000i,tres_node=1i 1723039486000000000
slurm_jobs,host=hoth,job_id=20359,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.AyLDSr",cpus=1i,current_working_directory="/home/sessiondir/JjJLDmwiYv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmpQJKDmHsppVn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo145",partition="atlas",priority=4294881370i,standard_error="/home/sessiondir/JjJLDmwiYv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmpQJKDmHsppVn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/JjJLDmwiYv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmpQJKDmHsppVn.comment",start_time=1722968537i,state="RUNNING",state_reason="None",submit_time=1722968536i,tasks=1i,time_limit=3600i,tres_billing=1i,tres_cpu=1i,tres_mem=2000i,tres_node=1i 1723039486000000000
slurm_jobs,host=hoth,job_id=20371,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.5sZMbX",cpus=1i,current_working_directory="/home/sessiondir/6Y8LDm38Yv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm1QJKDmxvi0So",group_id=2005i,nice=50i,node_count=1i,nodes="naboo145",partition="atlas",priority=4294881358i,standard_error="/home/sessiondir/6Y8LDm38Yv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm1QJKDmxvi0So.comment",standard_input="/dev/null",standard_output="/home/sessiondir/6Y8LDm38Yv5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm1QJKDmxvi0So.comment",start_time=1722969418i,state="RUNNING",state_reason="None",submit_time=1722969417i,tasks=1i,time_limit=3600i,tres_billing=1i,tres_cpu=1i,tres_mem=2000i,tres_node=1i 1723039486000000000
slurm_nodes,host=hoth,name=naboo145,source=slurm_primary.example.net alloc_cpu=32i,alloc_memory=48000i,architecture="x86_64",cores=18i,cpu_load=2707i,cpus=36i,free_memory=18745i,real_memory=94791i,slurmd_version="22.05.9",state="mixed",tres_billing=36i,tres_cpu=36i,tres_mem=94791i,tres_used_cpu=32i,tres_used_mem=48000i,weight=1i 1723039486000000000
slurm_nodes,host=hoth,name=naboo146,source=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=18i,cpu_load=0i,cpus=36i,free_memory=92167i,real_memory=94791i,slurmd_version="22.05.9",state="idle",tres_billing=36i,tres_cpu=36i,tres_mem=94791i,weight=1i 1723039486000000000
slurm_nodes,host=hoth,name=naboo147,source=slurm_primary.example.net alloc_cpu=33i,alloc_memory=50000i,architecture="x86_64",cores=18i,cpu_load=2174i,cpus=36i,free_memory=10837i,real_memory=94793i,slurmd_version="22.05.9",state="mixed",tres_billing=36i,tres_cpu=36i,tres_mem=94793i,tres_used_cpu=33i,tres_used_mem=50000i,weight=1i 1723039486000000000
slurm_nodes,host=hoth,name=naboo216,source=slurm_primary.example.net alloc_cpu=8i,alloc_memory=8000i,architecture="x86_64",cores=4i,cpu_load=554i,cpus=8i,free_memory=27101i,real_memory=31877i,slurmd_version="22.05.9",state="allocated",tres_billing=8i,tres_cpu=8i,tres_mem=31877i,tres_used_cpu=8i,tres_used_mem=8000i,weight=1i 1723039486000000000
slurm_nodes,host=hoth,name=naboo219,source=slurm_primary.example.net alloc_cpu=16i,alloc_memory=16000i,architecture="x86_64",cores=4i,cpu_load=919i,cpus=16i,free_memory=1841i,real_memory=31875i,slurmd_version="22.05.9",state="allocated",tres_billing=16i,tres_cpu=16i,tres_mem=31875i,tres_used_cpu=16i,tres_used_mem=16000i,weight=1i 1723039486000000000
slurm_partitions,host=hoth,name=atlas,source=slurm_primary.example.net nodes="naboo145,naboo146,naboo147,naboo216,naboo219,naboo222,naboo224,naboo225,naboo227,naboo228,naboo229,naboo234,naboo235,naboo236,naboo237,naboo238,naboo239,naboo240,naboo241,naboo242,naboo243",state="UP",total_cpu=632i,total_nodes=21i,tres_billing=632i,tres_cpu=632i,tres_mem=1415207i,tres_node=21i 1723039486000000000
```
