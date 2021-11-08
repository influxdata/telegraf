package graylog

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
	reuse "github.com/libp2p/go-reuseport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteUDP(t *testing.T) {
	tests := []struct {
		name     string
		instance Graylog
	}{
		{
			name: "default without scheme",
			instance: Graylog{
				Servers: []string{"127.0.0.1:12201"},
			},
		},
		{
			name: "UDP",
			instance: Graylog{
				Servers: []string{"udp://127.0.0.1:12201"},
			},
		},
		{
			name: "UDP non-standard name field",
			instance: Graylog{
				Servers:           []string{"udp://127.0.0.1:12201"},
				NameFieldNoPrefix: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			var wg2 sync.WaitGroup
			wg.Add(1)
			wg2.Add(1)
			go UDPServer(t, &wg, &wg2, &tt.instance)
			wg2.Wait()

			i := tt.instance
			err := i.Connect()
			require.NoError(t, err)
			defer i.Close()
			defer wg.Wait()

			metrics := testutil.MockMetrics()

			// UDP scenario:
			// 4 messages are send

			err = i.Write(metrics)
			require.NoError(t, err)
			err = i.Write(metrics)
			require.NoError(t, err)
			err = i.Write(metrics)
			require.NoError(t, err)
			err = i.Write(metrics)
			require.NoError(t, err)
		})
	}
}

func TestWriteTCP(t *testing.T) {
	pki := testutil.NewPKI("../../../testutil/pki")
	tlsClientConfig := pki.TLSClientConfig()
	tlsServerConfig, err := pki.TLSServerConfig().TLSConfig()
	require.NoError(t, err)

	tests := []struct {
		name            string
		instance        Graylog
		tlsServerConfig *tls.Config
	}{
		{
			name: "TCP",
			instance: Graylog{
				Servers: []string{"tcp://127.0.0.1:12201"},
			},
		},
		{
			name: "TLS",
			instance: Graylog{
				Servers: []string{"tcp://127.0.0.1:12201"},
				ClientConfig: tlsint.ClientConfig{
					ServerName: "localhost",
					TLSCA:      tlsClientConfig.TLSCA,
					TLSKey:     tlsClientConfig.TLSKey,
					TLSCert:    tlsClientConfig.TLSCert,
				},
			},
			tlsServerConfig: tlsServerConfig,
		},
		{
			name: "TLS no validation",
			instance: Graylog{
				Servers: []string{"tcp://127.0.0.1:12201"},
				ClientConfig: tlsint.ClientConfig{
					InsecureSkipVerify: true,
					ServerName:         "localhost",
					TLSKey:             tlsClientConfig.TLSKey,
					TLSCert:            tlsClientConfig.TLSCert,
				},
			},
			tlsServerConfig: tlsServerConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			var wg2 sync.WaitGroup
			var wg3 sync.WaitGroup
			wg.Add(1)
			wg2.Add(1)
			wg3.Add(1)
			go TCPServer(t, &wg, &wg2, &wg3, tt.tlsServerConfig)
			wg2.Wait()

			i := tt.instance
			err = i.Connect()
			require.NoError(t, err)
			defer i.Close()
			defer wg.Wait()

			metrics := testutil.MockMetrics()

			// TCP scenario:
			// 4 messages are send
			// -> connection gets forcefully broken after the 2nd message (server closes connection)
			// -> the 3rd write fails with error
			// -> during the 4th write connection is restored and write is successful

			err = i.Write(metrics)
			require.NoError(t, err)
			err = i.Write(metrics)
			require.NoError(t, err)
			wg3.Wait()
			err = i.Write(metrics)
			require.Error(t, err)
			err = i.Write(metrics)
			require.NoError(t, err)
		})
	}
}

type GelfObject map[string]interface{}

func UDPServer(t *testing.T, wg *sync.WaitGroup, wg2 *sync.WaitGroup, config *Graylog) {
	udpServer, err := reuse.ListenPacket("udp", "127.0.0.1:12201")
	require.NoError(t, err)
	defer udpServer.Close()
	defer wg.Done()
	wg2.Done()

	recv := func() {
		bufR := make([]byte, 1024)
		n, _, err := udpServer.ReadFrom(bufR)
		require.NoError(t, err)

		b := bytes.NewReader(bufR[0:n])
		r, _ := zlib.NewReader(b)

		bufW := bytes.NewBuffer(nil)
		_, _ = io.Copy(bufW, r)
		_ = r.Close()

		var obj GelfObject
		_ = json.Unmarshal(bufW.Bytes(), &obj)
		require.NoError(t, err)
		assert.Equal(t, obj["short_message"], "telegraf")
		if config.NameFieldNoPrefix {
			assert.Equal(t, obj["name"], "test1")
		} else {
			assert.Equal(t, obj["_name"], "test1")
		}
		assert.Equal(t, obj["_tag1"], "value1")
		assert.Equal(t, obj["_value"], float64(1))
	}

	// in UDP scenario all 4 messages are received

	recv()
	recv()
	recv()
	recv()
}

func TCPServer(t *testing.T, wg *sync.WaitGroup, wg2 *sync.WaitGroup, wg3 *sync.WaitGroup, tlsConfig *tls.Config) {
	tcpServer, err := reuse.Listen("tcp", "127.0.0.1:12201")
	require.NoError(t, err)
	defer tcpServer.Close()
	defer wg.Done()
	wg2.Done()

	accept := func() net.Conn {
		conn, err := tcpServer.Accept()
		require.NoError(t, err)
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			_ = tcpConn.SetLinger(0)
		}
		_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
		if tlsConfig != nil {
			conn = tls.Server(conn, tlsConfig)
		}
		return conn
	}

	recv := func(conn net.Conn) {
		bufR := make([]byte, 1)
		bufW := bytes.NewBuffer(nil)
		for {
			n, err := conn.Read(bufR)
			require.NoError(t, err)
			if n > 0 {
				if bufR[0] == 0 { // message delimiter found
					break
				}
				_, _ = bufW.Write(bufR)
			}
		}

		var obj GelfObject
		err = json.Unmarshal(bufW.Bytes(), &obj)
		require.NoError(t, err)
		assert.Equal(t, obj["short_message"], "telegraf")
		assert.Equal(t, obj["_name"], "test1")
		assert.Equal(t, obj["_tag1"], "value1")
		assert.Equal(t, obj["_value"], float64(1))
	}

	conn := accept()
	defer conn.Close()

	// in TCP scenario only 3 messages are received, the 3rd is lost due to simulated connection break after the 2nd

	recv(conn)
	recv(conn)
	_ = conn.Close()
	wg3.Done()
	conn = accept()
	defer conn.Close()
	recv(conn)
}
