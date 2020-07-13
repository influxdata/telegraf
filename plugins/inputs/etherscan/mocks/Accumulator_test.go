package mocks

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/mock"
)

func TestAccumulatorSuites(t *testing.T) {
	suite.Run(t, new(TestAccumulatorSuite))
}

type TestAccumulatorSuite struct {
	suite.Suite
	testAccumulator Accumulator
	testFields      map[string]interface{}
	testTags        map[string]string
	testTime        time.Time
	testMeasurement string
}

type _ suite.SetupTestSuite

func (suite *TestAccumulatorSuite) SetupTest() {
	suite.testAccumulator = Accumulator{}
	suite.testMeasurement = "test"
	suite.testFields = make(map[string]interface{})
	suite.testFields["test"] = 0
	suite.testFields["test2"] = 1
	suite.testTags = make(map[string]string)
	suite.testTags["blah"] = "test"
	suite.testTags["fa"] = "la"
	suite.testTime = time.Now()
}

func (suite *TestAccumulatorSuite) TestAccumulator_AddFields() {
	suite.testAccumulator.On("AddFields", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.testAccumulator.AddFields(suite.testMeasurement, suite.testFields, suite.testTags, suite.testTime)
}

func (suite *TestAccumulatorSuite) TestAccumulator_AddCounter() {
	suite.testAccumulator.On("AddCounter", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.testAccumulator.AddCounter(suite.testMeasurement, suite.testFields, suite.testTags, suite.testTime)
}

func (suite *TestAccumulatorSuite) TestAccumulator_AddGauge() {
	suite.testAccumulator.On("AddGauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.testAccumulator.AddGauge(suite.testMeasurement, suite.testFields, suite.testTags, suite.testTime)
}

func (suite *TestAccumulatorSuite) TestAccumulator_AddHistogram() {
	suite.testAccumulator.On("AddHistogram", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.testAccumulator.AddHistogram(suite.testMeasurement, suite.testFields, suite.testTags, suite.testTime)
}

func (suite *TestAccumulatorSuite) TestAccumulator_AddMetric() {
	var testMetric telegraf.Metric
	suite.testAccumulator.On("AddMetric", mock.Anything)
	suite.testAccumulator.AddMetric(testMetric)
}

func (suite *TestAccumulatorSuite) TestAccumulator_AddSummary() {
	suite.testAccumulator.On("AddSummary", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.testAccumulator.AddSummary(suite.testMeasurement, suite.testFields, suite.testTags, suite.testTime)
}

func (suite *TestAccumulatorSuite) TestAccumulator_SetPrecision() {
	suite.testAccumulator.On("SetPrecision", mock.Anything).Return(nil)
	suite.testAccumulator.SetPrecision(1 * time.Millisecond)
}

func (suite *TestAccumulatorSuite) TestAccumulator_WithTracking() {
	suite.testAccumulator.On("WithTracking", mock.Anything).Return(nil)
	suite.testAccumulator.WithTracking(10)
}
