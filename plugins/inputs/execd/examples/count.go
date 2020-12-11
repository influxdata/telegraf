package main

// Example using HUP signaling

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	counter := 0

	for {
		<-c

		fmt.Printf("counter_go count=%d\n", counter)
		counter++
	}
}
