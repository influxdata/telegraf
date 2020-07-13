package etherscan

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/suite"
)

func TestEtherscanRequestSuites(t *testing.T) {
	suite.Run(t, new(RequestListSuite))
}

type RequestListSuite struct {
	suite.Suite
	requestA *MockRequest
	requestB *MockRequest
	requestC *MockRequest

	testRequestList requestList
}

type _ suite.SetupTestSuite

func (suite *RequestListSuite) SetupTest() {
	suite.requestA = &MockRequest{}
	suite.requestB = &MockRequest{}
	suite.requestC = &MockRequest{}

	suite.testRequestList = requestList{
		suite.requestA,
		suite.requestB,
		suite.requestC,
	}
}

func (suite *RequestListSuite) TestNewRequestList() {
	assert.Equal(suite.T(), suite.testRequestList.Len(), 3)
}

func (suite *RequestListSuite) TestSwap() {
	suite.testRequestList.Swap(0, 2)
	assert.Equal(suite.T(), suite.testRequestList[0], suite.requestC)
	assert.Equal(suite.T(), suite.testRequestList[2], suite.requestA)

	suite.testRequestList.Swap(1, 0)
	assert.Equal(suite.T(), suite.testRequestList[1], suite.requestC)
	assert.Equal(suite.T(), suite.testRequestList[0], suite.requestB)
}

func (suite *RequestListSuite) TestLess() {
	assert.Equal(suite.T(), suite.testRequestList[0], suite.requestA)
	assert.Equal(suite.T(), suite.testRequestList[1], suite.requestB)
	assert.Equal(suite.T(), suite.testRequestList[2], suite.requestC)

	// 0
	suite.requestA.On("LastRequest").Return(time.Now())
	// 1
	suite.requestB.On("LastRequest").Return(time.Now().Add(10 * time.Second))
	// 2
	suite.requestC.On("LastRequest").Return(time.Now().Add(5 * time.Second))

	isAfter := suite.testRequestList.Less(0, 1)
	assert.False(suite.T(), isAfter)
	isAfter = suite.testRequestList.Less(0, 2)
	assert.False(suite.T(), isAfter)
	isAfter = suite.testRequestList.Less(1, 2)
	assert.True(suite.T(), isAfter)
}

func (suite *RequestListSuite) TestSort() {
	assert.Equal(suite.T(), suite.testRequestList[0], suite.requestA)
	assert.Equal(suite.T(), suite.testRequestList[1], suite.requestB)
	assert.Equal(suite.T(), suite.testRequestList[2], suite.requestC)

	suite.requestA.On("LastRequest").Return(time.Now())
	suite.requestB.On("LastRequest").Return(time.Now().Add(10 * time.Second))
	suite.requestC.On("LastRequest").Return(time.Now().Add(5 * time.Second))

	sort.Sort(suite.testRequestList)

	assert.Equal(suite.T(), suite.testRequestList[0], suite.requestB)
	assert.Equal(suite.T(), suite.testRequestList[1], suite.requestC)
	assert.Equal(suite.T(), suite.testRequestList[2], suite.requestA)
}
