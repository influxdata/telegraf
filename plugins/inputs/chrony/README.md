# chrony Input Plugin

This plugin queries metrics from a chrony NTP server. For details on the
meaning of the gathered fields please check the [chronyc manual][]

[chronyc manual]: https://chrony-project.org/doc/4.4/chronyc.html

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Get standard chrony metrics, requires chronyc executable.
[[inputs.chrony]]
  ## Server address of chronyd with address scheme
  ## If empty or not set, the plugin will mimic the behavior of chronyc and
  ## check "unix:///run/chrony/chronyd.sock", "udp://127.0.0.1:323"
  ## and "udp://[::1]:323".
  # server = ""

  ## Timeout for establishing the connection
  # timeout = "5s"

  ## Try to resolve received addresses to host-names via DNS lookups
  ## Disabled by default to avoid DNS queries especially for slow DNS servers.
  # dns_lookup = false

  ## Metrics to query named according to chronyc commands
  ## Available settings are:
  ##   activity    -- number of peers online or offline
  ##   tracking    -- information about system's clock performance
  ##   serverstats -- chronyd server statistics
  ##   sources     -- extended information about peers
  ##   sourcestats -- statistics on peers
  # metrics = ["tracking"]
```

## Metrics

- chrony
  - system_time (float, seconds)
  - last_offset (float, seconds)
  - rms_offset (float, seconds)
  - frequency (float, ppm)
  - residual_freq (float, ppm)
  - skew (float, ppm)
  - root_delay (float, seconds)
  - root_dispersion (float, seconds)
  - update_interval (float, seconds)

### Tags

- All measurements have the following tags:
  - reference_id
  - stratum
  - leap_status

## Example Output

```text
chrony,leap_status=not\ synchronized,reference_id=A29FC87B,stratum=3 frequency=-16.000999450683594,last_offset=0.000012651000361074694,residual_freq=0,rms_offset=0.000025576999178156257,root_delay=0.0016550000291317701,root_dispersion=0.00330700003542006,skew=0.006000000052154064,system_time=0.000020389999917824753,update_interval=507.1999816894531 1706271167571675297
```
