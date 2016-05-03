# Telegraf plugin: Haproxy

#### Configuration

```toml
 [[inputs.haproxy]]
#   ## An array of address to gather stats about. Specify an ip on hostname
#   ## with optional port. ie localhost, 10.10.3.33:1936, etc.
#
#   ## If no servers are specified, then default to 127.0.0.1:1936
#   servers = ["http://myhaproxy.com:1936/;csv", "http://anotherhaproxy.com:1936/;csv"]
    servers = ["http://localhost:8080/haproxy?stats;csv"]
```

URI will be different for each server. Plugin doesn't modify it when it queries to Haproxy stats webpage. 
