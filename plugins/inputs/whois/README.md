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
  # server = "whois.iana.org"

  ## Configuration to export name servers as a field
  # include_name_servers = true

  ## Timeout for WHOIS queries
  # timeout = "5s"
```

## Metrics

- whois
  - tags:
    - domain
  - fields:
    - creation_timestamp (int, seconds)
    - dnssec_enabled (int, 0 = false / error, 1 = true)
    - domain_status (string)
    - expiration_timestamp (int, seconds)
    - expiry (int, seconds) - Time when the domain will expire, in seconds
      since the Unix epoch. `SELECT (expiry / 60 / 60 / 24) as "expiry_in_days"`
    - registrar (string)
    - status (int, 0 = error, 1 = ok) - WHOIS scraping or parser status
    - updated_timestamp (int, seconds)

## Example Output

```text
whois,domain=example.com creation_timestamp=808372800i,dnssec_enabled=1i,domain_status="LOCKED",expiration_timestamp=1755057600i,expiry=15655393i,name_servers="a.iana-servers.net,b.iana-servers.net",registrar="RESERVED-Internet Assigned Numbers Authority",status=1i,updated_timestamp=1723618894i 1739402208000000000
whois,domain=influxdata.com creation_timestamp=1403603283i,dnssec_enabled=0i,domain_status="LOCKED",expiration_timestamp=1750758483i,expiry=11356276i,name_servers="ns-1200.awsdns-22.org,ns-127.awsdns-15.com,ns-2037.awsdns-62.co.uk,ns-820.awsdns-38.net",registrar="NameCheap, Inc.",status=1i,updated_timestamp=1716620263i 1739402215000000000
whois,domain=influxdata1245.com,domain_status=UNKNOWN status=0i 1739402216000000000
```
