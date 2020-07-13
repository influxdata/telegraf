package etherscan

import (
	"bytes"
	"errors"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/inputs/etherscan/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	testClient     *MockHTTPClient
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

	suite.log.On("Debugf", mock.Anything, mock.Anything).Return(nil)
	suite.log.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	suite.balanceRequest = testBalanceRequest
	suite.testClient = &MockHTTPClient{}
}

func (suite *BalanceRequestSuite) TestAddAddressToBalance() {
	testAddress := "0x0000000000000000000000000000000000000000"
	err := suite.balanceRequest.AddAddress(testAddress)
	assert.Nil(suite.T(), err)

	addresses := suite.balanceRequest.RequestAddressList()
	assert.Len(suite.T(), addresses, 1)
	assert.Contains(suite.T(), addresses, testAddress)

	suite.balanceRequest.URL()
}

func (suite *BalanceRequestSuite) TestAddAddressToMultiBalance() {
	testAddress := "0x0000000000000000000000000000000000000001"
	err := suite.balanceRequest.AddAddress(testAddress)
	assert.Nil(suite.T(), err)

	testAddress2 := "0x0000000000000000000000000000000000000002"
	err = suite.balanceRequest.AddAddress(testAddress2)
	assert.Nil(suite.T(), err)

	addresses := suite.balanceRequest.RequestAddressList()
	assert.Len(suite.T(), addresses, 2)
	assert.Contains(suite.T(), addresses, testAddress)
	assert.Contains(suite.T(), addresses, testAddress2)

	suite.balanceRequest.URL()
}

func (suite *BalanceRequestSuite) TestAddAddressCountExhausted() {
	for i := 0; i < 20; i++ {
		err := suite.balanceRequest.AddAddress("0x0000000000000000000000000000000000000000")
		assert.Nil(suite.T(), err)
	}

	testAddress := "0x0000000000000000000000000000000000000000"
	err := suite.balanceRequest.AddAddress(testAddress)
	assert.Equal(suite.T(), ErrBalanceAddressCountExhausted, err)
}

func (suite *BalanceRequestSuite) TestAddAddressUnknownAction() {
	query := suite.balanceRequest.URL().Query()
	query.Set("action", "bart")
	suite.balanceRequest.URL().RawQuery = query.Encode()

	testAddress := "0x0000000000000000000000000000000000000000"

	err := suite.balanceRequest.AddAddress(testAddress)
	assert.Error(suite.T(), err)
}

func (suite *BalanceRequestSuite) TestAddTag() {
	testAddress := "0x0000000000000000000000000000000000000000"
	testTag := "marge"
	suite.balanceRequest.AddTag(testAddress, testTag)

	assert.Contains(suite.T(), suite.balanceRequest.tags[testAddress], testTag)
}

func (suite *BalanceRequestSuite) TestUpdateURL() {
	testURL := "http://new.test.url"
	parsedURL, _ := url.Parse(testURL)

	suite.balanceRequest.UpdateURL(parsedURL)
	assert.Equal(suite.T(), suite.balanceRequest.URL(), parsedURL)
}

func (suite *BalanceRequestSuite) TestLastRequest() {
	testTime := time.Now()
	suite.balanceRequest.lastTriggered = testTime

	reqTime := suite.balanceRequest.LastRequest()
	assert.Equal(suite.T(), testTime, reqTime)
}

func (suite *BalanceRequestSuite) TestMarkTriggered() {
	suite.balanceRequest.MarkTriggered()
	assert.WithinDuration(suite.T(), time.Now(), suite.balanceRequest.LastRequest(), 2*time.Second)
}

func (suite *BalanceRequestSuite) TestTags() {
	testTags := map[string]string{
		"0x0000000000000000000000000000000000000001": "homer",
		"0x0000000000000000000000000000000000000002": "marge",
	}

	suite.balanceRequest.tags = testTags
	retrievedTags := suite.balanceRequest.Tags()
	assert.Equal(suite.T(), testTags, retrievedTags)
}

func (suite *BalanceRequestSuite) TestSendGetError() {
	testAddress := "0x0000000000000000000000000000000000000000"
	testErr := errors.New("test")
	suite.testClient.On("Get", mock.Anything).Return(nil, testErr)

	err := suite.balanceRequest.AddAddress(testAddress)
	assert.Nil(suite.T(), err)

	_, err = suite.balanceRequest.Send(suite.testClient)
	assert.Error(suite.T(), err)
}

func (suite *BalanceRequestSuite) TestSingleSend() {
	testAddress := "0x0000000000000000000000000000000000000000"
	_ = suite.balanceRequest.AddAddress(testAddress)

	testBalance := 0
	testRecorder := httptest.NewRecorder()
	testMessage := fmt.Sprintf(
		`{"status":"0","message":"OK","result":"%d"}`,
		testBalance,
	)
	testRecorder.Body = bytes.NewBufferString(testMessage)

	suite.testClient.On("Get", mock.Anything).Return(testRecorder.Result(), nil)
	resp, err := suite.balanceRequest.Send(suite.testClient)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), float64(testBalance), resp[testAddress])
}

func (suite *BalanceRequestSuite) TestMultiSend() {
	testAddress := "0x0000000000000000000000000000000000000001"
	err := suite.balanceRequest.AddAddress(testAddress)
	assert.Nil(suite.T(), err)

	testAddress2 := "0x0000000000000000000000000000000000000002"
	err = suite.balanceRequest.AddAddress(testAddress2)
	assert.Nil(suite.T(), err)

	testAddress3 := "0x0000000000000000000000000000000000000003"
	err = suite.balanceRequest.AddAddress(testAddress3)
	assert.Nil(suite.T(), err)

	testRecorder := httptest.NewRecorder()
	testMessage := fmt.Sprintf(
		`{"status":"0","message":"OK","result":[{"account":"%s","balance":"40891631566070000000000"},{"account":"%s","balance":"332567136222827062478"},{"account":"%s","balance":"0"}]}`,
		testAddress,
		testAddress2,
		testAddress3,
	)
	testRecorder.Body = bytes.NewBufferString(testMessage)

	suite.testClient.On("Get", mock.Anything).Return(testRecorder.Result(), nil)
	resp, err := suite.balanceRequest.Send(suite.testClient)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), resp)
}

func (suite *BalanceRequestSuite) TestDefaultSendError() {
	testRecorder := httptest.NewRecorder()
	testMessage := fmt.Sprint(`{"status":"0","message":"OK","result":"0"}`)
	testRecorder.Body = bytes.NewBufferString(testMessage)
	query := suite.balanceRequest.URL().Query()
	query.Set("action", "bart")
	suite.balanceRequest.URL().RawQuery = query.Encode()
	suite.testClient.On("Get", mock.Anything).Return(testRecorder.Result(), nil)

	resp, err := suite.balanceRequest.Send(suite.testClient)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), resp)
}

type errReader int

var testError = errors.New("test error")

func (errReader) Read(p []byte) (n int, err error) {
	return 0, testError
}

func (errReader) Close() error {
	return testError
}

func (suite *BalanceRequestSuite) TestSendRespBodyErrors() {
	testRecorder := httptest.NewRecorder()
	testMessage := fmt.Sprint(123)
	testRecorder.Body = bytes.NewBufferString(testMessage)
	testRecorder.Result().Body = errReader(1)
	suite.testClient.On("Get", mock.Anything).Return(testRecorder.Result(), nil)

	respBodyCloseErrMsg := fmt.Sprintf(
		"response body close error: %s", testError.Error(),
	)
	suite.log.On("Errorf", mock.Anything, mock.Anything).Return(respBodyCloseErrMsg)

	resp, err := suite.balanceRequest.Send(suite.testClient)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), resp)
	expectedError := testError.Error()
	assert.Equal(suite.T(), expectedError, err.Error())
}

func (suite *BalanceRequestSuite) TestAssignBalancesUnmarshalError() {
	mockResponder := MockValidatingBalanceResponder{}

	testRecorder := httptest.NewRecorder()
	testMessage := fmt.Sprintf("broken")
	testRecorder.Body = bytes.NewBufferString(testMessage)

	unmarshalErr := "invalid character 'b' looking for beginning of value"
	balanceReqProcessRespErr := fmt.Sprintf("error unmarshaling: [%s]", unmarshalErr)

	suite.testClient.On("Get", mock.Anything).Return(testRecorder.Result(), nil)
	resp, err := suite.balanceRequest.assignBalances(testRecorder.Body.Bytes(), &mockResponder)
	assert.Nil(suite.T(), resp)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), balanceReqProcessRespErr, err.Error())
}

func (suite *BalanceRequestSuite) TestAssignBalancesParseEthSumError() {
	mockResponder := MockValidatingBalanceResponder{}
	testBalance := "*"
	testAccountBalances := []AccountBalance{
		{
			Account: "abc",
			Balance: testBalance,
		},
	}
	mockResponder.On("ValidateResponseStatus").Return(nil)
	mockResponder.On("Result").Return(testAccountBalances)

	testRecorder := httptest.NewRecorder()
	testMessage := fmt.Sprintf(
		`{"status":"0","message":"OK","result":"%s"}`,
		testBalance,
	)
	testRecorder.Body = bytes.NewBufferString(testMessage)

	unmarshalErr := "strconv.ParseFloat: parsing \"*\": invalid syntax"
	balanceReqProcessRespErr := fmt.Sprintf("parsing balance [%s] error: [%s]", testBalance, unmarshalErr)

	resp, err := suite.balanceRequest.assignBalances(testRecorder.Body.Bytes(), &mockResponder)
	assert.Nil(suite.T(), resp)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), balanceReqProcessRespErr, err.Error())
}

func (suite *BalanceRequestSuite) TestSendResponseStatusParsingError() {
	mockResponder := MockValidatingBalanceResponder{}
	testError := errors.New("test error")
	mockResponder.On("ValidateResponseStatus").Return(testError)

	testStatus := "A"
	testRecorder := httptest.NewRecorder()
	testMessage := fmt.Sprintf(`{"status":"%s","message":"OK","result":"0"}`, testStatus)
	testRecorder.Body = bytes.NewBufferString(testMessage)

	respBodyCloseErrMsg := fmt.Errorf(
		"error parsing status: [%s]", testError,
	)
	resp, err := suite.balanceRequest.assignBalances(testRecorder.Body.Bytes(), &mockResponder)
	assert.Nil(suite.T(), resp)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), respBodyCloseErrMsg.Error(), err.Error())
}

func (suite *BalanceRequestSuite) TestSendParsingEthSumError() {
	resp, err := parseEthSum("*")
	testError := "strconv.ParseFloat: parsing \"*\": invalid syntax"
	assert.Equal(suite.T(), float64(-1), resp)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), testError, err.Error())
}

func (suite *BalanceRequestSuite) TestValidateReturnCodeError() {
	err := validateReturnCode(9)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrReturnCodeNotOK, err)
}

func (suite *BalanceRequestSuite) TestValidateReturnCode() {
	err := validateReturnCode(0)
	assert.Nil(suite.T(), err)
}

func (suite *BalanceRequestSuite) TestMultiValidateResponseStatus() {
	testMultiBalanceResponse := MultiBalanceResponse{
		Status: "0",
	}

	err := testMultiBalanceResponse.ValidateResponseStatus()
	assert.Nil(suite.T(), err)
}

func (suite *BalanceRequestSuite) TestMultiValidateResponseStatusCodeError() {
	testMultiBalanceResponse := MultiBalanceResponse{
		Status: "99",
	}

	err := testMultiBalanceResponse.ValidateResponseStatus()
	assert.Error(suite.T(), err)
}

func (suite *BalanceRequestSuite) TestMultiValidateResponseStatusParseErrorError() {
	testMultiBalanceResponse := MultiBalanceResponse{
		Status: "*",
	}

	err := testMultiBalanceResponse.ValidateResponseStatus()
	assert.Error(suite.T(), err)
}

func (suite *BalanceRequestSuite) TestMultiResult() {
	testAccountBalances := []AccountBalance{
		{
			Account: "0x0000000000000000000000000000000000000001",
			Balance: "123",
		},
		{
			Account: "0x0000000000000000000000000000000000000002",
			Balance: "987",
		},
	}
	testMultiBalanceResponse := MultiBalanceResponse{
		Results: testAccountBalances,
	}

	results := testMultiBalanceResponse.Result()
	assert.Equal(suite.T(), testAccountBalances, results)
}

func (suite *BalanceRequestSuite) TestSingleValidateResponseStatus() {
	testMultiBalanceResponse := SingleBalanceResponse{
		Status: "0",
	}

	err := testMultiBalanceResponse.ValidateResponseStatus()
	assert.Nil(suite.T(), err)
}

func (suite *BalanceRequestSuite) TestSingleValidateResponseStatusCodeError() {
	testMultiBalanceResponse := SingleBalanceResponse{
		Status: "99",
	}

	err := testMultiBalanceResponse.ValidateResponseStatus()
	assert.Error(suite.T(), err)
}

func (suite *BalanceRequestSuite) TestSingleValidateResponseStatusParseErrorError() {
	testMultiBalanceResponse := SingleBalanceResponse{
		Status: "*",
	}

	err := testMultiBalanceResponse.ValidateResponseStatus()
	assert.Error(suite.T(), err)
}
