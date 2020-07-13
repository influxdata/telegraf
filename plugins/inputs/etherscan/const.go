package etherscan

// Plugin specific constants required for operation
const (
	_ExampleEtherscanConfig = `
    [[inputs.etherscan]]
      api_key = "ETHERSCAN-APIKEY-HERE"
      api_rate_limit_calls = 5
      api_rate_limit_duration = "1s"
      req_timeout_duration = "3s"
    [[inputs.etherscan.balance]]
      address = "0x0000000000000000000000000000000000000000"
      network = "Mainnet"
    `

	_Description = `Gather information from Ethereum networks via the Etherscan Block Explorer API`
)

// Etherscan specific API default values
const (
	_RequestLimit    = 5
	_RequestInterval = "1s"
	_RequestTimeout  = "3s"
)
