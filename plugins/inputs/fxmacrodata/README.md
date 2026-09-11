# FXMacroData Input Plugin

This plugin collects official macroeconomic, foreign exchange and
release-calendar data from the [FXMacroData][fxmacrodata] service, which
aggregates central banks and national statistics agencies across 18 currencies.

Each point is stamped with the instant the figure was **published**, not the
instant it was collected. A macro figure is only meaningful alongside when it
became known — the value for a given month is not knowable during that month —
so using the publication instant keeps the series honest when it is graphed or
compared against market data.

> [!IMPORTANT]
> USD macro data is public and needs no credential. An [API key][api_key] widens
> access to the other seventeen currencies, the full history, and FX.

⭐ Telegraf v1.37.0
🏷️ applications, web
💻 all

[fxmacrodata]: https://fxmacrodata.com/?utm_source=github&utm_medium=referral&utm_campaign=telegraf&utm_content=readme
[api_key]: https://fxmacrodata.com/subscribe?utm_source=github&utm_medium=referral&utm_campaign=telegraf&utm_content=subscribe

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `api_key` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Read official macroeconomic, FX and central-bank data from FXMacroData
[[inputs.fxmacrodata]]
  ## API key. USD macro data is public, so this can be omitted; a key widens
  ## access to the other seventeen currencies, the full history, and FX.
  # api_key = ""

  ## Currencies to gather indicators for.
  # currencies = ["USD"]

  ## Indicator slugs to gather for each currency. Query
  ## /v1/data_catalogue/{currency} to see what a currency publishes.
  # indicators = ["inflation", "policy_rate"]

  ## Currency pairs to gather reference rates for.
  # fx_pairs = ["EUR/USD"]

  ## Override the API base URL.
  # base_url = "https://api.fxmacrodata.com"

  ## HTTP response timeout.
  # response_timeout = "5s"

  ## Optional TLS config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  # insecure_skip_verify = false
```

Indicator slugs differ by currency. `GET /v1/data_catalogue/{currency}` lists
what a given currency publishes.

A currency the key does not cover is reported as an error for that series and
the remaining series are still gathered, so a mixed `currencies` list stays
usable.

## Metrics

- fxmacrodata_indicator
  - tags:
    - currency
    - indicator
    - source (the publishing authority, when reported)
  - fields:
    - value (float)
    - reference_date (string, the period the figure describes)

- fxmacrodata_fx
  - tags:
    - base
    - quote
  - fields:
    - rate (float)
    - reference_date (string)

A period the authority did not report comes back as a null value and is skipped
rather than recorded as zero, which would be indistinguishable from a real
reading of zero.

## Example Output

```text
fxmacrodata_indicator,currency=USD,indicator=inflation,source=BLS value=3.4,reference_date="2026-07-31" 1786537800000000000
fxmacrodata_fx,base=EUR,quote=USD rate=1.1616,reference_date="2026-09-10" 1789036200000000000
```
