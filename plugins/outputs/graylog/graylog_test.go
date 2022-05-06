package graylog

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/metric"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
)

func TestSerializer(t *testing.T) {
	m1 := metric.New("testing",
		map[string]string{
			"verb": "GET",
			"host": "hostname",
		},
		map[string]interface{}{
			"full_message":  "full",
			"short_message": "short",
			"level":         "1",
			"facility":      "demo",
			"line":          "42",
			"file":          "graylog.go",
		},
		time.Now(),
	)

	graylog := Graylog{}
	result, err := graylog.serialize(m1)

	require.NoError(t, err)

	for _, r := range result {
		var obj GelfObject
		err = json.Unmarshal([]byte(r), &obj)
		require.NoError(t, err)

		require.Equal(t, obj["version"], "1.1")
		require.Equal(t, obj["_name"], "testing")
		require.Equal(t, obj["_verb"], "GET")
		require.Equal(t, obj["host"], "hostname")
		require.Equal(t, obj["full_message"], "full")
		require.Equal(t, obj["short_message"], "short")
		require.Equal(t, obj["level"], "1")
		require.Equal(t, obj["facility"], "demo")
		require.Equal(t, obj["line"], "42")
		require.Equal(t, obj["file"], "graylog.go")
	}
}

func TestWriteUDP(t *testing.T) {
	tests := []struct {
		name     string
		instance Graylog
	}{
		{
			name: "default without scheme",
		},
		{
			name: "UDP",
		},
		{
			name: "UDP non-standard name field",
			instance: Graylog{
				NameFieldNoPrefix: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)
			address := make(chan string, 1)
			errs := make(chan error)
			go UDPServer(t, &wg, &tt.instance, address, errs)
			require.NoError(t, <-errs)

			i := tt.instance
			i.Servers = []string{fmt.Sprintf("udp://%s", <-address)}
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
		},
		{
			name: "TLS",
			instance: Graylog{
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
			wg.Add(1)
			address := make(chan string, 1)
			errs := make(chan error)
			fmt.Println("test: staring TCP server")
			go TCPServer(t, &wg, tt.tlsServerConfig, address, errs)
			require.NoError(t, <-errs)

			i := tt.instance
			i.Servers = []string{fmt.Sprintf("tcp://%s", <-address)}
			fmt.Println("client: connecting to TCP server")
			err = i.Connect()
			require.NoError(t, err)
			fmt.Println("client: connected")
			defer i.Close()
			defer wg.Wait()

			metrics := testutil.MockMetrics()

			// TCP scenario:
			// 4 messages are send
			// -> connection gets forcefully broken after the 2nd message (server closes connection)
			// -> the 3rd write fails with error
			// -> during the 4th write connection is restored and write is successful

			fmt.Println("client: writting packet 1")
			err = i.Write(metrics)
			require.NoError(t, err)

			fmt.Println("client: writting packet 2")
			err = i.Write(metrics)
			require.NoError(t, err)

			fmt.Println("client: checking for errors")
			require.NoError(t, <-errs)

			fmt.Println("client: writting packet 3")
			err = i.Write(metrics)

			fmt.Println("client: writting packet 4")
			err = i.Write(metrics)
			require.NoError(t, err)
		})
	}
}

type GelfObject map[string]interface{}

func UDPServer(t *testing.T, wg *sync.WaitGroup, config *Graylog, address chan string, errs chan error) {
	udpServer, err := net.ListenPacket("udp", "127.0.0.1:0")
	errs <- err
	if err != nil {
		return
	}

	// Send the address with the random port to the channel for the graylog instance to use it
	address <- udpServer.LocalAddr().String()
	defer udpServer.Close()
	defer wg.Done()

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

		bufW := bytes.NewBuffer(nil)
		_, err = io.Copy(bufW, r)
		if err != nil {
			return err
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
		if config.NameFieldNoPrefix {
			require.Equal(t, obj["name"], "test1")
		} else {
			require.Equal(t, obj["_name"], "test1")
		}
		require.Equal(t, obj["_tag1"], "value1")
		require.Equal(t, obj["_value"], float64(1))

		return nil
	}

	// in UDP scenario all 4 messages are received

	err = recv()
	if err != nil {
		fmt.Println(err)
	}
	err = recv()
	if err != nil {
		fmt.Println(err)
	}
	err = recv()
	if err != nil {
		fmt.Println(err)
	}
	err = recv()
	if err != nil {
		fmt.Println(err)
	}
}

func TCPServer(t *testing.T, wg *sync.WaitGroup, tlsConfig *tls.Config, address chan string, errs chan error) {
	tcpServer, err := net.Listen("tcp", "127.0.0.1:0")
	errs <- err
	if err != nil {
		return
	}

	// Send the address with the random port to the channel for the graylog instance to use it
	address <- tcpServer.Addr().String()
	defer tcpServer.Close()
	defer wg.Done()

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
}
