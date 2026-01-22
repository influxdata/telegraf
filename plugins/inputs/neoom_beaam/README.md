# Neoom Beaam Input Plugin

This plugin gathers metrics from a [Neoom Beaam gateway][beaam] using the
[Beaam API][beaam_api] with access token that can be created in the Neoom web
interface. Please follow the [developer instructions][devpage] to create the
token.

⭐ Telegraf v1.33.0
🏷️ iot
💻 all

[beaam]: https://neoom.com/en/products/beaam
[beaam_api]: https://developer.neoom.com/reference/concepts-terms-1
[devpage]:https://neoom.com/developers

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `token` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Read energy data from the local Neoom Beaam gateway
[[inputs.neoom_beaam]]
  ## Address of the gateway
  # address = "https://10.10.10.10"

  ## Bearer token for accessing the data
  # token = ""

  ## Enforce refreshing the site configuration in each gather cycle
  # refresh_configuration = false

  ## Optional TLS Config
  ## Set to true/false to enforce TLS being enabled/disabled. If not set,
  ## enable TLS only if any of the other options are specified.
  # tls_enable =
  ## Trusted root certificates for server
  # tls_ca = "/path/to/cafile"
  ## Used for TLS client certificate authentication
  # tls_cert = "/path/to/certfile"
  ## Used for TLS client certificate authentication
  # tls_key = "/path/to/keyfile"
  ## Password for the key file if it is encrypted
  # tls_key_pwd = ""
  ## Send the specified TLS server name via SNI
  # tls_server_name = "kubernetes.example.com"
  ## Minimal TLS version to accept by the client
  # tls_min_version = "TLS12"
  ## List of ciphers to accept, by default all secure ciphers will be accepted
  ## See https://pkg.go.dev/crypto/tls#pkg-constants for supported values.
  ## Use "all", "secure" and "insecure" to add all support ciphers, secure
  ## suites or insecure suites respectively.
  # tls_cipher_suites = ["secure"]
  ## Renegotiation method, "never", "once" or "freely"
  # tls_renegotiation_method = "never"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

## Example Output

For energy flow metrics:

```text
neoom_beaam_energy_flow,datapoint=SELF_SUFFICIENCY,source=127.0.0.1,unit=% value=100 1723883145024999936
neoom_beaam_energy_flow,datapoint=POWER_PRODUCTION,source=127.0.0.1,unit=W value=1184 1723883150001999872
neoom_beaam_energy_flow,datapoint=POWER_GRID,source=127.0.0.1,unit=W value=15 1723883150001999872
neoom_beaam_energy_flow,datapoint=POWER_STORAGE,source=127.0.0.1,unit=W value=3504 1723883150001999872
neoom_beaam_energy_flow,datapoint=ENERGY_PRODUCED,source=127.0.0.1,unit=Wh value=639300 1723882905007000064
neoom_beaam_energy_flow,datapoint=ENERGY_IMPORTED,source=127.0.0.1,unit=Wh value=22696.926152777138 1723883150001999872
neoom_beaam_energy_flow,datapoint=ENERGY_EXPORTED,source=127.0.0.1,unit=Wh value=237421.33112499933 1723883135003000064
neoom_beaam_energy_flow,datapoint=ENERGY_CHARGED,source=127.0.0.1,unit=Wh value=215200 1723882765001999872
neoom_beaam_energy_flow,datapoint=ENERGY_DISCHARGED,source=127.0.0.1,unit=Wh value=144800 1723880490004999936
neoom_beaam_energy_flow,datapoint=STATE_OF_CHARGE,source=127.0.0.1,unit=% value=59 1723883090001999872
neoom_beaam_energy_flow,datapoint=POWER_CONSUMPTION_CALC,source=127.0.0.1,unit=W value=4703 1723883150001999872
neoom_beaam_energy_flow,datapoint=ENERGY_CONSUMED_CALC,source=127.0.0.1,unit=Wh value=354175.59502777783 1723883150001999872
```

for thing (aka device) metrics

```text
neoom_beaam_thing,datapoint=CONNECTION,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=None value=true 1723890960060000000
neoom_beaam_thing,datapoint=VOLTAGE_P1,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=V value=245 1723905135056000000
neoom_beaam_thing,datapoint=VOLTAGE_P2,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=V value=242.3 1723905135056000000
neoom_beaam_thing,datapoint=VOLTAGE_P3,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=V value=245.5 1723905135056000000
neoom_beaam_thing,datapoint=FREQUENCY,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=Hz value=49.98 1723905135056000000
neoom_beaam_thing,datapoint=ACTIVE_POWER,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=W value=-826 1723905135056000000
neoom_beaam_thing,datapoint=OUTPUT_ENERGY,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=Wh value=241826.44736444377 1723905135075000064
neoom_beaam_thing,datapoint=INPUT_ENERGY,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=Wh value=22988.396324999278 1723904090096000000
neoom_beaam_thing,datapoint=POWER_P1,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=W value=-305 1723905135056000000
neoom_beaam_thing,datapoint=POWER_P2,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=W value=-268 1723905135056000000
neoom_beaam_thing,datapoint=POWER_P3,source=127.0.0.1,thing=ELECTRICITY_METER_AC,unit=W value=-214 1723905135056000000
```

## Metrics

- neoom_beaam_energy_flow
  - tags:
    - source (address of gateway)
    - datapoint (name of the datapoint)
    - unit (unit of the measurement)
  - fields:
    - value (value of the datapoint)

- neoom_beaam_thing
  - tags:
    - source (address of gateway)
    - thing (name of the device)
    - datapoint (name of the datapoint)
    - unit (unit of the measurement)
  - fields:
    - value (value of the datapoint)
