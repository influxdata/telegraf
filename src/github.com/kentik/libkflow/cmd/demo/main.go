package main

import (
	"log"
	"net"
	"time"

	"github.com/kentik/libkflow"
	"github.com/kentik/libkflow/flow"
)

func main() {
	var (
		email    = "test@example.com"
		token    = "token"
		deviceID = 1
		host     = net.ParseIP("127.0.0.1")
		port     = 8080
		program  = "demo"
		version  = "0.0.1"
	)

	errors := make(chan error, 100)

	config := libkflow.NewConfig(email, token, program, version)
	config.SetServer(host, port)
	config.SetVerbose(1)

	s, err := libkflow.NewSenderWithDeviceID(deviceID, errors, config)
	if err != nil {
		log.Fatal(err)
	}

	s.Send(&flow.Flow{
		Ipv4SrcAddr: uint32(0),
		Ipv4DstAddr: uint32(1),
	})

	s.Stop(10 * time.Second)
}
