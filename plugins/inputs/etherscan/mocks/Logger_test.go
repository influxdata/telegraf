package mocks

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestLoggerSuites(t *testing.T) {
	suite.Run(t, new(TestLoggerSuite))
}

type TestLoggerSuite struct {
	suite.Suite
	testLogger Logger
}

type _ suite.SetupTestSuite

func (suite *TestLoggerSuite) SetupTest() {
	suite.testLogger = Logger{}
}

func (suite *TestLoggerSuite) TestLogger_Debug() {
	suite.testLogger.On("Debug", mock.Anything).Return(nil)
	suite.testLogger.Debug("test")
}

func (suite *TestLoggerSuite) TestLogger_Error() {
	suite.testLogger.On("Error", mock.Anything).Return(nil)
	suite.testLogger.Error("test")
}

func (suite *TestLoggerSuite) TestLogger_Info() {
	suite.testLogger.On("Info", mock.Anything).Return(nil)
	suite.testLogger.Info("test")
}

func (suite *TestLoggerSuite) TestLogger_Infof() {
	suite.testLogger.On("Infof", mock.Anything, mock.Anything).Return(nil)
	suite.testLogger.Infof("test", "test2")
}

func (suite *TestLoggerSuite) TestLogger_Warn() {
	suite.testLogger.On("Warn", mock.Anything).Return(nil)
	suite.testLogger.Warn("test")
}
