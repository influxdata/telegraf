# Jenkins Plugin

The jenkins plugin gathers information about the nodes and jobs running in a jenkins instance.

This plugin does not require a plugin on jenkins and it makes use of Jenkins API to retrieve all the information needed.

### Configuration:

```toml
  ## The Jenkins URL
  url = "http://my-jenkins-instance:8080"
  #  username = "admin"
  #  password = "admin"

  ## Set response_timeout
  response_timeout = "5s"

  ## Optional SSL Config
  #  ssl_ca = /path/to/cafile
  #  ssl_cert = /path/to/certfile
  #  ssl_key = /path/to/keyfile
  ## Use SSL but skip chain & host verification
  #  insecure_skip_verify = false

  ## Optional Max Job Build Age filter
  ## Default 1 hour, ignore builds older than max_build_age
  #  max_build_age = "1h"

  ## Optional Sub Job Depth filter
  ## Jenkins can have unlimited layer of sub jobs
  ## This config will limit the layers of pulling, default value 0 means
  ## unlimited pulling until no more sub jobs
  #  max_subjob_depth = 0

  ## Optional Sub Job Per Layer
  ## In workflow-multibranch-plugin, each branch will be created as a sub job.
  ## This config will limit to call only the lasted branches in each layer,
  ## empty will use default value 10
  #  max_subjob_per_layer = 10

  ## Jobs to exclude from gathering
  #  job_exclude = [ "job1", "job2/subjob1/subjob2", "job3/*"]

  ## Nodes to exclude from gathering
  #  node_exclude = [ "node1", "node2" ]

  ## Worker pool for jenkins plugin only
  ## Empty this field will use default value 5
  #  max_connections = 5
```

### Metrics:

- jenkins_node
  - tags:
    - arch
    - disk_path
    - temp_path
    - node_name
    - status ("online", "offline")
  - fields:
    - disk_available
    - temp_available
    - memory_available
    - memory_total
    - swap_available
    - swap_total
    - response_time

- jenkins_job
  - tags:
    - name
    - parents
    - result
  - fields:
    - duration
    - result_code (0 = SUCCESS, 1 = FAILURE, 2 = NOT_BUILD, 3 = UNSTABLE, 4 = ABORTED)

### Sample Queries:

```
SELECT mean("memory_available") AS "mean_memory_available", mean("memory_total") AS "mean_memory_total", mean("temp_available") AS "mean_temp_available" FROM "jenkins_node" WHERE time > now() - 15m GROUP BY time(:interval:) FILL(null)
```

```
SELECT mean("duration") AS "mean_duration" FROM "jenkins_job" WHERE time > now() - 24h GROUP BY time(:interval:) FILL(null)
```

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter jenkins --test
jenkins_node,arch=Linux\ (amd64),disk_path=/var/jenkins_home,temp_path=/tmp,host=myhost,node_name=master swap_total=4294963200,memory_available=586711040,memory_total=6089498624,status=online,response_time=1000i,disk_available=152392036352,temp_available=152392036352,swap_available=3503263744 1516031535000000000
jenkins_job,host=myhost,name=JOB1,parents=apps/br1,result=SUCCESS duration=2831i,result_code=0i 1516026630000000000
jenkins_job,host=myhost,name=JOB2,parents=apps/br2,result=SUCCESS duration=2285i,result_code=0i 1516027230000000000
```