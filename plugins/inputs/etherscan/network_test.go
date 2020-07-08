package etherscan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestNetworkSuites(t *testing.T) {
	suite.Run(t, new(NetworkMainnetSuite))
	suite.Run(t, new(NetworkRopstenSuite))
	suite.Run(t, new(NetworkKovanSuite))
	suite.Run(t, new(NetworkRinkebySuite))
	suite.Run(t, new(NetworkGoerliSuite))
	suite.Run(t, new(NetworkUnknownSuite))
}

type NetworkMainnetSuite struct {
	suite.Suite
	mainnet Network
}

type _ suite.SetupTestSuite

func (suite *NetworkMainnetSuite) SetupTest() {
	suite.mainnet = mainnet
}

func (suite *NetworkMainnetSuite) TestMainnetIOTA() {
	assert.Equal(suite.T(), suite.mainnet, Network(0))
}

func (suite *NetworkMainnetSuite) TestMainnetString() {
	assert.Equal(suite.T(), suite.mainnet.String(), "Mainnet")
}

func (suite *NetworkMainnetSuite) TestMainnetURL() {
	assert.Equal(suite.T(), suite.mainnet.URL(), _etherscanAPIMainnetURL)
}

func (suite *NetworkMainnetSuite) TestMainnetLookup() {
	net, err := NetworkLookup("Mainnet")
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), net)
	assert.Equal(suite.T(), net, suite.mainnet)
}

type NetworkRopstenSuite struct {
	suite.Suite
	ropsten Network
}

type _ suite.SetupTestSuite

func (suite *NetworkRopstenSuite) SetupTest() {
	suite.ropsten = ropsten
}

func (suite *NetworkRopstenSuite) TestRopstenIOTA() {
	assert.Equal(suite.T(), suite.ropsten, Network(1))
}

func (suite *NetworkRopstenSuite) TestRopstenString() {
	assert.Equal(suite.T(), suite.ropsten.String(), "Ropsten")
}

func (suite *NetworkRopstenSuite) TestRopstenURL() {
	assert.Equal(suite.T(), suite.ropsten.URL(), _etherscanAPIRopstenURL)
}

func (suite *NetworkRopstenSuite) TestRopstenLookup() {
	net, err := NetworkLookup("Ropsten")
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), net)
	assert.Equal(suite.T(), net, suite.ropsten)
}

type NetworkKovanSuite struct {
	suite.Suite
	kovan Network
}

type _ suite.SetupTestSuite

func (suite *NetworkKovanSuite) SetupTest() {
	suite.kovan = kovan
}

func (suite *NetworkKovanSuite) TestKovanIOTA() {
	assert.Equal(suite.T(), suite.kovan, Network(2))
}

func (suite *NetworkKovanSuite) TestKovanString() {
	assert.Equal(suite.T(), suite.kovan.String(), "Kovan")
}

func (suite *NetworkKovanSuite) TestKovanURL() {
	assert.Equal(suite.T(), suite.kovan.URL(), _etherscanAPIKovanURL)
}

func (suite *NetworkKovanSuite) TestKovanLookup() {
	net, err := NetworkLookup("Kovan")
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), net)
	assert.Equal(suite.T(), net, suite.kovan)
}

type NetworkRinkebySuite struct {
	suite.Suite
	rinkeby Network
}

type _ suite.SetupTestSuite

func (suite *NetworkRinkebySuite) SetupTest() {
	suite.rinkeby = rinkeby
}

func (suite *NetworkRinkebySuite) TestRinkebyIOTA() {
	assert.Equal(suite.T(), suite.rinkeby, Network(3))
}

func (suite *NetworkRinkebySuite) TestRinkebyString() {
	assert.Equal(suite.T(), suite.rinkeby.String(), "Rinkeby")
}

func (suite *NetworkRinkebySuite) TestRinkebyURL() {
	assert.Equal(suite.T(), suite.rinkeby.URL(), _etherscanAPIRinkebyURL)
}

func (suite *NetworkRinkebySuite) TestRinkebyLookup() {
	net, err := NetworkLookup("Rinkeby")
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), net)
	assert.Equal(suite.T(), net, suite.rinkeby)
}

type NetworkGoerliSuite struct {
	suite.Suite
	goerli Network
}

type _ suite.SetupTestSuite

func (suite *NetworkGoerliSuite) SetupTest() {
	suite.goerli = goerli
}

func (suite *NetworkGoerliSuite) TestGoerliIOTA() {
	assert.Equal(suite.T(), suite.goerli, Network(4))
}

func (suite *NetworkGoerliSuite) TestGoerliString() {
	assert.Equal(suite.T(), suite.goerli.String(), "Goerli")
}

func (suite *NetworkGoerliSuite) TestGoerliURL() {
	assert.Equal(suite.T(), suite.goerli.URL(), _etherscanAPIGoerliURL)
}

func (suite *NetworkGoerliSuite) TestGoerliLookup() {
	net, err := NetworkLookup("Goerli")
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), net)
	assert.Equal(suite.T(), net, suite.goerli)
}

type NetworkUnknownSuite struct {
	suite.Suite
	unknown Network
}

type _ suite.SetupTestSuite

func (suite *NetworkUnknownSuite) SetupTest() {
	suite.unknown = Network(99)
}

func (suite *NetworkUnknownSuite) TestUnknownIOTA() {
	assert.Equal(suite.T(), suite.unknown, Network(99))
}

func (suite *NetworkUnknownSuite) TestUnknownString() {
	assert.Equal(suite.T(), suite.unknown.String(), "unknown")
}

func (suite *NetworkUnknownSuite) TestUnknownURL() {
	assert.Equal(suite.T(), suite.unknown.URL(), "")
}

func (suite *NetworkUnknownSuite) TestUnknownLookup() {
	net, err := NetworkLookup("Unknown")
	assert.NotNil(suite.T(), net)
	assert.Equal(suite.T(), net, Network(-1))
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), err, ErrNetworkNotValid)
}
