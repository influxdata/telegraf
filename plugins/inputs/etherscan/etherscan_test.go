package etherscan

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/inputs/etherscan/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestEtherscanInputPluginSuites(t *testing.T) {
	suite.Run(t, new(EtherscanSuite))
}

type EtherscanSuite struct {
	suite.Suite

	mockHTTPClient  *MockHTTPClient
	mockLogger      *mocks.Logger
	inputPlugin     *Etherscan
	balanceAccounts []balance
}

type _ suite.SetupTestSuite

func (suite *EtherscanSuite) SetupTest() {
	suite.mockHTTPClient = &MockHTTPClient{}
	suite.mockLogger = &mocks.Logger{}
	suite.balanceAccounts = make([]balance, 0)

	suite.inputPlugin = NewEtherscanInputPlugin(
		WithHTTPClient(suite.mockHTTPClient),
		WithLogger(suite.mockLogger),
	)

}

func (suite *EtherscanSuite) TestNewEtherscanInputPlugin() {
	nep := NewEtherscanInputPlugin()
	assert.NotNil(suite.T(), nep)
	assert.IsType(suite.T(), &http.Client{}, nep.client)
	assert.IsType(suite.T(), &models.Logger{}, nep.log)
	assert.IsType(suite.T(), make(requestQueue), nep.networkRequests)
	assert.Equal(suite.T(), "", nep.Key)
	assert.Equal(suite.T(), _RequestTimeout, nep.Timeout)
	assert.Equal(suite.T(), _RequestLimit, nep.RequestLimit)
	assert.Equal(suite.T(), _RequestInterval, nep.RequestInterval)
	assert.IsType(suite.T(), make([]balance, 0), nep.Balances)
}

func (suite *EtherscanSuite) TestNewEtherscanInputPluginOptions() {
	nep := NewEtherscanInputPlugin(
		WithHTTPClient(suite.mockHTTPClient),
		WithLogger(suite.mockLogger),
	)
	assert.Equal(suite.T(), suite.mockHTTPClient, nep.client)
	assert.Equal(suite.T(), suite.mockLogger, nep.log)
}

func (suite *EtherscanSuite) TestDescription() {
	desc := suite.inputPlugin.Description()
	assert.Equal(suite.T(), _Description, desc)
}

func (suite *EtherscanSuite) TestSampleConfig() {
	sampleConfig := suite.inputPlugin.SampleConfig()
	assert.Equal(suite.T(), _ExampleEtherscanConfig, sampleConfig)
}

func (suite *EtherscanSuite) TestSetupParseRequestIntervalError() {
	suite.inputPlugin.RequestInterval = "^&*"
	err := suite.inputPlugin.Setup()
	assert.Error(suite.T(), err)
}

func (suite *EtherscanSuite) TestSetupParseTimeoutError() {
	suite.inputPlugin.Timeout = "%$#"
	err := suite.inputPlugin.Setup()
	assert.Error(suite.T(), err)
}

func (suite *EtherscanSuite) TestBurstLimiter() {
	testFunc := func() bool {
		<-suite.inputPlugin.burstLimiter
		return true
	}

	err := suite.inputPlugin.Setup()
	assert.Nil(suite.T(), err)
	assert.Eventually(suite.T(), testFunc, 10*time.Second, 1*time.Second)
}

func (suite *EtherscanSuite) TestSetupInvalidNetworkError() {
	testBalance := balance{
		Network: "Bart",
		Address: "0x0000000000000000000000000000000000000000",
		Tag:     "Simpsons",
	}

	suite.inputPlugin.Balances = append(suite.inputPlugin.Balances, testBalance)
	err := suite.inputPlugin.Setup()
	assert.Error(suite.T(), err)
}

func (suite *EtherscanSuite) TestSetupSingleBalanceAdd() {
	testNetwork, err := NetworkLookup("Mainnet")
	assert.Nil(suite.T(), err)
	testBalance := balance{
		Network: testNetwork.String(),
		Address: "0x0000000000000000000000000000000000000000",
		Tag:     "Marge",
	}

	suite.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return(nil)

	suite.inputPlugin.Balances = append(suite.inputPlugin.Balances, testBalance)
	err = suite.inputPlugin.Setup()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.inputPlugin.networkRequests.Len(), 1)

	gatheredTags := suite.inputPlugin.networkRequests[testNetwork][0].Tags()
	assert.Len(suite.T(), gatheredTags, 1)
}

func (suite *EtherscanSuite) TestSetupMultiBalanceAdd() {
	suite.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return(nil)
	testNetwork, err := NetworkLookup("Mainnet")
	assert.Nil(suite.T(), err)
	testBalances := []balance{
		{
			Network: testNetwork.String(),
			Address: "0x0000000000000000000000000000000000000001",
			Tag:     "Homer",
		},
		{
			Network: testNetwork.String(),
			Address: "0x0000000000000000000000000000000000000002",
			Tag:     "Marge",
		},
		{
			Network: testNetwork.String(),
			Address: "0x0000000000000000000000000000000000000003",
			Tag:     "Lisa",
		},
	}

	suite.inputPlugin.Balances = testBalances
	err = suite.inputPlugin.Setup()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, suite.inputPlugin.networkRequests[testNetwork].Len())
	assert.Equal(suite.T(), 3, len(suite.inputPlugin.Balances))

	gatheredTags := suite.inputPlugin.networkRequests[testNetwork][0].Tags()
	assert.Len(suite.T(), gatheredTags, 3)
}

func (suite *EtherscanSuite) TestSetupBalanceAddressCountExhausted() {
	suite.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return(nil)
	testNetwork, err := NetworkLookup("Mainnet")
	assert.Nil(suite.T(), err)

	testBalances := make([]balance, 0)
	for i := 0; i < 45; i++ {
		nba := balance{
			Network: testNetwork.String(),
			Address: "0x0000000000000000000000000000000000000000",
			Tag:     strconv.Itoa(i),
		}
		testBalances = append(testBalances, nba)
	}

	suite.inputPlugin.Balances = testBalances
	assert.Len(suite.T(), testBalances, 45)

	err = suite.inputPlugin.Setup()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, suite.inputPlugin.networkRequests[testNetwork].Len())

	addrList := suite.inputPlugin.networkRequests[testNetwork][0].(*BalanceRequest).RequestAddressList()
	assert.Equal(suite.T(), 20, len(addrList))

	addrList2 := suite.inputPlugin.networkRequests[testNetwork][1].(*BalanceRequest).RequestAddressList()
	assert.Equal(suite.T(), 20, len(addrList2))

	addrList3 := suite.inputPlugin.networkRequests[testNetwork][2].(*BalanceRequest).RequestAddressList()
	assert.Equal(suite.T(), 5, len(addrList3))
}

func (suite *EtherscanSuite) TestSetupAddAddressError() {
	testNetwork, err := NetworkLookup("Mainnet")
	assert.Nil(suite.T(), err)
	testBalance := balance{
		Network: testNetwork.String(),
		Address: "0x98765",
		Tag:     "Marge",
	}
	suite.inputPlugin.Balances = append(suite.inputPlugin.Balances, testBalance)

	testURL, _ := url.Parse("test")
	tbr := NewBalanceRequest(testURL, "", suite.mockLogger)
	query := tbr.URL().Query()
	query.Set("action", "bart")
	tbr.URL().RawQuery = query.Encode()

	suite.inputPlugin.networkRequests[testNetwork] = append(suite.inputPlugin.networkRequests[testNetwork], tbr)

	err = suite.inputPlugin.Setup()
	assert.Error(suite.T(), err)
}

func (suite *EtherscanSuite) TestSetupNewBalanceReqAddAddressError() {
	testNetwork, err := NetworkLookup("Mainnet")
	assert.Nil(suite.T(), err)
	testBalance := balance{
		Network: testNetwork.String(),
		Address: "1234",
		Tag:     "Marge",
	}
	suite.inputPlugin.Balances = append(suite.inputPlugin.Balances, testBalance)

	err = suite.inputPlugin.Setup()
	assert.Error(suite.T(), err)
}

func (suite *EtherscanSuite) TestPrepareInterval() {
	testRequestCount := 5
	suite.inputPlugin.requestCount = testRequestCount
	suite.mockLogger.On("Warnf", mock.Anything, testRequestCount).Return(nil)
	err := suite.inputPlugin.PrepareInterval()
	assert.Nil(suite.T(), err)
}

func (suite *EtherscanSuite) TestPrepareIntervalRequestCountSet() {
	assert.Len(suite.T(), suite.inputPlugin.networkRequests, 0)

	suite.inputPlugin.networkRequests = requestQueue{
		Network(0): requestList{
			&BalanceRequest{
				log:           suite.mockLogger,
				lastTriggered: time.Time{},
				request:       nil,
				tags:          nil,
			},
		},
	}

	err := suite.inputPlugin.PrepareInterval()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, suite.inputPlugin.requestCount)
}

func (suite *EtherscanSuite) TestFetch() {
	suite.inputPlugin.requestCount = 1
	testResponseBalances := make(map[string]interface{})
	testResponseTags := make(map[string]string)
	testKey := "test"
	testValue := 0
	testResponseBalances[testKey] = testValue
	mockRequest := MockRequest{}
	mockAccumulator := mocks.Accumulator{}

	mockRequest.On("MarkTriggered").Return(nil)
	mockRequest.On("Send", mock.Anything).Return(testResponseBalances, nil)
	suite.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return(nil)
	mockRequest.On("Tags").Return(testResponseTags)
	mockAccumulator.On("AddFields", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.inputPlugin.Fetch(&mockRequest, &mockAccumulator)
	assert.Equal(suite.T(), 0, suite.inputPlugin.requestCount)
}

func (suite *EtherscanSuite) TestFetchAddError() {
	suite.inputPlugin.requestCount = 1
	mockRequest := MockRequest{}
	mockAccumulator := mocks.Accumulator{}
	testErr := errors.New("test")

	mockRequest.On("MarkTriggered").Return(nil)
	mockRequest.On("Send", mock.Anything).Return(nil, testErr)
	mockAccumulator.On("AddError", mock.Anything).Return(nil)
	suite.inputPlugin.Fetch(&mockRequest, &mockAccumulator)
	assert.Equal(suite.T(), 0, suite.inputPlugin.requestCount)
}

func (suite *EtherscanSuite) TestGather() {
	suite.mockLogger.On("Debugf", mock.Anything, mock.Anything).Return(nil)
	suite.mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	testNetwork, err := NetworkLookup("Mainnet")
	assert.Nil(suite.T(), err)

	mockAccumulator := mocks.Accumulator{}

	nep := NewEtherscanInputPlugin(
		WithLogger(suite.mockLogger),
		WithHTTPClient(suite.mockHTTPClient),
	)

	testRecorder := httptest.NewRecorder()
	testMessage := fmt.Sprint(`{"status":"0","message":"OK","result":"0"}`)
	testRecorder.Body = bytes.NewBufferString(testMessage)

	suite.mockHTTPClient.On("Get", mock.Anything).Return(testRecorder.Result(), nil)
	suite.mockHTTPClient.On("Get", mock.Anything).Return(testRecorder.Result(), nil)

	testBalances := []balance{
		{
			Network: testNetwork.String(),
			Address: "0x0000000000000000000000000000000000000001",
			Tag:     "Homer",
		},
		{
			Network: testNetwork.String(),
			Address: "0x0000000000000000000000000000000000000002",
			Tag:     "Marge",
		},
	}

	nep.Balances = testBalances

	err = nep.Gather(&mockAccumulator)
	assert.Nil(suite.T(), err)
}

func (suite *EtherscanSuite) TestGatherSetupError() {
	mockAccumulator := mocks.Accumulator{}
	mockAccumulator.On("AddError", mock.Anything).Return(nil)
	nep := NewEtherscanInputPlugin(
		WithLogger(suite.mockLogger),
		WithHTTPClient(suite.mockHTTPClient),
	)

	nep.Timeout = "%$#"
	err := nep.Gather(&mockAccumulator)
	assert.Nil(suite.T(), err)
}
