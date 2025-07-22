# Network Time Protocol Query Input Plugin

This plugin gathers metrics about [Network Time Protocol][ntp] queries.

> [!IMPORTANT]
> This plugin requires the `ntpq` executable to be installed on the system.

⭐ Telegraf v0.11.0
🏷️ network
💻 all

[ntp]: https://ntp.org/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Get standard NTP query metrics, requires ntpq executable.
[[inputs.ntpq]]
  ## Servers to query with ntpq.
  ## If no server is given, the local machine is queried.
  # servers = []

  ## Options to pass to the ntpq command.
  # options = "-p"

  ## Output format for the 'reach' field.
  ## Available values are
  ##   octal   --  output as is in octal representation e.g. 377 (default)
  ##   decimal --  convert value to decimal representation e.g. 371 -> 249
  ##   count   --  count the number of bits in the value. This represents
  ##               the number of successful reaches, e.g. 37 -> 5
  ##   ratio   --  output the ratio of successful attempts e.g. 37 -> 5/8 = 0.625
  # reach_format = "octal"
```

You can pass arbitrary options accepted by the `ntpq` command using the
`options` setting. In case you want to skip DNS lookups use

```toml
  options = "-p -n"
```

for example.

Below is the documentation of the various headers returned from the NTP query
command when running `ntpq -p`.

- `remote` – The remote peer or server being synced to. “LOCAL” is this local
    host (included in case there are no remote peers or servers available);
- `refid` – Where or what the remote peer or server is itself synchronised to;
- `st` (stratum) – The remote peer or server Stratum
- `t` (type) – Type (u: unicast or manycast client, b: broadcast or multicast
    client, l: local reference clock, s: symmetric peer, A: manycast server,
    B: broadcast server, M: multicast server, see “Automatic Server Discovery“);
- `when` – When last polled (seconds ago, “h” hours ago, or “d” days ago);
- `poll` – Polling frequency: rfc5905 suggests this ranges in NTPv4 from 4 (16s)
    to 17 (36h) (log2 seconds), however observation suggests the actual
    displayed value is seconds for a much smaller range of 64 (26) to 1024
    (210) seconds;
- `reach` – An 8-bit left-shift shift register value recording polls
    (bit set = successful, bit reset = fail) displayed in octal;
- `delay` – Round trip communication delay to the remote peer or server
    (milliseconds);
- `offset` – Mean offset (phase) in the times reported between this local host
    and the remote peer or server (RMS, milliseconds);
- `jitter` – Mean deviation (jitter) in the time reported for that remote peer
    or server (RMS of difference of multiple time samples, milliseconds);

## Metrics

- `ntpq`
  - delay (float, milliseconds)
  - jitter (float, milliseconds)
  - offset (float, milliseconds)
  - poll (int, seconds)
  - reach (int)
  - when (int, seconds)

### Tags

All measurements have the following tags:

- refid
- remote
- type
- stratum

In case you are specifying `servers`, the measurement has an
additional `source` tag.

## Example Output

```text
ntpq,refid=.GPSs.,remote=*time.apple.com,stratum=1,type=u delay=91.797,jitter=3.735,offset=12.841,poll=64i,reach=377i,when=35i 1457960478909556134
```
