# Dovecot Input Plugin

The dovecot plugin uses the dovecot Stats protocol to gather metrics on configured
domains. You can read Dovecot's documentation
[here](http://wiki2.dovecot.org/Statistics)

### Configuration:

```
# Read metrics about dovecot servers
[[inputs.dovecot]]
  # Dovecot servers
  #  specify dovecot servers via an address:port list
  #  e.g.
  #    localhost:24242
  #
  # If no servers are specified, then localhost is used as the host.
  servers = ["localhost:24242"]
  # Only collect metrics for these domains, collect all if empty
  domains = []
```

### Tags:
	server: hostname
	domain: domain name

### Fields:

	reset_timestamp        time.Time
	last_update            time.Time
	num_logins             int64
	num_cmds               int64
	num_connected_sessions int64
	user_cpu               float32
	sys_cpu                float32
	clock_time             float64
	min_faults             int64
	maj_faults             int64
	vol_cs                 int64
	invol_cs               int64
	disk_input             int64
	disk_output            int64
	read_count             int64
	read_bytes             int64
	write_count            int64
	write_bytes            int64
	mail_lookup_path       int64
	mail_lookup_attr       int64
	mail_read_count        int64
	mail_read_bytes        int64
	mail_cache_hits        int64
