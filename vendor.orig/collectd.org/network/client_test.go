package network // import "collectd.org/network"

import (
	"context"
	"log"
	"net"
	"time"

	"collectd.org/api"
)

func ExampleClient() {
	ctx := context.Background()
	conn, err := Dial(net.JoinHostPort("example.com", DefaultService), ClientOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	vl := &api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "golang",
			Type:   "gauge",
		},
		Time:     time.Now(),
		Interval: 10 * time.Second,
		Values:   []api.Value{api.Gauge(42.0)},
	}

	if err := conn.Write(ctx, vl); err != nil {
		log.Fatal(err)
	}
}
