package request_aggregates

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
	"time"
)

func TestRequestParser_ParseLine_Nanos(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "ns", IsTimeEpoch: true, TimePosition: 1}

	// Test format nanoseconds
	r, err := rp.ParseLine("1462380541003228260,123,\"thisissuccessful\"")
	require.NoError(t, err)
	require.Equal(t, time.Unix(0, 1462380541003228260), r.Timestamp)
}

func TestRequestParser_ParseLine_Micros(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "us", IsTimeEpoch: true, TimePosition: 1}

	// Test format nanoseconds
	r, err := rp.ParseLine("1462380541003228,123,\"thisissuccessful\"")
	require.NoError(t, err)
	require.Equal(t, time.Unix(0, 1462380541003228000), r.Timestamp)
}

func TestRequestParser_ParseLine_Milis(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "ms", IsTimeEpoch: true, TimePosition: 1}

	// Test format nanoseconds
	r, err := rp.ParseLine("1462380541003,123,\"thisissuccessful\"")
	require.NoError(t, err)
	require.Equal(t, time.Unix(0, 1462380541003000000), r.Timestamp)
}

func TestRequestParser_ParseLine_Seconds(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 1}

	// Test format nanoseconds
	r, err := rp.ParseLine("1462380541,123,\"thisissuccessful\"")
	require.NoError(t, err)
	require.Equal(t, time.Unix(1462380541, 0), r.Timestamp)
}

func TestRequestParser_ParseLine_WrongUnit(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 1}

	// Test format nanoseconds
	_, err := rp.ParseLine("1462380541003228260,123,\"thisissuccessful\"")
	require.Error(t, err)
}

func TestRequestParser_ParseLine_Layout(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: time.RFC3339Nano,
		IsTimeEpoch: false, TimePosition: 1}

	// Test format nanoseconds
	r, err := rp.ParseLine("2006-01-02T15:04:05.999999999Z,123,\"thisissuccessful\"")
	require.NoError(t, err)
	parsed, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z")
	require.Equal(t, parsed, r.Timestamp)
}

func TestRequestParser_ParseLine_WrongLayout(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: time.RFC3339Nano,
		IsTimeEpoch: false, TimePosition: 1}

	// Test format nanoseconds
	_, err := rp.ParseLine("2006-01-02T15:04:05,123,\"thisissuccessful\"")
	require.Error(t, err)
}

func TestRequestParser_ParseLine_Int(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 1}

	// Test format nanoseconds
	r, err := rp.ParseLine("1462380541,123,\"thisissuccessful\"")
	require.NoError(t, err)
	require.Equal(t, float64(123), r.Time)
}

func TestRequestParser_ParseLine_Float(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 1}

	// Test format nanoseconds
	r, err := rp.ParseLine("1462380541,123.45,\"thisissuccessful\"")
	require.NoError(t, err)
	require.Equal(t, float64(123.45), r.Time)
}

func TestRequestParser_ParseLine_NoRegexp(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 1}

	// Test format nanoseconds
	r, err := rp.ParseLine("1462380541,123.45,\"thisissuccessful\"")
	require.NoError(t, err)
	require.Equal(t, false, r.Failure)
}

func TestRequestParser_ParseLine_Success(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 1,
		ResultPosition: 2, SuccessRegexp: regexp.MustCompile(".*success.*")}

	// Test format nanoseconds
	r, err := rp.ParseLine("1462380541,123.45,\"thisissuccessful\"")
	require.NoError(t, err)
	require.Equal(t, false, r.Failure)
}

func TestRequestParser_ParseLine_Failure(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 1,
		ResultPosition: 2, SuccessRegexp: regexp.MustCompile(".*success.*")}

	// Test format nanoseconds
	r, err := rp.ParseLine("1462380541,123.45,\"thisonefailed\"")
	require.NoError(t, err)
	require.Equal(t, true, r.Failure)
}

func TestRequestParser_ParseLine_TimestampOutOfBounds(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 6, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 1}

	// Test format nanoseconds
	_, err := rp.ParseLine("1462380541,123.45,\"thisissuccessful\"")
	require.Error(t, err)
}

func TestRequestParser_ParseLine_TimeOutOfBounds(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 6}

	// Test format nanoseconds
	_, err := rp.ParseLine("1462380541,123.45,\"thisissuccessful\"")
	require.Error(t, err)
}

func TestRequestParser_ParseLine_SuccessOutOfBounds(t *testing.T) {
	rp := &RequestParser{TimestampPosition: 0, TimestampFormat: "s", IsTimeEpoch: true, TimePosition: 1,
		ResultPosition: 8, SuccessRegexp: regexp.MustCompile(".*success.*")}

	// Test format nanoseconds
	_, err := rp.ParseLine("1462380541,123.45,\"thisissuccessful\"")
	require.Error(t, err)
}
