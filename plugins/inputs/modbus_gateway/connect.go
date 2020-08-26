package modbus_gateway

import (
	mb "github.com/wz2b/modbus"
	"net"
	"net/url"
)

func connect(m *ModbusGateway) error {
	u, err := url.Parse(m.Gateway)
	if err != nil {
		return err
	}
	var host, port string
	host, port, err = net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}
	m.tcpHandler = mb.NewTCPClientHandler(host + ":" + port)
	m.tcpHandler.Timeout = m.Timeout.Duration
	m.client = mb.NewClient(m.tcpHandler)
	err = m.tcpHandler.Connect()
	if err != nil {
		return err
	}
	m.isConnected = true
	return err
}

func disconnect(m *ModbusGateway) error {
	m.tcpHandler.Close()
	return nil
}
