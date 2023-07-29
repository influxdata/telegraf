# Jenkins Input Plugin

The jenkins plugin gathers information about the nodes and jobs running in a
jenkins instance.

This plugin does not require a plugin on jenkins and it makes use of Jenkins API
to retrieve all the information needed.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Read jobs and cluster metrics from Jenkins instances
[[inputs.jenkins]]
  ## The Jenkins URL in the format "schema://host:port"
  url = "http://my-jenkins-instance:8080"
  # username = "admin"
  # password = "admin"

  ## Set response_timeout
  response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional Max Job Build Age filter
  ## Default 1 hour, ignore builds older than max_build_age
  # max_build_age = "1"

  ## Jobs to include or exclude from gathering
  ## When using both lists, job_exclude has priority.
  ## Wildcards are supported: [ "jobA/*", "jobB/subjob1/*"]
  # job_include = [ "*" ]
  # job_exclude = [ ]

  ## Worker pool for jenkins plugin only
  ## Max idle connections in the pool.
  # max_idle_connections = 10

  ## Max number of builds per job
  ## Defaults to 30
  # max_num_builds = 30

  ## Max number workers to get builds
  # max_workers = 5
```

## Metrics

- jenkins_job_v2
  - tags:
    - server
    - name
    - parents
    - result
  - fields:
    - estimated_duration
    - duration
    - result_code
    - number
  - jenkins_executors_v2:
    - tags:
      - server
    - fields:
      - busy_executors
      - total_executors


## Sample Queries

```sql
SELECT mean("memory_available") AS "mean_memory_available", mean("memory_total") AS "mean_memory_total", mean("temp_available") AS "mean_temp_available" FROM "jenkins_node" WHERE time > now() - 15m GROUP BY time(:interval:) FILL(null)
```

```sql
SELECT mean("duration") AS "mean_duration" FROM "jenkins_job" WHERE time > now() - 24h GROUP BY time(:interval:) FILL(null)
```

## Example Output

```text
jenkins,host=myhost,port=80,source=my-jenkins-instance busy_executors=4i,total_executors=8i 1580418261000000000
jenkins_node,arch=Linux\ (amd64),disk_path=/var/jenkins_home,temp_path=/tmp,host=myhost,node_name=master,source=my-jenkins-instance,port=8080 swap_total=4294963200,memory_available=586711040,memory_total=6089498624,status=online,response_time=1000i,disk_available=152392036352,temp_available=152392036352,swap_available=3503263744,num_executors=2i 1516031535000000000
jenkins_job,host=myhost,name=JOB1,parents=apps/br1,result=SUCCESS,source=my-jenkins-instance,port=8080 duration=2831i,result_code=0i 1516026630000000000
jenkins_job,host=myhost,name=JOB2,parents=apps/br2,result=SUCCESS,source=my-jenkins-instance,port=8080 duration=2285i,result_code=0i 1516027230000000000
```
