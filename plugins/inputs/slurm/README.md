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
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Metrics

Given the great deal of metrics offered by SLURM's API, an attempt has been
done to strike a balance between verbosity and usefulness in terms of the
gathered information.

- slurm_diag
  - tags:
    - url
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
    - url
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
    - tres_req_str
- slurm_nodes
  - tags:
    - url
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
    - tres
    - tres_used
    - weight
    - slurmd_version
    - architecture
- slurm_partitions
  - tags:
    - url
    - name
  - fields:
    - state
    - total_cpu
    - total_nodes
    - nodes
    - tres
- slurm_reservations
  - tags:
    - url
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
slurm_diag,host=hoth,url=slurm_primary.example.net bf_active=false,bf_queue_len=1i,bf_queue_len_mean=1i,jobs_canceled=0i,jobs_completed=396i,jobs_failed=0i,jobs_pending=0i,jobs_running=100i,jobs_started=396i,jobs_submitted=396i,schedule_cycle_last=301i,schedule_cycle_mean=137i,server_thread_count=3i 1722599914000000000
slurm_jobs,host=hoth,job_id=16869,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.rmuCZn",cpus=8i,current_working_directory="/home/sessiondir/iF8NDm0gZt5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmvYIKDm5RcaYo",group_id=2005i,nice=50i,node_count=1i,nodes="naboo219",partition="atlas",priority=4294884860i,standard_error="/home/sessiondir/iF8NDm0gZt5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmvYIKDm5RcaYo.comment",standard_input="/dev/null",standard_output="/home/sessiondir/iF8NDm0gZt5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmvYIKDm5RcaYo.comment",start_time=1722506622i,state="RUNNING",state_reason="None",submit_time=1722500413i,tasks=8i,time_limit=3600i,tres_req_str="cpu=8,mem=8000M,node=1,billing=8" 1722599914000000000
slurm_jobs,host=hoth,job_id=16965,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.OHQ65g",cpus=2i,current_working_directory="/home/sessiondir/9tIMDmS6at5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmkaIKDmU2Bmin",group_id=2005i,nice=50i,node_count=1i,nodes="naboo228",partition="atlas",priority=4294884764i,standard_error="/home/sessiondir/9tIMDmS6at5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmkaIKDmU2Bmin.comment",standard_input="/dev/null",standard_output="/home/sessiondir/9tIMDmS6at5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmkaIKDmU2Bmin.comment",start_time=1722520367i,state="COMPLETED",state_reason="None",submit_time=1722520366i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=1000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17036,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.DmbM7i",cpus=8i,current_working_directory="/home/sessiondir/M59LDmmyct5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmzbIKDmEzwdwn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo225",partition="atlas",priority=4294884693i,standard_error="/home/sessiondir/M59LDmmyct5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmzbIKDmEzwdwn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/M59LDmmyct5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmzbIKDmEzwdwn.comment",start_time=1722543684i,state="RUNNING",state_reason="None",submit_time=1722543684i,tasks=8i,time_limit=3600i,tres_req_str="cpu=8,mem=8000M,node=1,billing=8" 1722599914000000000
slurm_jobs,host=hoth,job_id=17483,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.UM7Uy8",cpus=1i,current_working_directory="/home/sessiondir/PLuMDmUV2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmGiIKDmy0UFLn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo145",partition="atlas",priority=4294884246i,standard_error="/home/sessiondir/PLuMDmUV2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmGiIKDmy0UFLn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/PLuMDmUV2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmGiIKDmy0UFLn.comment",start_time=1722597933i,state="RUNNING",state_reason="None",submit_time=1722597932i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17484,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.c7w3hs",cpus=1i,current_working_directory="/home/sessiondir/Y4kNDmUV2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmHiIKDmd9hu4m",group_id=2005i,nice=50i,node_count=1i,nodes="naboo145",partition="atlas",priority=4294884245i,standard_error="/home/sessiondir/Y4kNDmUV2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmHiIKDmd9hu4m.comment",standard_input="/dev/null",standard_output="/home/sessiondir/Y4kNDmUV2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmHiIKDmd9hu4m.comment",start_time=1722598133i,state="RUNNING",state_reason="None",submit_time=1722598133i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17485,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.CfJRgA",cpus=1i,current_working_directory="/home/sessiondir/b5MKDmVV2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmIiIKDmDpl2qn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo146",partition="atlas",priority=4294884244i,standard_error="/home/sessiondir/b5MKDmVV2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmIiIKDmDpl2qn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/b5MKDmVV2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmIiIKDmDpl2qn.comment",start_time=1722598133i,state="RUNNING",state_reason="None",submit_time=1722598133i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17487,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.Sp1flx",cpus=1i,current_working_directory="/home/sessiondir/tj2KDmm6qt5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm9eIKDmnvWPWm",group_id=2005i,nice=50i,node_count=1i,nodes="naboo145",partition="atlas",priority=4294884242i,standard_error="/home/sessiondir/tj2KDmm6qt5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm9eIKDmnvWPWm.comment",standard_input="/dev/null",standard_output="/home/sessiondir/tj2KDmm6qt5nKG01gq4B3BRpm7wtQmABFKDmbnHPDm9eIKDmnvWPWm.comment",start_time=1722598373i,state="RUNNING",state_reason="None",submit_time=1722598373i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17488,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.CQW3VD",cpus=1i,current_working_directory="/home/sessiondir/yIZKDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmJiIKDmWua4Im",group_id=2005i,nice=50i,node_count=1i,nodes="naboo147",partition="atlas",priority=4294884241i,standard_error="/home/sessiondir/yIZKDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmJiIKDmWua4Im.comment",standard_input="/dev/null",standard_output="/home/sessiondir/yIZKDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmJiIKDmWua4Im.comment",start_time=1722598413i,state="RUNNING",state_reason="None",submit_time=1722598413i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17489,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.9AJ71s",cpus=2i,current_working_directory="/home/sessiondir/AANLDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmKiIKDml7ZIBo",group_id=2005i,nice=50i,node_count=1i,nodes="naboo228",partition="atlas",priority=4294884240i,standard_error="/home/sessiondir/AANLDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmKiIKDml7ZIBo.comment",standard_input="/dev/null",standard_output="/home/sessiondir/AANLDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmKiIKDml7ZIBo.comment",start_time=1722598575i,state="RUNNING",state_reason="None",submit_time=1722598573i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17490,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.h6Sb0u",cpus=1i,current_working_directory="/home/sessiondir/IqBMDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmLiIKDmRqth1m",group_id=2005i,nice=50i,node_count=1i,nodes="naboo145",partition="atlas",priority=4294884239i,standard_error="/home/sessiondir/IqBMDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmLiIKDmRqth1m.comment",standard_input="/dev/null",standard_output="/home/sessiondir/IqBMDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmLiIKDmRqth1m.comment",start_time=1722598614i,state="RUNNING",state_reason="None",submit_time=1722598613i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17491,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.cj7tCa",cpus=1i,current_working_directory="/home/sessiondir/nuzMDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmMiIKDm0BPL4n",group_id=2005i,nice=50i,node_count=1i,nodes="naboo147",partition="atlas",priority=4294884238i,standard_error="/home/sessiondir/nuzMDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmMiIKDm0BPL4n.comment",standard_input="/dev/null",standard_output="/home/sessiondir/nuzMDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmMiIKDm0BPL4n.comment",start_time=1722599054i,state="RUNNING",state_reason="None",submit_time=1722599054i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17492,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.F3ZNhV",cpus=1i,current_working_directory="/home/sessiondir/E3nNDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmNiIKDm6gEWPo",group_id=2005i,nice=50i,node_count=1i,nodes="naboo147",partition="atlas",priority=4294884237i,standard_error="/home/sessiondir/E3nNDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmNiIKDm6gEWPo.comment",standard_input="/dev/null",standard_output="/home/sessiondir/E3nNDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmNiIKDm6gEWPo.comment",start_time=1722599094i,state="RUNNING",state_reason="None",submit_time=1722599094i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17493,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.Y20pKI",cpus=1i,current_working_directory="/home/sessiondir/QznKDmIe2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmOiIKDmcxBZgm",group_id=2005i,nice=50i,node_count=1i,nodes="naboo147",partition="atlas",priority=4294884236i,standard_error="/home/sessiondir/QznKDmIe2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmOiIKDmcxBZgm.comment",standard_input="/dev/null",standard_output="/home/sessiondir/QznKDmIe2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmOiIKDmcxBZgm.comment",start_time=1722599655i,state="RUNNING",state_reason="None",submit_time=1722599655i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17494,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.XuKz2N",cpus=1i,current_working_directory="/home/sessiondir/RXbLDmIe2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmPiIKDmNtLiUn",group_id=2005i,nice=50i,node_count=1i,nodes="naboo147",partition="atlas",priority=4294884235i,standard_error="/home/sessiondir/RXbLDmIe2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmPiIKDmNtLiUn.comment",standard_input="/dev/null",standard_output="/home/sessiondir/RXbLDmIe2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmPiIKDmNtLiUn.comment",start_time=1722599816i,state="RUNNING",state_reason="None",submit_time=1722599815i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_jobs,host=hoth,job_id=17495,name=gridjob,url=slurm_primary.example.net command="/tmp/SLURM_job_script.jDwqdW",cpus=1i,current_working_directory="/home/sessiondir/i6RMDm6j2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmQiIKDmcfLIhm",group_id=2005i,nice=50i,node_count=1i,nodes="naboo147",partition="atlas",priority=4294884234i,standard_error="/home/sessiondir/i6RMDm6j2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmQiIKDmcfLIhm.comment",standard_input="/dev/null",standard_output="/home/sessiondir/i6RMDm6j2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmQiIKDmcfLIhm.comment",start_time=1722599896i,state="RUNNING",state_reason="None",submit_time=1722599895i,tasks=1i,time_limit=3600i,tres_req_str="cpu=1,mem=2000M,node=1,billing=1" 1722599914000000000
slurm_nodes,host=hoth,name=naboo145,url=slurm_primary.example.net alloc_cpu=20i,alloc_memory=31000i,architecture="x86_64",cores=18i,cpu_load=3100i,cpus=36i,free_memory=1839i,real_memory=94791i,slurmd_version="22.05.9",state="mixed",tres="cpu=36,mem=94791M,billing=36",tres_used="cpu=20,mem=31000M",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo234,url=slurm_primary.example.net alloc_cpu=15i,alloc_memory=23000i,architecture="x86_64",cores=8i,cpu_load=1618i,cpus=16i,free_memory=8793i,real_memory=63779i,slurmd_version="22.05.9",state="mixed",tres="cpu=16,mem=63779M,billing=16",tres_used="cpu=15,mem=23000M",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo235,url=slurm_primary.example.net alloc_cpu=8i,alloc_memory=16000i,architecture="x86_64",cores=8i,cpu_load=910i,cpus=16i,free_memory=9144i,real_memory=63779i,slurmd_version="22.05.9",state="mixed",tres="cpu=16,mem=63779M,billing=16",tres_used="cpu=8,mem=16000M",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo236,url=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=8i,cpu_load=0i,cpus=16i,free_memory=49749i,real_memory=63779i,slurmd_version="22.05.9",state="idle",tres="cpu=16,mem=63779M,billing=16",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo237,url=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=8i,cpu_load=0i,cpus=16i,free_memory=54614i,real_memory=63779i,slurmd_version="22.05.9",state="idle",tres="cpu=16,mem=63779M,billing=16",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo238,url=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=24i,cpu_load=0i,cpus=48i,free_memory=106862i,real_memory=104223i,slurmd_version="20.11.8",state="idle",tres="cpu=48,mem=104223M,billing=48",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo239,url=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=24i,cpu_load=0i,cpus=96i,free_memory=105203i,real_memory=104223i,slurmd_version="20.11.8",state="idle",tres="cpu=96,mem=104223M,billing=96",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo240,url=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=18i,cpu_load=1i,cpus=36i,free_memory=72383i,real_memory=94789i,slurmd_version="22.05.9",state="idle",tres="cpu=36,mem=94789M,billing=36",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo241,url=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=32i,cpu_load=0i,cpus=64i,free_memory=118235i,real_memory=127901i,slurmd_version="22.05.9",state="idle",tres="cpu=64,mem=127901M,billing=64",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo242,url=slurm_primary.example.net alloc_cpu=0i,alloc_memory=0i,architecture="x86_64",cores=24i,cpu_load=0i,cpus=48i,free_memory=83883i,real_memory=94789i,slurmd_version="22.05.9",state="idle",tres="cpu=48,mem=94789M,billing=48",weight=1i 1722599914000000000
slurm_nodes,host=hoth,name=naboo243,url=slurm_primary.example.net alloc_cpu=16i,alloc_memory=32000i,architecture="x86_64",cores=24i,cpu_load=1601i,cpus=48i,free_memory=10205i,real_memory=94789i,slurmd_version="22.05.9",state="mixed",tres="cpu=48,mem=94789M,billing=48",tres_used="cpu=16,mem=32000M",weight=1i 1722599914000000000
slurm_partitions,host=hoth,name=atlas,url=slurm_primary.example.net nodes="naboo145,naboo146,naboo147,naboo216,naboo219,naboo222,naboo224,naboo225,naboo227,naboo228,naboo229,naboo234,naboo235,naboo236,naboo237,naboo238,naboo239,naboo240,naboo241,naboo242,naboo243",state="UP",total_cpu=632i,total_nodes=21i,tres="cpu=632,mem=1415207M,node=21,billing=632" 1722599914000000000
```
