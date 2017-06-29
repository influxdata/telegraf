package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/influxdata/telegraf/plugins/inputs/zipkin"
)

func main() {
	e := make(chan error)
	d := make(chan zipkin.SpanData)
	s := zipkin.NewHTTPServer(9411, e, d)
	go s.Serve()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	sigHandle(sigChan, s)

	for {
		select {
		case err := <-e:
			fmt.Println("error: ", err)
		case data := <-d:
			fmt.Println("Got zipkin data: %#+v", data)
		}
	}

}

func sigHandle(c chan os.Signal, server *zipkin.Server) {
	select {
	case <-c:
		fmt.Println("received SIGINT, stopping server")
		//server.Done <- struct{}{}
		server.Shutdown()
		server.CloseAllChannels()
	}
}
