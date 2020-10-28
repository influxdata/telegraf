package timestream

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"

	"github.com/stretchr/testify/assert"
)

func TestGetTimestreamTime(t *testing.T) {
	assertions := assert.New(t)

	tWithNanos := time.Date(2020, time.November, 10, 23, 44, 20, 123, time.UTC)
	tWithMicros := time.Date(2020, time.November, 10, 23, 44, 20, 123000, time.UTC)
	tWithMillis := time.Date(2020, time.November, 10, 23, 44, 20, 123000000, time.UTC)
	tOnlySeconds := time.Date(2020, time.November, 10, 23, 44, 20, 0, time.UTC)

	tUnitNanos, tValueNanos := getTimestreamTime(tWithNanos)
	assertions.Equal("NANOSECONDS", tUnitNanos)
	assertions.Equal("1605051860000000123", tValueNanos)

	tUnitMicros, tValueMicros := getTimestreamTime(tWithMicros)
	assertions.Equal("MICROSECONDS", tUnitMicros)
	assertions.Equal("1605051860000123", tValueMicros)

	tUnitMillis, tValueMillis := getTimestreamTime(tWithMillis)
	assertions.Equal("MILLISECONDS", tUnitMillis)
	assertions.Equal("1605051860123", tValueMillis)

	tUnitSeconds, tValueSeconds := getTimestreamTime(tOnlySeconds)
	assertions.Equal("SECONDS", tUnitSeconds)
	assertions.Equal("1605051860", tValueSeconds)
}

func TestPartitionRecords(t *testing.T) {

	assertions := assert.New(t)

	testDatum := timestreamwrite.Record{
		MeasureName:      aws.String("Foo"),
		MeasureValueType: aws.String("DOUBLE"),
		MeasureValue:     aws.String("123"),
	}

	var zeroDatum []*timestreamwrite.Record
	oneDatum := []*timestreamwrite.Record{&testDatum}
	twoDatum := []*timestreamwrite.Record{&testDatum, &testDatum}
	threeDatum := []*timestreamwrite.Record{&testDatum, &testDatum, &testDatum}

	assertions.Equal([][]*timestreamwrite.Record{}, partitionRecords(2, zeroDatum))
	assertions.Equal([][]*timestreamwrite.Record{oneDatum}, partitionRecords(2, oneDatum))
	assertions.Equal([][]*timestreamwrite.Record{oneDatum}, partitionRecords(2, oneDatum))
	assertions.Equal([][]*timestreamwrite.Record{twoDatum}, partitionRecords(2, twoDatum))
	assertions.Equal([][]*timestreamwrite.Record{twoDatum, oneDatum}, partitionRecords(2, threeDatum))
}

func TestConvertValueSupported(t *testing.T) {
	intInputValues := []interface{}{-1, int8(-2), int16(-3), int32(-4), int64(-5)}
	intOutputValues := []string{"-1", "-2", "-3", "-4", "-5"}
	intOutputValueTypes := []string{"BIGINT", "BIGINT", "BIGINT", "BIGINT", "BIGINT"}
	testConvertValueSupportedCases(t, intInputValues, intOutputValues, intOutputValueTypes)

	uintInputValues := []interface{}{uint(1), uint8(2), uint16(3), uint32(4), uint64(5)}
	uintOutputValues := []string{"1", "2", "3", "4", "5"}
	uintOutputValueTypes := []string{"BIGINT", "BIGINT", "BIGINT", "BIGINT", "BIGINT"}
	testConvertValueSupportedCases(t, uintInputValues, uintOutputValues, uintOutputValueTypes)

	otherInputValues := []interface{}{"foo", float32(22.123), 22.1234, true}
	otherOutputValues := []string{"foo", "22.123", "22.1234", "true"}
	otherOutputValueTypes := []string{"VARCHAR", "DOUBLE", "DOUBLE", "BOOLEAN"}
	testConvertValueSupportedCases(t, otherInputValues, otherOutputValues, otherOutputValueTypes)
}

func TestConvertValueUnsupported(t *testing.T) {
	assertions := assert.New(t)
	_, _, ok := convertValue(time.Date(2020, time.November, 10, 23, 44, 20, 0, time.UTC))
	assertions.False(ok, "Expected unsuccessful conversion")
}

func testConvertValueSupportedCases(t *testing.T,
	inputValues []interface{}, outputValues []string, outputValueTypes []string) {
	assertions := assert.New(t)
	for i, inputValue := range inputValues {
		v, vt, ok := convertValue(inputValue)
		assertions.Equal(true, ok, "Expected successful conversion")
		assertions.Equal(outputValues[i], v, "Expected different string representation of converted value")
		assertions.Equal(outputValueTypes[i], vt, "Expected different value type of converted value")
	}
}
