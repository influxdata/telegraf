# RDAP Input Plugin

This plugin queries domain registration data over [RDAP][rdap] and exposes
metrics such as expiration timestamps, registrar details and domain status.
RDAP is the successor to WHOIS and returns structured data, which lets it pick
up expiry information for TLDs that no longer publish it over WHOIS (for
example `.dev` and other Google-operated zones).

The metrics are kept compatible with the [`whois`][whois] plugin so both can
feed the same dashboards.

⭐ Telegraf v1.40.0
🏷️ network, web
💻 all

[rdap]: https://www.rfc-editor.org/rfc/rfc7480
[whois]: /plugins/inputs/whois

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
# Reads domain registration data via RDAP and exposes it as metrics
[[inputs.rdap]]
  ## List of domains to query
  domains = ["example.com", "influxdata.com"]

  ## Custom RDAP server to query directly, e.g. "https://rdap.org"
  ## When empty the IANA bootstrap registry is used to find the
  ## authoritative server for each domain.
  # server = ""

  ## Timeout for RDAP queries
  # timeout = "30s"
```

By default each domain is bootstrapped through the [IANA registry][bootstrap]
to locate its authoritative RDAP server. Set `server` to send every query to a
single endpoint instead, such as the aggregator at `https://rdap.org`.

[bootstrap]: https://data.iana.org/rdap/dns.json

## Metrics

- rdap
  - tags:
    - domain
    - status (string)
  - fields:
    - creation_timestamp (int, seconds)
    - dnssec_enabled (bool)
    - error (string) - only set when a lookup fails
    - expiration_timestamp (int, seconds)
    - expiry (int, seconds) - Remaining time until the domain expires, in
        seconds. This value can be **negative** if the domain is already
        expired. `SELECT (expiry / 60 / 60 / 24) as "expiry_in_days"`
    - name_servers (string) - comma separated list
    - registrar (string)
    - registrant (string)
    - updated_timestamp (int, seconds)

## Example Output

```text
rdap,domain=example.com,status=client\ delete\ prohibited\,client\ transfer\ prohibited creation_timestamp=771891456i,dnssec_enabled=true,expiration_timestamp=2145916800i,expiry=6300000i,name_servers="a.iana-servers.net,b.iana-servers.net",registrar="RESERVED-Internet Assigned Numbers Authority",registrant="not set",updated_timestamp=1692181820i 1700000000000000000
```
