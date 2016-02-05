# Dovecot Input Plugin

The dovecot plugin uses the dovecot Stats protocol to gather metrics on configured
domains. You can read Dovecot's documentation
[here](http://wiki2.dovecot.org/Statistics)

### Configuration:

```
# Read metrics about docker containers
[[inputs.dovecot]]
  # Dovecot Endpoint
  #   To use TCP, set endpoint = "tcp://[ip]:[port]"
  #   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"
  # Only collect metrics for these domains, collect all if empty
  domains_names = []

```

### Fields:

	domain                 string
	reset_timestamp        time.Time
	last_update            time.Time
	num_logins             int
	num_cmds               int
	num_connected_sessions int
	user_cpu               float32
	sys_cpu                float32
	clock_time             float64
	min_faults             int
	maj_faults             int
	vol_cs                 int
	invol_cs               int
	disk_input             int
	disk_output            int
	read_count             int
	read_bytes             int
	write_count            int
	write_bytes            int
	mail_lookup_path       int
	mail_lookup_attr       int
	mail_read_count        int
	mail_read_bytes        int
	mail_cache_hits        int

### Example Output:

```
% ./telegraf -config ~/ws/telegraf.conf -input-filter docker -test
* Plugin: docker, Collection 1

```