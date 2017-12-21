package kairosdb

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_KairosDB_initInnerOutput_http(t *testing.T) {
	subject := &KairosDB{Timeout: 2, Protocol: protocolHTTP, Address: "address1:9090", User: "user1", Password: "pwd1"}
	err := subject.initInnerOutput()
	require.NoError(t, err)
	require.IsType(t, &httpOutput{}, subject.innerOutput)

	output := subject.innerOutput.(*httpOutput)
	require.Equal(t, output.url, "http://address1:9090")
	require.Equal(t, output.timeout, 2*time.Second)
	require.Equal(t, output.user, "user1")
	require.Equal(t, output.password, "pwd1")
}

func Test_KairosDB_initInnerOutput_https(t *testing.T) {
	subject := &KairosDB{Timeout: 2, Protocol: protocolHTTPS, Address: "address1:9090", User: "user1", Password: "pwd1"}
	err := subject.initInnerOutput()
	require.NoError(t, err)
	require.IsType(t, &httpOutput{}, subject.innerOutput)

	output := subject.innerOutput.(*httpOutput)
	require.Equal(t, output.url, "https://address1:9090")
}

func Test_KairosDB_initInnerOutput_tcp(t *testing.T) {
	subject := &KairosDB{Timeout: 2, Protocol: protocolTCP, Address: "address1:9090"}
	err := subject.initInnerOutput()
	require.NoError(t, err)
	require.IsType(t, &tcpOutput{}, subject.innerOutput)

	output := subject.innerOutput.(*tcpOutput)
	require.Equal(t, output.address, "address1:9090")
	require.Equal(t, output.timeout, 2*time.Second)
}

func Test_KairosDB_initInnerOutput_unsupportedProtocol(t *testing.T) {
	subject := &KairosDB{Protocol: "foo"}
	err := subject.initInnerOutput()
	require.Error(t, err)
}
