package test

import (
	"encoding/base64"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"github.com/kentik/libkflow/api"
)

func NewClientServer() (*api.Client, *Server, *api.Device, error) {
	var (
		email  = randstr(8)
		token  = randstr(8)
		device = &api.Device{
			ID:          int(nextid()),
			Name:        randstr(8),
			IP:          net.ParseIP("127.0.0.1"),
			MaxFlowRate: 10,
			CompanyID:   int(rand.Uint32()),
		}
	)

	if ifs, err := net.Interfaces(); err == nil && len(ifs) > 0 {
		if addrs, err := ifs[0].Addrs(); err == nil && len(addrs) > 0 {
			addr := addrs[rand.Intn(len(addrs))]
			if ip, _, err := net.ParseCIDR(addr.String()); err == nil {
				device.IP = ip
			}
		}
	}

	server, err := NewServer("127.0.0.1", 0, false, true)
	if err != nil {
		return nil, nil, nil, err
	}
	go server.Serve(email, token, device)

	client := api.NewClient(api.ClientConfig{
		Email:   email,
		Token:   token,
		Timeout: 1 * time.Second,
		API:     server.URL(API),
		Proxy:   nil,
	})

	return client, server, device, nil
}

func randstr(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func nextid() uint64 {
	return atomic.AddUint64(&counter, 1)
}

var counter uint64 = 0
