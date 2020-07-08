package etherscan

import (
	"errors"
	"strings"
)

const (
	_etherscanAPIMainnetURL = "https://api.etherscan.io/api"
	_etherscanAPIRopstenURL = "https://api-ropsten.etherscan.io/api"
	_etherscanAPIKovanURL   = "https://api-kovan.etherscan.io/api"
	_etherscanAPIRinkebyURL = "https://api-rinkeby.etherscan.io/api"
	_etherscanAPIGoerliURL  = "https://api-goerli.etherscan.io/api"
)

var (
	ErrNetworkNotValid = errors.New("not a valid network")
)

type Network int

const (
	mainnet Network = iota
	ropsten
	kovan
	rinkeby
	goerli
)

func (n Network) String() string {
	switch n {
	case mainnet:
		return "Mainnet"
	case ropsten:
		return "Ropsten"
	case kovan:
		return "Kovan"
	case rinkeby:
		return "Rinkeby"
	case goerli:
		return "Goerli"
	default:
		return "unknown"
	}
}

func NetworkLookup(network string) (Network, error) {
	normStr := strings.ToLower(network)
	switch normStr {
	case "mainnet":
		return mainnet, nil
	case "ropsten":
		return ropsten, nil
	case "kovan":
		return kovan, nil
	case "rinkeby":
		return rinkeby, nil
	case "goerli":
		return goerli, nil
	default:
		return -1, ErrNetworkNotValid
	}
}

func (n Network) URL() string {
	switch n {
	case mainnet:
		return _etherscanAPIMainnetURL
	case ropsten:
		return _etherscanAPIRopstenURL
	case kovan:
		return _etherscanAPIKovanURL
	case rinkeby:
		return _etherscanAPIRinkebyURL
	case goerli:
		return _etherscanAPIGoerliURL
	default:
		return ""
	}
}
