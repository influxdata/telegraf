# Raindrops Input Plugin

The [raindrops](http://raindrops.bogomips.org/) plugin reads from
specified raindops middleware URI and adds stats to InfluxDB.
### Configuration:

```toml
# Read raindrops stats
[[inputs.raindrops]]
  urls = ["http://localhost/_raindrops"]
```

### Tags:

- Multiple listeners are tagged with IP:Port/Socket, ie `0.0.0.0:8080` or  `/tmp/unicorn`
