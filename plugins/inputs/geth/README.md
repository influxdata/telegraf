## Geth Input Plugin

Geth ([github.com/ethereum/go-ethereum](https://github.com/ethereum/go-ethereum) / [geth.ethereum.org](https://geth.ethereum.org) is an official Ethereum client. This plugin allows users to collect metrics from their Geth nodes.

### Requirements

This plugin requires a running `geth` node that has been configured with the `--metrics` flag as well as the `--rpc --rpcapi debug,...` flags. An example node configuration can be found in the [dev/](dev/) test directory.

### Selecting Metrics

In the toml config, a subset of all metrics can be chosen using the gjson syntax, for example:

```
[[inputs.geth]]
  servers = [
    "http://localhost:8545"
  ]
  ## Each metric in this list is a gjson query path to specify a specific chunk of JSON to be parsed.
  ## gjson query paths are described here: https://github.com/tidwall/gjson#path-syntax
  metrics = [
    "eth.db.chaindata.compact.input",
    "eth.db.chaindata.compact.output"
  ]
```

This would result in the following fields getting collected under the `"geth"` metric:

```
eth_db_chaindata_compact_input_avgrate01min
eth_db_chaindata_compact_input_avgrate05min
eth_db_chaindata_compact_input_avgrate15min
eth_db_chaindata_compact_input_meanrate
eth_db_chaindata_compact_input_overall
eth_db_chaindata_compact_output_avgrate01min
eth_db_chaindata_compact_output_avgrate05min
eth_db_chaindata_compact_output_avgrate15min
eth_db_chaindata_compact_output_meanrate
eth_db_chaindata_compact_output_overall
```

### All Metrics

You can find the full list of metrics available in your geth instance with the following curl command:

```
curl -XPOST \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -d '{"jsonrpc":"2.0","method":"debug_metrics","params":[true],"id":1}' \
  http://localhost:8545
```
