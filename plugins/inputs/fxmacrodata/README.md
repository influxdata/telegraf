# FXMacroData Input Plugin

This plugin collects official macroeconomic indicators and FX reference rates
from the [FXMacroData][fxmacrodata] service, which publishes data from central
banks and national statistics agencies across 22 currencies.

Each metric is stamped with the instant the figure was **published**, not the
instant it was collected. A macro figure is only meaningful together with the
time it became known, so the publication instant keeps the series usable when
it is graphed or compared against market data.

> [!NOTE]
> USD indicators are available without a credential. Other currencies and FX
> rates require an [API key][api_key].

⭐ Telegraf v1.41.0
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
# Read macroeconomic and FX data from FXMacroData
[[inputs.fxmacrodata]]
  ## API endpoint
  # url = "https://api.fxmacrodata.com/v1"

  ## API key for additional features like currencies, history, etc
  # api_key = ""

  ## Currencies to gather indicators for, in uppercase
  # currencies = ["USD"]

  ## Indicators to gather for each currency
  # indicators = ["inflation", "policy_rate"]

  ## Currency pairs to gather reference rates for, in uppercase
  # fx_pairs = ["EUR/USD"]

  ## HTTP response timeout, zero means no timeout
  # response_timeout = "5s"

  ## Optional TLS config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  # insecure_skip_verify = false
```

The `url` setting only needs to be changed when going through a proxy or a
mirror of the API.

The `api_key` is sent in the `X-API-Key` request header. Without a key only USD
indicators can be gathered, and the most recent releases become available 15
minutes after publication. With a key all currencies covered by the key, FX
rates and releases without delay are available. See the
[API reference][reference] for details.

Currencies and currency pairs must be specified as uppercase ISO 4217 codes,
e.g. `USD` or `EUR/USD`. The available indicators differ by currency, query
`https://api.fxmacrodata.com/v1/data_catalogue/{currency}` to list the
indicators of a currency.

A series the key does not cover is reported as an error for that series while
the remaining series are still gathered, so a mixed `currencies` list stays
usable. A series that does not exist for a currency is silently skipped.

Setting `response_timeout` to zero disables the timeout.

[reference]: https://fxmacrodata.com/documentation/reference

## Metrics

- fxmacrodata_indicator
  - tags:
    - currency
    - indicator
    - source (the publishing authority)
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

Every metric carries the same set of tags and fields. If the latest period was
not reported by the authority the API returns a null value. In this case no
metric is produced, as any placeholder value would be indistinguishable from a
real reading.

## Example Output

```text
fxmacrodata_indicator,currency=USD,indicator=inflation,source=BLS value=3.4,reference_date="2026-08-31" 1789129800000000000
fxmacrodata_fx,base=EUR,quote=USD rate=1.1616,reference_date="2026-09-10" 1789036200000000000
```
