package proxy

import (
	"net"
	"testing"
	"time"

	"github.com/armon/go-socks5"
	"github.com/stretchr/testify/require"
)

func TestSocks5ProxyConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const (
		proxyAddress  = "0.0.0.0:12345"
		proxyUsername = "user"
		proxyPassword = "password"
	)

	l, err := net.Listen("tcp", "0.0.0.0:0")
	require.NoError(t, err)

	server, err := socks5.New(&socks5.Config{
		AuthMethods: []socks5.Authenticator{socks5.UserPassAuthenticator{
			Credentials: socks5.StaticCredentials{
				proxyUsername: proxyPassword,
			},
		}},
	})
	require.NoError(t, err)

	go func() { require.NoError(t, server.ListenAndServe("tcp", proxyAddress)) }()

	conf := Socks5ProxyConfig{
		Socks5ProxyEnabled:  true,
		Socks5ProxyAddress:  proxyAddress,
		Socks5ProxyUsername: proxyUsername,
		Socks5ProxyPassword: proxyPassword,
	}
	dialer, err := conf.GetDialer()
	require.NoError(t, err)

	var proxyConn net.Conn
	for i := 0; i < 10; i++ {
		proxyConn, err = dialer.Dial("tcp", l.Addr().String())
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NotNil(t, proxyConn)
	defer func() { require.NoError(t, proxyConn.Close()) }()

	serverConn, err := l.Accept()
	require.NoError(t, err)
	defer func() { require.NoError(t, serverConn.Close()) }()

	writePayload := []byte("test")
	_, err = proxyConn.Write(writePayload)
	require.NoError(t, err)

	receivePayload := make([]byte, 4)
	_, err = serverConn.Read(receivePayload)
	require.NoError(t, err)

	require.Equal(t, writePayload, receivePayload)
}
