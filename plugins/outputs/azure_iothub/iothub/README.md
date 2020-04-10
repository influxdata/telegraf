# iothub

Azure IoT Hub SDK for Golang, provides both device-to-cloud ([`iotdevice`](iotdevice)) and cloud-to-device ([`iotservice`](iotservice)) packages for end-to-end communication.

API is subject to change until `v1.0.0`. Bumping minor version may indicate breaking changes.

See [TODO](#todo) section to see what's missing in the library.

## Installation

To install the library as a dependency:

```bash
go get -u github.com/amenzhinsky/iothub
```

To install CLI applications:

```bash
GO111MODULE=on go get -u github.com/amenzhinsky/iothub/cmd/{iothub-service,iothub-device}
```

## Usage Example

Receive and print messages from IoT devices in a backend application:

```go
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
```

Send a message from an IoT device:

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/amenzhinsky/iothub/iotdevice"
	iotmqtt "github.com/amenzhinsky/iothub/iotdevice/transport/mqtt"
)

func main() {
	c, err := iotdevice.NewFromConnectionString(
		iotmqtt.New(), os.Getenv("IOTHUB_DEVICE_CONNECTION_STRING"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// connect to the iothub
	if err = c.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}

	// send a device-to-cloud message
	if err = c.SendEvent(context.Background(), []byte(`hello`)); err != nil {
		log.Fatal(err)
	}
}
```

[cmd/iothub-service](https://github.com/amenzhinsky/iothub/blob/master/cmd/iothub-service) and [cmd/iothub-device](https://github.com/amenzhinsky/iothub/blob/master/cmd/iothub-device) are reference implementations of almost all available features. 

## CLI

The project provides two command line utilities: `iothub-device` and `iothub-sevice`. First is for using it on IoT devices and the second manages and interacts with them. 

You can perform operations like publishing, subscribing to events and feedback, registering and invoking direct methods and so on straight from the command line.

`iothub-service` is a [iothub-explorer](https://github.com/Azure/iothub-explorer) replacement that can be distributed as a single binary opposed to a typical nodejs app.

See `-help` for more details.

## Testing

`TEST_IOTHUB_SERVICE_CONNECTION_STRING` is required for end-to-end testing, which is a shared access policy connection string with all permissions.

`TEST_EVENTHUB_CONNECTION_STRING` is required for `eventhub` package testing.

## TODO

### iotservice

1. Complete IoT Edge support
1. Stabilize API
1. Fix TODOs

### iotdevice

1. Device modules support.
1. HTTP transport (files uploading).
1. AMQP transport (batch sending, WS).

## Contributing

All contributions are welcome.
