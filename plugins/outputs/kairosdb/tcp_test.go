package kairosdb

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func TestTcpWriteNormal(t *testing.T) {
	now := time.Now()
	nowUnix := toUnixtime(now)
	conn := &mockTCPConn{}
	subject := &tcpOutput{conn: conn}
	m1, _ := metric.New("name123", nil, map[string]interface{}{"field1": 0.5, "field2": 5}, now)
	m2, _ := metric.New("name234", nil, map[string]interface{}{"field2": 6}, now)
	err := subject.Write([]telegraf.Metric{m1, m2})
	require.NoError(t, err)
	require.Len(t, conn.reported, 3)
	require.Contains(t, conn.reported, "put name123.field2 "+nowUnix+" 5\n")
	require.Contains(t, conn.reported, "put name123.field1 "+nowUnix+" 0.5\n")
	require.Contains(t, conn.reported, "put name234.field2 "+nowUnix+" 6\n")
}

func TestTcpWriteNoMetrics(t *testing.T) {
	conn := &mockTCPConn{}
	subject := &tcpOutput{conn: conn}
	err := subject.Write([]telegraf.Metric{})
	require.NoError(t, err)
	require.Empty(t, conn.reported)
}

func TestTcpWriteUnsupportedType(t *testing.T) {
	now := time.Now()
	nowUnix := toUnixtime(now)
	conn := &mockTCPConn{}
	subject := &tcpOutput{conn: conn}
	m1, _ := metric.New("name123", nil, map[string]interface{}{"field1": true, "field2": 5}, now)
	err := subject.Write([]telegraf.Metric{m1})
	require.NoError(t, err)
	require.Len(t, conn.reported, 1)
	require.Contains(t, conn.reported, "put name123.field2 "+nowUnix+" 5\n")
}

func TestTcpWriteStringIgnored(t *testing.T) {
	now := time.Now()
	nowUnix := toUnixtime(now)
	conn := &mockTCPConn{}
	subject := &tcpOutput{conn: conn}
	m1, _ := metric.New("name123", nil, map[string]interface{}{"field1": "ignored", "field2": 5}, now)
	err := subject.Write([]telegraf.Metric{m1})
	require.NoError(t, err)
	require.Len(t, conn.reported, 1)
	require.Contains(t, conn.reported, "put name123.field2 "+nowUnix+" 5\n")
}

func TestTcpWriteError(t *testing.T) {
	conn := &mockTCPConn{writeError: errors.New("err1")}
	subject := &tcpOutput{conn: conn, connectionLost: make(chan struct{}, 1)}
	m1, _ := metric.New("name123", nil, map[string]interface{}{"field1": "ignored", "field2": 5}, time.Now())
	err := subject.Write([]telegraf.Metric{m1})
	require.Error(t, err)
}

type mockTCPConn struct {
	reported   []string
	writeError error
}

func (*mockTCPConn) Close() error {
	return nil
}

func (c *mockTCPConn) Write(b []byte) (_ int, _ error) {
	c.reported = append(c.reported, string(b))
	return 0, c.writeError
}

func toUnixtime(t time.Time) string {
	return strconv.FormatInt(t.UnixNano()/int64(time.Millisecond/time.Nanosecond), 10)
}

func TestPostedName(t *testing.T) {
	m, _ := metric.New("name123", nil, map[string]interface{}{"ignored": 0}, time.Now())
	require.Equal(t, postedName(m, "value"), "name123")

	m, _ = metric.New("name123", nil, map[string]interface{}{"ignored": 0}, time.Now())
	require.Equal(t, postedName(m, "field45"), "name123.field45")
}

func TestFormat(t *testing.T) {
	ints := []interface{}{345, (int32)(345), (int64)(345)}
	for _, v := range ints {
		actual, err := format(v)
		require.NoError(t, err)
		require.Equal(t, "345", actual)
	}

	floats := []interface{}{(float32)(345.4), (float64)(345.4)}
	for _, v := range floats {
		actual, err := format(v)
		require.NoError(t, err)
		require.Equal(t, "345.4", actual)
	}

	_, err := format("str")
	require.Error(t, err)
}
