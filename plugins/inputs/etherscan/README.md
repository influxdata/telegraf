# Etherscan Input Plugin

Gather account balance (ETH) from [Etherscan][] Ethereum block explorer.

**Note:** Telegraf also contains the [webhook][] input which can be used as an
alternative method for gathering API data.

**Note:** Currently only ETH reporting is available but erc20 support is on the way.

### Configuration

```toml
[[inputs.etherscan]]
    # Etherscan API token
    api_key = "ETHERSCAN-APIKEY-HERE"

    # Etherscan rate limit number of calls allowed (default: 5)
    api_rate_limit_calls = 5

    # Etherscan rate limit duration (default: 1 second)
    api_rate_limit_duration = "1s"

    # HTTP client request timeout duration. Useful if connection is poor. (default: 3 second)
    req_timeout_duration = "3s"

    [[inputs.etherscan.balance]]
        address = "0x0000000000000000000000000000000000000000"
        network = "Mainnet"
    [[inputs.etherscan.balance]]
        address = "0x0000000000000000000000000000000000000001"
        network = "Mainnet"
```

### Metrics

+ internal_etherscan
  - fields:
    - address - balance

### Example Output

```
balances,0x0000000000000000000000000000000000000000=709,0x0000000000000000000000000000000000000001=2344 1563901372000000000
```

[Etherscan]: https://etherscan.io/
[webhook]: /plugins/inputs/webhooks/github
