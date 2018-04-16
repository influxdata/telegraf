# Jenkins Plugin

The jenkins plugin gathers information about the nodes and jobs running in a jenkins instance.

This plugin does not require a plugin on jenkins and it makes use of Jenkins API to retrieve all the information needed.

### Configuration:

```toml
  url = "http://my-jenkins-instance:8080"
  user = "admin"
  passwd = "admin"
  ## Set response_timeout
  response_timeout = "5s"

  ## Optional SSL Config
  # ssl_ca = /path/to/cafile
  # ssl_cert = /path/to/certfile
  # ssl_key = /path/to/keyfile
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Job & build filter
  lastbuild_interval = "1h"
  # job_exclude = [ "MyJob", "MyOtherJob" ]
```

### Measurements & Fields:

- jenkins_node
    - disk_available
    - temp_available
    - memory_available
    - memory_total
    - swap_available
    - swap_total
    - response_time
    - online

- jenkins_job
    - duration
    - result_code (0 = SUCCESS, 1 = FAILURE, 2 = NOT_BUILD, 3 = UNSTABLE, 4 = ABORTED)

### Tags:

- jenkins_node
    - arch
    - disk_path
    - temp_path
    - node_name

- jenkins_job
    - job_name
    - result

### Example Output:

```
$ ./telegraf --config telegraf.conf --input-filter jenkins --test
jenkins_node,arch=Linux\ (amd64),disk_path=/var/jenkins_home,temp_path=/tmp,host=myhost,node_name=master swap_total=4294963200,memory_available=586711040,memory_total=6089498624,online=true,response_time=1000i,disk_available=152392036352,temp_available=152392036352,swap_available=3503263744 1516031535000000000
jenkins_job,host=myhost,job_name=JOB1,result=SUCCESS duration=2831i,result_code=0i 1516026630000000000
jenkins_job,host=myhost,job_name=JOB2,result=SUCCESS duration=2285i,result_code=0i 1516027230000000000
```
