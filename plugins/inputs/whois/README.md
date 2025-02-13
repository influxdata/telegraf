# WHOIS Input Plugin

The `whois` input plugin queries **WHOIS** information for configured
domains and provides metrics such as expiration timestamps, registrar
details, and domain status.

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Reads WHOIS data and expose as metrics
[[inputs.whois]]
  ## List of domains to query
  domains = ["example.com", "influxdata.com"]

  ## Use Custom WHOIS server
  ##   example: server = "whois.iana.org"
  # server = ""

  ## Timeout for WHOIS queries
  # timeout = "5s"
```

## Metrics

- whois
  - tags:
    - domain
  - fields:
    - creation_timestamp (int, seconds)
    - dnssec_enabled (bool)
    - expiration_timestamp (int, seconds)
    - expiry (int, seconds) - Remaining time until the domain expires, in seconds.
        This value can be **negative** if the domain is already expired.
        `SELECT (expiry / 60 / 60 / 24) as "expiry_in_days"`
    - registrar (string)
    - registrant (string)
    - status_code (int)
      - 0 - Unknown
      - 1 - Pending Delete
      - 2 - Expired
      - 3 - Locked
      - 4 - Registered
      - 5 - Active
    - updated_timestamp (int, seconds)

## Example Output

```text
domain=example.com creation_timestamp=808372800i,dnssec_enabled=true,expiration_timestamp=1755057600i,expiry=15596396i,name_servers="a.iana-servers.net,b.iana-servers.net",registrant="",registrar="RESERVED-Internet Assigned Numbers Authority",status_code=3i,updated_timestamp=1723618894i 1739461204000000000
domain=influxdata.com creation_timestamp=1403603283i,dnssec_enabled=false,expiration_timestamp=1750758483i,expiry=11297277i,name_servers="ns-1200.awsdns-22.org,ns-127.awsdns-15.com,ns-2037.awsdns-62.co.uk,ns-820.awsdns-38.net",registrant="Redacted for Privacy",registrar="NameCheap, Inc.",status_code=3i,updated_timestamp=1716620263i 1739461206000000000
```
