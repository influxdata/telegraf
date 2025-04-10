# WHOIS Input Plugin

This plugin queries [WHOIS information][whois] for configured
domains and provides metrics such as expiration timestamps, registrar
details and domain status from e.g. [IANA][iana] or [ICANN][icann]
servers.

⭐ Telegraf v1.35.0
🏷️ network, web
💻 all

[whois]: https://datatracker.ietf.org/doc/html/rfc3912
[icann]: https://lookup.icann.org/
[iana]: https://www.iana.org/whois

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Reads whois data and expose as metrics
[[inputs.whois]]
  ## List of domains to query
  domains = ["example.com", "influxdata.com"]

  ## Use Custom WHOIS server
  # server = "whois.iana.org"

  ## Timeout for WHOIS queries
  # timeout = "30s"

  ## Enable WHOIS referral chain query
  # referral_chain_query = false
```

## Metrics

- whois
  - tags:
    - domain
    - status (string)
  - fields:
    - creation_timestamp (int, seconds)
    - dnssec_enabled (bool)
    - error (string)
    - expiration_timestamp (int, seconds)
    - expiry (int, seconds) - Remaining time until the domain expires, in seconds.
        This value can be **negative** if the domain is already expired.
        `SELECT (expiry / 60 / 60 / 24) as "expiry_in_days"`
    - registrar (string)
    - registrant (string)
    - updated_timestamp (int, seconds)

## Example Output

```text
whois,domain=example.com,status=unknown creation_timestamp=694224000i,dnssec_enabled=false,expiration_timestamp=0i,expiry=0i,name_servers="",registrant="",registrar="",updated_timestamp=0i 1741128738000000000
whois,domain=influxdata.com,status=clientTransferProhibited creation_timestamp=1403603283i,dnssec_enabled=false,expiration_timestamp=1750758483i,expiry=9629744i,name_servers="ns-1200.awsdns-22.org,ns-127.awsdns-15.com,ns-2037.awsdns-62.co.uk,ns-820.awsdns-38.net",registrant="",registrar="NameCheap, Inc.",updated_timestamp=1716620263i 1741128738000000000
whois,domain=influxdata-test.com,status=not\ found error="whoisparser: domain is not found" 1741128739000000000
```
