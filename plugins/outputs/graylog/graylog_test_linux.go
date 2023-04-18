//go:build !windows && !darwin

package graylog

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestWriteUDP(t *testing.T) {
	tests := []struct {
		name              string
		namefieldnoprefix bool
	}{
		{
			name: "default without scheme",
		},
		{
			name: "UDP",
		},
		{
			name:              "UDP non-standard name field",
			namefieldnoprefix: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			address := UDPServer(t, &wg, tt.namefieldnoprefix)
			plugin := Graylog{
				NameFieldNoPrefix: tt.namefieldnoprefix,
				Servers:           []string{"udp://" + address},
			}
			require.NoError(t, plugin.Connect())
			defer plugin.Close()
			defer wg.Wait()

			metrics := testutil.MockMetrics()

			// UDP scenario:
			// 4 messages are send
			require.NoError(t, plugin.Write(metrics))
			require.NoError(t, plugin.Write(metrics))
			require.NoError(t, plugin.Write(metrics))
			require.NoError(t, plugin.Write(metrics))
		})
	}
}

func TestWriteTCP(t *testing.T) {
	pki := testutil.NewPKI("../../../testutil/pki")
	tlsClientConfig := pki.TLSClientConfig()
	tlsServerConfig, err := pki.TLSServerConfig().TLSConfig()
	require.NoError(t, err)

	tests := []struct {
		name         string
		tlsClientCfg tlsint.ClientConfig
	}{
		{
			name: "TCP",
		},
		{
			name: "TLS",
			tlsClientCfg: tlsint.ClientConfig{
				ServerName: "localhost",
				TLSCA:      tlsClientConfig.TLSCA,
				TLSKey:     tlsClientConfig.TLSKey,
				TLSCert:    tlsClientConfig.TLSCert,
			},
		},
		{
			name: "TLS no validation",
			tlsClientCfg: tlsint.ClientConfig{
				InsecureSkipVerify: true,
				ServerName:         "localhost",
				TLSKey:             tlsClientConfig.TLSKey,
				TLSCert:            tlsClientConfig.TLSCert,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			errs := make(chan error)
			address := TCPServer(t, &wg, tlsServerConfig, errs)

			plugin := Graylog{
				ClientConfig: tlsint.ClientConfig{
					InsecureSkipVerify: true,
					ServerName:         "localhost",
					TLSKey:             tlsClientConfig.TLSKey,
					TLSCert:            tlsClientConfig.TLSCert,
				},
				Servers: []string{"tcp://" + address},
			}
			require.NoError(t, plugin.Connect())
			defer plugin.Close()
			defer wg.Wait()

			metrics := testutil.MockMetrics()

			// TCP scenario:
			// 4 messages are send
			// -> connection gets forcefully broken after the 2nd message (server closes connection)
			// -> the 3rd write fails with error
			// -> during the 4th write connection is restored and write is successful

			require.NoError(t, plugin.Write(metrics))
			require.NoError(t, plugin.Write(metrics))
			require.NoError(t, <-errs)
			require.ErrorContains(t, plugin.Write(metrics), "error writing message")
			require.NoError(t, plugin.Write(metrics))
		})
	}
}

type GelfObject map[string]interface{}

func UDPServer(t *testing.T, wg *sync.WaitGroup, namefieldnoprefix bool) string {
	udpServer, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	recv := func() error {
		bufR := make([]byte, 1024)
		n, _, err := udpServer.ReadFrom(bufR)
		if err != nil {
			return err
		}

		b := bytes.NewReader(bufR[0:n])
		r, err := zlib.NewReader(b)
		if err != nil {
			return err
		}

		var maxDecompressionSize int64 = 500 * 1024 * 1024
		bufW := bytes.NewBuffer(nil)
		written, err := io.CopyN(bufW, r, maxDecompressionSize)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		} else if written == maxDecompressionSize {
			return fmt.Errorf("size of decoded data exceeds allowed size %d", maxDecompressionSize)
		}

		err = r.Close()
		if err != nil {
			return err
		}

		var obj GelfObject
		err = json.Unmarshal(bufW.Bytes(), &obj)
		if err != nil {
			return err
		}
		require.Equal(t, obj["short_message"], "telegraf")
		if namefieldnoprefix {
			require.Equal(t, obj["name"], "test1")
		} else {
			require.Equal(t, obj["_name"], "test1")
		}
		require.Equal(t, obj["_tag1"], "value1")
		require.Equal(t, obj["_value"], float64(1))

		return nil
	}

	// Send the address with the random port to the channel for the graylog instance to use it
	address := udpServer.LocalAddr().String()
	wg.Add(1)
	go func() {
		defer udpServer.Close()
		defer wg.Done()

		// in UDP scenario all 4 messages are received
		require.NoError(t, recv())
		require.NoError(t, recv())
		require.NoError(t, recv())
		require.NoError(t, recv())
	}()
	return address
}

func TCPServer(t *testing.T, wg *sync.WaitGroup, tlsConfig *tls.Config, errs chan error) string {
	tcpServer, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// Send the address with the random port to the channel for the graylog instance to use it
	address := tcpServer.Addr().String()

	accept := func() (net.Conn, error) {
		conn, err := tcpServer.Accept()
		require.NoError(t, err)
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			err = tcpConn.SetLinger(0)
			if err != nil {
				return nil, err
			}
		}
		err = conn.SetDeadline(time.Now().Add(15 * time.Second))
		if err != nil {
			return nil, err
		}
		if tlsConfig != nil {
			conn = tls.Server(conn, tlsConfig)
		}
		return conn, nil
	}

	recv := func(conn net.Conn) error {
		bufR := make([]byte, 1)
		bufW := bytes.NewBuffer(nil)
		for {
			n, err := conn.Read(bufR)
			if err != nil {
				return err
			}

			if n > 0 {
				if bufR[0] == 0 { // message delimiter found
					break
				}
				_, err = bufW.Write(bufR)
				if err != nil {
					return err
				}
			}
		}

		var obj GelfObject
		err = json.Unmarshal(bufW.Bytes(), &obj)
		require.NoError(t, err)
		require.Equal(t, obj["short_message"], "telegraf")
		require.Equal(t, obj["_name"], "test1")
		require.Equal(t, obj["_tag1"], "value1")
		require.Equal(t, obj["_value"], float64(1))
		return nil
	}

	wg.Add(1)
	go func() {
		defer tcpServer.Close()
		defer wg.Done()

		fmt.Println("server: opening connection")
		conn, err := accept()
		if err != nil {
			fmt.Println(err)
		}
		defer conn.Close()

		// in TCP scenario only 3 messages are received, the 3rd is lost due to simulated connection break after the 2nd

		fmt.Println("server: receving packet 1")
		err = recv(conn)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("server: receving packet 2")
		err = recv(conn)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("server: closing connection")
		err = conn.Close()
		if err != nil {
			fmt.Println(err)
		}

		errs <- err
		if err != nil {
			return
		}

		fmt.Println("server: re-opening connection")
		conn, err = accept()
		if err != nil {
			fmt.Println(err)
		}
		defer conn.Close()

		fmt.Println("server: receving packet 4")
		err = recv(conn)
		if err != nil {
			fmt.Println(err)
		}
	}()
	return address
}

func TestWriteUDPServerDown(t *testing.T) {
	dummy, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	plugin := Graylog{
		NameFieldNoPrefix: true,
		Servers:           []string{"udp://" + dummy.LocalAddr().String()},
		Log:               testutil.Logger{},
	}
	require.NoError(t, dummy.Close())
	require.NoError(t, plugin.Connect())
}

func TestWriteUDPServerUnavailableOnWrite(t *testing.T) {
	dummy, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	plugin := Graylog{
		NameFieldNoPrefix: true,
		Servers:           []string{"udp://" + dummy.LocalAddr().String()},
		Log:               testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())
	require.NoError(t, dummy.Close())
	require.NoError(t, plugin.Write(testutil.MockMetrics()))
}

func TestWriteTCPServerDown(t *testing.T) {
	dummy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	plugin := Graylog{
		NameFieldNoPrefix: true,
		Servers:           []string{"tcp://" + dummy.Addr().String()},
		Log:               testutil.Logger{},
	}
	require.NoError(t, dummy.Close())
	require.ErrorContains(t, plugin.Connect(), "connect: connection refused")
}

func TestWriteTCPServerUnavailableOnWrite(t *testing.T) {
	dummy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	plugin := Graylog{
		NameFieldNoPrefix: true,
		Servers:           []string{"tcp://" + dummy.Addr().String()},
		Log:               testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())
	require.NoError(t, dummy.Close())
	err = plugin.Write(testutil.MockMetrics())
	require.ErrorContains(t, err, "error writing message")
}

func TestWriteUDPServerDownRetry(t *testing.T) {
	dummy, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	plugin := Graylog{
		NameFieldNoPrefix: true,
		Servers:           []string{"udp://" + dummy.LocalAddr().String()},
		Reconnection:      true,
		Log:               testutil.Logger{},
	}
	require.NoError(t, dummy.Close())
	require.NoError(t, plugin.Connect())
	require.NoError(t, plugin.Close())
}

func TestWriteUDPServerUnavailableOnWriteRetry(t *testing.T) {
	dummy, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	plugin := Graylog{
		NameFieldNoPrefix: true,
		Servers:           []string{"udp://" + dummy.LocalAddr().String()},
		Reconnection:      true,
		Log:               testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())
	require.NoError(t, dummy.Close())
	err = plugin.Write(testutil.MockMetrics())
	require.ErrorContains(t, err, "not connected")
	require.NoError(t, plugin.Close())
}

func TestWriteTCPServerDownRetry(t *testing.T) {
	dummy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	logger := &testutil.CaptureLogger{}
	plugin := Graylog{
		NameFieldNoPrefix: true,
		Servers:           []string{"tcp://" + dummy.Addr().String()},
		Reconnection:      true,
		ReconnectionTime:  config.Duration(100 * time.Millisecond),
		Log:               logger,
	}
	require.NoError(t, dummy.Close())
	require.NoError(t, plugin.Connect())
	require.Eventually(t, func() bool {
		return strings.Contains(logger.LastError(), "after attempt #5...")
	}, 5*time.Second, 100*time.Millisecond)
	require.NoError(t, plugin.Close())
}

func TestWriteTCPServerUnavailableOnWriteRetry(t *testing.T) {
	dummy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	plugin := Graylog{
		NameFieldNoPrefix: true,
		Servers:           []string{"tcp://" + dummy.Addr().String()},
		Reconnection:      true,
		Log:               testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())
	require.NoError(t, dummy.Close())
	err = plugin.Write(testutil.MockMetrics())
	require.ErrorContains(t, err, "not connected")
	require.NoError(t, plugin.Close())
}

func TestWriteTCPRetryStopping(t *testing.T) {
	dummy, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	logger := &testutil.CaptureLogger{}
	plugin := Graylog{
		NameFieldNoPrefix: true,
		Servers:           []string{"tcp://" + dummy.Addr().String()},
		Reconnection:      true,
		ReconnectionTime:  config.Duration(10 * time.Millisecond),
		Log:               logger,
	}
	require.NoError(t, dummy.Close())
	require.NoError(t, plugin.Connect())
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, plugin.Close())
}
