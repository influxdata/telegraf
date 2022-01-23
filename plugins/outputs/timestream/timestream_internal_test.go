package timestream

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
)

func TestGetTimestreamTime(t *testing.T) {
	tWithNanos := time.Date(2020, time.November, 10, 23, 44, 20, 123, time.UTC)
	tWithMicros := time.Date(2020, time.November, 10, 23, 44, 20, 123000, time.UTC)
	tWithMillis := time.Date(2020, time.November, 10, 23, 44, 20, 123000000, time.UTC)
	tOnlySeconds := time.Date(2020, time.November, 10, 23, 44, 20, 0, time.UTC)

	tUnitNanos, tValueNanos := getTimestreamTime(tWithNanos)
	require.Equal(t, types.TimeUnitNanoseconds, tUnitNanos)
	require.Equal(t, "1605051860000000123", tValueNanos)

	tUnitMicros, tValueMicros := getTimestreamTime(tWithMicros)
	require.Equal(t, types.TimeUnitMicroseconds, tUnitMicros)
	require.Equal(t, "1605051860000123", tValueMicros)

	tUnitMillis, tValueMillis := getTimestreamTime(tWithMillis)
	require.Equal(t, types.TimeUnitMilliseconds, tUnitMillis)
	require.Equal(t, "1605051860123", tValueMillis)

	tUnitSeconds, tValueSeconds := getTimestreamTime(tOnlySeconds)
	require.Equal(t, types.TimeUnitSeconds, tUnitSeconds)
	require.Equal(t, "1605051860", tValueSeconds)
}

func TestPartitionRecords(t *testing.T) {
	testDatum := types.Record{
		MeasureName:      aws.String("Foo"),
		MeasureValueType: types.MeasureValueTypeDouble,
		MeasureValue:     aws.String("123"),
	}

	var zeroDatum []types.Record
	oneDatum := []types.Record{testDatum}
	twoDatum := []types.Record{testDatum, testDatum}
	threeDatum := []types.Record{testDatum, testDatum, testDatum}

	require.Equal(t, [][]types.Record{}, partitionRecords(2, zeroDatum))
	require.Equal(t, [][]types.Record{oneDatum}, partitionRecords(2, oneDatum))
	require.Equal(t, [][]types.Record{oneDatum}, partitionRecords(2, oneDatum))
	require.Equal(t, [][]types.Record{twoDatum}, partitionRecords(2, twoDatum))
	require.Equal(t, [][]types.Record{twoDatum, oneDatum}, partitionRecords(2, threeDatum))
}

func TestConvertValueSupported(t *testing.T) {
	intInputValues := []interface{}{-1, int8(-2), int16(-3), int32(-4), int64(-5)}
	intOutputValues := []string{"-1", "-2", "-3", "-4", "-5"}
	intOutputValueTypes := []types.MeasureValueType{types.MeasureValueTypeBigint, types.MeasureValueTypeBigint, types.MeasureValueTypeBigint, types.MeasureValueTypeBigint, types.MeasureValueTypeBigint}
	testConvertValueSupportedCases(t, intInputValues, intOutputValues, intOutputValueTypes)

	uintInputValues := []interface{}{uint(1), uint8(2), uint16(3), uint32(4), uint64(5)}
	uintOutputValues := []string{"1", "2", "3", "4", "5"}
	uintOutputValueTypes := []types.MeasureValueType{types.MeasureValueTypeBigint, types.MeasureValueTypeBigint, types.MeasureValueTypeBigint, types.MeasureValueTypeBigint, types.MeasureValueTypeBigint}
	testConvertValueSupportedCases(t, uintInputValues, uintOutputValues, uintOutputValueTypes)

	otherInputValues := []interface{}{"foo", float32(22.123), 22.1234, true}
	otherOutputValues := []string{"foo", "22.123", "22.1234", "true"}
	otherOutputValueTypes := []types.MeasureValueType{types.MeasureValueTypeVarchar, types.MeasureValueTypeDouble, types.MeasureValueTypeDouble, types.MeasureValueTypeBoolean}
	testConvertValueSupportedCases(t, otherInputValues, otherOutputValues, otherOutputValueTypes)
}

func TestConvertValueUnsupported(t *testing.T) {
	_, _, ok := convertValue(time.Date(2020, time.November, 10, 23, 44, 20, 0, time.UTC))
	require.False(t, ok, "Expected unsuccessful conversion")
}

func testConvertValueSupportedCases(t *testing.T,
	inputValues []interface{}, outputValues []string, outputValueTypes []types.MeasureValueType) {
	for i, inputValue := range inputValues {
		v, vt, ok := convertValue(inputValue)
		require.Equal(t, true, ok, "Expected successful conversion")
		require.Equal(t, outputValues[i], v, "Expected different string representation of converted value")
		require.Equal(t, outputValueTypes[i], vt, "Expected different value type of converted value")
	}
}
