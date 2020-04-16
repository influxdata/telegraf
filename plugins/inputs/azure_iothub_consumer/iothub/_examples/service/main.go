package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/amenzhinsky/iothub/iotservice"
)

func main() {
	c, err := iotservice.NewFromConnectionString(
		os.Getenv("IOTHUB_SERVICE_CONNECTION_STRING"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// subscribe to device-to-cloud events
	log.Fatal(c.SubscribeEvents(context.Background(), func(msg *iotservice.Event) error {
		fmt.Printf("%q sends %q", msg.ConnectionDeviceID, msg.Payload)
		return nil
	}))
}
