package etherscan

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf/plugins/inputs/etherscan/mocks"
	"github.com/stretchr/testify/suite"
)

func TestBalanceRequestSuites(t *testing.T) {
	suite.Run(t, new(BalanceRequestSuite))
}

type BalanceRequestSuite struct {
	suite.Suite
	apiKey         string
	network        Network
	log            *mocks.Logger
	balanceRequest *BalanceRequest
}

type _ suite.SetupTestSuite

func (suite *BalanceRequestSuite) SetupTest() {
	suite.network = mainnet
	suite.log = &mocks.Logger{}
	suite.apiKey = "abc123"

	parsedURL, err := url.Parse(suite.network.URL())
	assert.Nil(suite.T(), err)

	testBalanceRequest := NewBalanceRequest(parsedURL, suite.apiKey, suite.log)
	assert.NotNil(suite.T(), testBalanceRequest.log)
	assert.Equal(suite.T(), testBalanceRequest.log, suite.log)
	assert.NotNil(suite.T(), testBalanceRequest.lastTriggered)
	assert.IsType(suite.T(), time.Time{}, testBalanceRequest.lastTriggered)
	assert.NotNil(suite.T(), testBalanceRequest.request)
	assert.IsType(suite.T(), &url.URL{}, testBalanceRequest.request)
	assert.NotNil(suite.T(), testBalanceRequest.tags)
	assert.IsType(suite.T(), map[string]string{}, testBalanceRequest.tags)

	testURL := testBalanceRequest.URL()
	apiModule := testURL.Query().Get("module")
	assert.NotEmpty(suite.T(), apiModule)
	assert.Equal(suite.T(), "account", apiModule)

	apiTag := testURL.Query().Get("tag")
	assert.NotEmpty(suite.T(), apiTag)
	assert.Equal(suite.T(), "latest", apiTag)

	apiKey := testURL.Query().Get("apikey")
	assert.NotEmpty(suite.T(), apiKey)
	assert.Equal(suite.T(), suite.apiKey, apiKey)

	suite.balanceRequest = testBalanceRequest
}

func (suite *BalanceRequestSuite) TestAddAddressToBalance() {
	suite.log.On("Debugf", mock.Anything, mock.Anything).Return(nil)

	testAddress := "0x123456789123456789"
	err := suite.balanceRequest.AddAddress(testAddress)
	assert.Nil(suite.T(), err)

	addresses := suite.balanceRequest.RequestAddressList()
	assert.Len(suite.T(), addresses, 1)
	assert.Contains(suite.T(), addresses, testAddress)

	suite.balanceRequest.URL()
}

func (suite *BalanceRequestSuite) TestAddAddressToMultiBalance() {
	suite.log.On("Debugf", mock.Anything, mock.Anything).Return(nil)

	testAddress := "0x123456789123456789"
	err := suite.balanceRequest.AddAddress(testAddress)
	assert.Nil(suite.T(), err)

	testAddress2 := "0x987654321987654321"
	err = suite.balanceRequest.AddAddress(testAddress2)
	assert.Nil(suite.T(), err)

	addresses := suite.balanceRequest.RequestAddressList()
	assert.Len(suite.T(), addresses, 2)
	assert.Contains(suite.T(), addresses, testAddress)
	assert.Contains(suite.T(), addresses, testAddress2)

	suite.balanceRequest.URL()
}
