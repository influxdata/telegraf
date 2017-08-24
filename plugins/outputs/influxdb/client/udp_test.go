package client

import (
	"net"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUDPClient(t *testing.T) {
	config := UDPConfig{
		URL: "udp://localhost:8089",
	}
	client, err := NewUDP(config)
	assert.NoError(t, err)

	err = client.Query("ANY QUERY RETURNS NIL")
	assert.NoError(t, err)

	assert.NoError(t, client.Close())
}

func TestNewUDPClient_Errors(t *testing.T) {
	// url.Parse Error
	config := UDPConfig{
		URL: "udp://localhost%35:8089",
	}
	_, err := NewUDP(config)
	assert.Error(t, err)

	// ResolveUDPAddr Error
	config = UDPConfig{
		URL: "udp://localhost:999999",
	}
	_, err = NewUDP(config)
	assert.Error(t, err)
}

func TestUDPClient_Write(t *testing.T) {
	config := UDPConfig{
		URL: "udp://localhost:8199",
	}
	client, err := NewUDP(config)
	assert.NoError(t, err)

	packets := make(chan string, 100)
	address, err := net.ResolveUDPAddr("udp", "localhost:8199")
	assert.NoError(t, err)
	listener, err := net.ListenUDP("udp", address)
	defer listener.Close()
	assert.NoError(t, err)
	go func() {
		buf := make([]byte, 200)
		for {
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				packets <- err.Error()
			}
			packets <- string(buf[0:n])
		}
	}()

	// test sending simple metric
	n, err := client.Write([]byte("cpu value=99\n"))
	assert.Equal(t, n, 13)
	assert.NoError(t, err)
	pkt := <-packets
	assert.Equal(t, "cpu value=99\n", pkt)

	wp := WriteParams{}
	//
	// Using WriteStream() & a metric.Reader:
	config3 := UDPConfig{
		URL:         "udp://localhost:8199",
		PayloadSize: 40,
	}
	client3, err := NewUDP(config3)
	assert.NoError(t, err)

	now := time.Unix(1484142942, 0)
	m1, _ := metric.New("test", map[string]string{},
		map[string]interface{}{"value": 1.1}, now)
	m2, _ := metric.New("test", map[string]string{},
		map[string]interface{}{"value": 1.1}, now)
	m3, _ := metric.New("test", map[string]string{},
		map[string]interface{}{"value": 1.1}, now)
	ms := []telegraf.Metric{m1, m2, m3}
	mReader := metric.NewReader(ms)
	n, err = client3.WriteStreamWithParams(mReader, 10, wp)
	// 3 metrics at 35 bytes each (including the newline)
	assert.Equal(t, 105, n)
	assert.NoError(t, err)
	pkt = <-packets
	assert.Equal(t, "test value=1.1 1484142942000000000\n", pkt)
	pkt = <-packets
	assert.Equal(t, "test value=1.1 1484142942000000000\n", pkt)
	pkt = <-packets
	assert.Equal(t, "test value=1.1 1484142942000000000\n", pkt)

	assert.NoError(t, client.Close())

	config = UDPConfig{
		URL:         "udp://localhost:8199",
		PayloadSize: 40,
	}
	client4, err := NewUDP(config)
	assert.NoError(t, err)

	ts := time.Unix(1484142943, 0)
	m1, _ = metric.New("test", map[string]string{},
		map[string]interface{}{"this_is_a_very_long_field_name": 1.1}, ts)
	m2, _ = metric.New("test", map[string]string{},
		map[string]interface{}{"value": 1.1}, ts)
	ms = []telegraf.Metric{m1, m2}
	reader := metric.NewReader(ms)
	n, err = client4.WriteStream(reader, 0)
	assert.NoError(t, err)
	require.Equal(t, 35, n)
	assert.NoError(t, err)
	pkt = <-packets
	assert.Equal(t, "test value=1.1 1484142943000000000\n", pkt)

	assert.NoError(t, client4.Close())
}
