# Read metrics about docker containers
[[inputs.docker]]
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"

  ## Set to true to collect Swarm metrics(desired_replicas, running_replicas)
  ## Note: configure this in one of the manager nodes in a Swarm cluster.
  ## configuring in multiple Swarm managers results in duplication of metrics.
  gather_services = false

  ## Only collect metrics for these containers. Values will be appended to
  ## container_name_include.
  ## Deprecated (1.4.0), use container_name_include
  container_names = []

  ## Set the source tag for the metrics to the container ID hostname, eg first 12 chars
  source_tag = false

  ## Containers to include and exclude. Collect all if empty. Globs accepted.
  container_name_include = []
  container_name_exclude = []

  ## Container states to include and exclude. Globs accepted.
  ## When empty only containers in the "running" state will be captured.
  ## example: container_state_include = ["created", "restarting", "running", "removing", "paused", "exited", "dead"]
  ## example: container_state_exclude = ["created", "restarting", "running", "removing", "paused", "exited", "dead"]
  # container_state_include = []
  # container_state_exclude = []

  ## Timeout for docker list, info, and stats commands
  timeout = "5s"

  ## Whether to report for each container per-device blkio (8:0, 8:1...),
  ## network (eth0, eth1, ...) and cpu (cpu0, cpu1, ...) stats or not.
  ## Usage of this setting is discouraged since it will be deprecated in favor of 'perdevice_include'.
  ## Default value is 'true' for backwards compatibility, please set it to 'false' so that 'perdevice_include' setting
  ## is honored.
  perdevice = true

  ## Specifies for which classes a per-device metric should be issued
  ## Possible values are 'cpu' (cpu0, cpu1, ...), 'blkio' (8:0, 8:1, ...) and 'network' (eth0, eth1, ...)
  ## Please note that this setting has no effect if 'perdevice' is set to 'true'
  # perdevice_include = ["cpu"]

  ## Whether to report for each container total blkio and network stats or not.
  ## Usage of this setting is discouraged since it will be deprecated in favor of 'total_include'.
  ## Default value is 'false' for backwards compatibility, please set it to 'true' so that 'total_include' setting
  ## is honored.
  total = false

  ## Specifies for which classes a total metric should be issued. Total is an aggregated of the 'perdevice' values.
  ## Possible values are 'cpu', 'blkio' and 'network'
  ## Total 'cpu' is reported directly by Docker daemon, and 'network' and 'blkio' totals are aggregated by this plugin.
  ## Please note that this setting has no effect if 'total' is set to 'false'
  # total_include = ["cpu", "blkio", "network"]

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  docker_label_include = []
  docker_label_exclude = []

  ## Which environment variables should we use as a tag
  tag_env = ["JAVA_HOME", "HEAP_SIZE"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
