# SLURM Input Plugin

This plugin gather diagnoses, jobs, nodes, partitions and reservation metrics
for a [SLURM][slurm] instance using the REST API provided by the `slurmrestd`
daemon.

> [!NOTE]
> This plugin supports the [REST API v0.0.38][api] which must be enabled in the
> `slurmrestd` daemon. For more information, check the [documentation][config].

⭐ Telegraf v1.32.0
🏷️ server
💻 all

[slurm]: https://slurm.schedmd.com
[api]: https://slurm.schedmd.com/rest.html
[config]: https://slurm.schedmd.com/rest_quickstart.html#customization

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

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

  ## Enabled endpoints
  ## List of endpoints a user can acquire data from.
  ## Available values are: diag, jobs, nodes, partitions, reservations.
  # enabled_endpoints = ["diag", "jobs", "nodes", "partitions", "reservations"]

  ## Maximum time to receive a response. If set to 0s, the
  ## request will not time out.
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
slurm_diag,host=hoth,source=slurm_primary.example.net bf_active=false,bf_queue_len=1i,bf_queue_len_mean=1i,jobs_canceled=0i,jobs_completed=137i,jobs_failed=0i,jobs_pending=0i,jobs_running=100i,jobs_started=137i,jobs_submitted=137i,schedule_cycle_last=27i,schedule_cycle_mean=86i,server_thread_count=3i 1723466497000000000
slurm_jobs,host=hoth,job_id=23160,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.11BCgQ",cpus=2i,current_working_directory="/home/sessiondir/7CQODmQ3uw5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmG9JKDmILUkln",group_id=2005i,nice=50i,node_count=1i,nodes="naboo225",partition="atlas",priority=4294878569i,standard_error="/home/sessiondir/7CQODmQ3uw5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmG9JKDmILUkln.comment",standard_input="/dev/null",standard_output="/home/sessiondir/7CQODmQ3uw5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmG9JKDmILUkln.comment",start_time=1723354525i,state="RUNNING",state_reason="None",submit_time=1723354525i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=2000,tres_node=1 1723466497000000000
slurm_jobs,host=hoth,job_id=23365,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.yRcFYL",cpus=2i,current_working_directory="/home/sessiondir/LgwNDmTLAx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm2BKKDm8bFZsm",group_id=2005i,nice=50i,node_count=1i,nodes="naboo224",partition="atlas",priority=4294878364i,standard_error="/home/sessiondir/LgwNDmTLAx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm2BKKDm8bFZsm.comment",standard_input="/dev/null",standard_output="/home/sessiondir/LgwNDmTLAx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm2BKKDm8bFZsm.comment",start_time=1723376763i,state="RUNNING",state_reason="None",submit_time=1723376761i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=1000,tres_node=1 1723466497000000000
slurm_jobs,host=hoth,job_id=23366,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.5Y9Ngb",cpus=2i,current_working_directory="/home/sessiondir/HFYKDmULAx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm3BKKDmiyK3em",group_id=2005i,nice=50i,node_count=1i,nodes="naboo225",partition="atlas",priority=4294878363i,standard_error="/home/sessiondir/HFYKDmULAx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm3BKKDmiyK3em.comment",standard_input="/dev/null",standard_output="/home/sessiondir/HFYKDmULAx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm3BKKDmiyK3em.comment",start_time=1723376883i,state="RUNNING",state_reason="None",submit_time=1723376882i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=1000,tres_node=1 1723466497000000000
slurm_jobs,host=hoth,job_id=23367,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.NmOqMU",cpus=2i,current_working_directory="/home/sessiondir/nnLLDmULAx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm4BKKDmfhjFPn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo225",partition="atlas",priority=4294878362i,standard_error="/home/sessiondir/nnLLDmULAx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm4BKKDmfhjFPn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/nnLLDmULAx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm4BKKDmfhjFPn.comment",start_time=1723376883i,state="RUNNING",state_reason="None",submit_time=1723376882i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=1000,tres_node=1 1723466497000000000
slurm_jobs,host=hoth,job_id=23385,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.NNsI08",cpus=2i,current_working_directory="/home/sessiondir/PWvNDmH7tw5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmz7JKDmqgKyRo",group_id=2005i,nice=50i,node_count=1i,nodes="naboo225",partition="atlas",priority=4294878344i,standard_error="/home/sessiondir/PWvNDmH7tw5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmz7JKDmqgKyRo.comment",standard_input="/dev/null",standard_output="/home/sessiondir/PWvNDmH7tw5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmz7JKDmqgKyRo.comment",start_time=1723378725i,state="RUNNING",state_reason="None",submit_time=1723378725i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=1000,tres_node=1 1723466497000000000
slurm_jobs,host=hoth,job_id=23386,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.bcmS4h",cpus=2i,current_working_directory="/home/sessiondir/ZNHMDmI7tw5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm27JKDm3Ve66n",group_id=2005i,nice=50i,node_count=1i,nodes="naboo224",partition="atlas",priority=4294878343i,standard_error="/home/sessiondir/ZNHMDmI7tw5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm27JKDm3Ve66n.comment",standard_input="/dev/null",standard_output="/home/sessiondir/ZNHMDmI7tw5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm27JKDm3Ve66n.comment",start_time=1723379206i,state="RUNNING",state_reason="None",submit_time=1723379205i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=1000,tres_node=1 1723466497000000000
slurm_jobs,host=hoth,job_id=23387,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.OgpoQZ",cpus=2i,current_working_directory="/home/sessiondir/qohNDmUqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmMCKKDmzM4Yhn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo222",partition="atlas",priority=4294878342i,standard_error="/home/sessiondir/qohNDmUqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmMCKKDmzM4Yhn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/qohNDmUqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmMCKKDmzM4Yhn.comment",start_time=1723379246i,state="RUNNING",state_reason="None",submit_time=1723379245i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=1000,tres_node=1 1723466497000000000
slurm_jobs,host=hoth,job_id=23388,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.xYbxSe",cpus=2i,current_working_directory="/home/sessiondir/u9HODmXqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmWCKKDmRlccYn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo224",partition="atlas",priority=4294878341i,standard_error="/home/sessiondir/u9HODmXqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmWCKKDmRlccYn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/u9HODmXqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmWCKKDmRlccYn.comment",start_time=1723379326i,state="RUNNING",state_reason="None",submit_time=1723379326i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=1000,tres_node=1 1723466497000000000
slurm_jobs,host=hoth,job_id=23389,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.QHtIIm",cpus=2i,current_working_directory="/home/sessiondir/ZLvKDmYqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmXCKKDmjp19km",group_id=2005i,nice=50i,node_count=1i,nodes="naboo227",partition="atlas",priority=4294878340i,standard_error="/home/sessiondir/ZLvKDmYqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmXCKKDmjp19km.comment",standard_input="/dev/null",standard_output="/home/sessiondir/ZLvKDmYqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmXCKKDmjp19km.comment",start_time=1723379326i,state="RUNNING",state_reason="None",submit_time=1723379326i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=1000,tres_node=1 1723466497000000000
slurm_jobs,host=hoth,job_id=23393,name=gridjob,source=slurm_primary.example.net command="/tmp/SLURM_job_script.IH19bN",cpus=2i,current_working_directory="/home/sessiondir/YdPODmVqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmSCKKDmrYDOwm",group_id=2005i,nice=50i,node_count=1i,nodes="naboo224",partition="atlas",priority=4294878336i,standard_error="/home/sessiondir/YdPODmVqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmSCKKDmrYDOwm.comment",standard_input="/dev/null",standard_output="/home/sessiondir/YdPODmVqBx5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmSCKKDmrYDOwm.comment",start_time=1723379767i,state="RUNNING",state_reason="None",submit_time=1723379766i,tasks=1i,time_limit=3600i,tres_billing=1,tres_cpu=1,tres_mem=1000,tres_node=1 1723466497000000000
slurm_nodes,host=hoth,name=naboo145,source=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=18i,cpu_load=0i,cpus=36i,free_memory=86450i,real_memory=94791i,slurmd_version="22.05.9",state="idle",tres_billing=36,tres_cpu=36,tres_mem=94791,weight=1i 1723466497000000000
slurm_nodes,host=hoth,name=naboo146,source=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=18i,cpu_load=0i,cpus=36i,free_memory=92148i,real_memory=94791i,slurmd_version="22.05.9",state="idle",tres_billing=36,tres_cpu=36,tres_mem=94791,weight=1i 1723466497000000000
slurm_nodes,host=hoth,name=naboo147,source=slurm_primary.example.net alloc_cpu=36i,alloc_memory=45000i,architecture="x86_64",cores=18i,cpu_load=3826i,cpus=36i,free_memory=1607i,real_memory=94793i,slurmd_version="22.05.9",state="allocated",tres_billing=36,tres_cpu=36,tres_mem=94793,tres_used_cpu=36,tres_used_mem=45000,weight=1i 1723466497000000000
slurm_nodes,host=hoth,name=naboo216,source=slurm_primary.example.net alloc_cpu=8i,alloc_memory=8000i,architecture="x86_64",cores=4i,cpu_load=891i,cpus=8i,free_memory=17972i,real_memory=31877i,slurmd_version="22.05.9",state="allocated",tres_billing=8,tres_cpu=8,tres_mem=31877,tres_used_cpu=8,tres_used_mem=8000,weight=1i 1723466497000000000
slurm_nodes,host=hoth,name=naboo219,source=slurm_primary.example.net alloc_cpu=16i,alloc_memory=16000i,architecture="x86_64",cores=4i,cpu_load=1382i,cpus=16i,free_memory=15645i,real_memory=31875i,slurmd_version="22.05.9",state="allocated",tres_billing=16,tres_cpu=16,tres_mem=31875,tres_used_cpu=16,tres_used_mem=16000,weight=1i 1723466497000000000
slurm_partitions,host=hoth,name=atlas,source=slurm_primary.example.net nodes="naboo145,naboo146,naboo147,naboo216,naboo219,naboo222,naboo224,naboo225,naboo227,naboo228,naboo229,naboo234,naboo235,naboo236,naboo237,naboo238,naboo239,naboo240,naboo241,naboo242,naboo243",state="UP",total_cpu=632i,total_nodes=21i,tres_billing=632,tres_cpu=632,tres_mem=1415207,tres_node=21 1723466497000000000
```
