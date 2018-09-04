package network // import "collectd.org/network"

import (
	"context"
	"log"
	"net"
	"os"

	"collectd.org/format"
)

// This example demonstrates how to listen to encrypted network traffic and
// dump it to STDOUT using format.Putval.
func ExampleServer_ListenAndWrite() {
	srv := &Server{
		Addr:           net.JoinHostPort("::", DefaultService),
		Writer:         format.NewPutval(os.Stdout),
		PasswordLookup: NewAuthFile("/etc/collectd/users"),
	}

	// blocks
	log.Fatal(srv.ListenAndWrite(context.Background()))
}

// This example demonstrates how to forward received IPv6 multicast traffic to
// a unicast address, using PSK encryption.
func ExampleListenAndWrite() {
	opts := ClientOptions{
		SecurityLevel: Encrypt,
		Username:      "collectd",
		Password:      "dieXah7e",
	}
	client, err := Dial(net.JoinHostPort("example.com", DefaultService), opts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// blocks
	log.Fatal(ListenAndWrite(context.Background(), ":"+DefaultService, client))
}
