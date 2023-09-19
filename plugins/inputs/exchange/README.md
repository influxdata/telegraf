# Exchange Rate Input Plugin

## Configuration

```toml @sample.conf
[[inputs.exchange]]
  ## Required currency exchange API token. Visit https://freecurrencyapi.com/ to get token.
  apikey = ""

  ## Required base currency.
  base_currency = "EUR"

  ## Target currency.
  target_currency = "USD"
```
