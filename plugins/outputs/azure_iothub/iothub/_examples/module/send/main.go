package main

import (
	"context"
	"log"
	"os"

	"github.com/amenzhinsky/iothub/iotdevice"
	iotmqtt "github.com/amenzhinsky/iothub/iotdevice/transport/mqtt"
)

func main() {

	cs := "HostName=myiothub.azure-devices.net;DeviceId=mydevice;ModuleId=mymodule;SharedAccessKey=MyAcc355K3y!=" // replace with primary module-specific connection string from IoT Hub
	gwhn := os.Getenv("IOTEDGE_GATEWAYHOSTNAME") // when running on edge device
	mgid := os.Getenv("IOTEDGE_MODULEGENERATIONID") // when running on edge device
	wluri := os.Getenv("IOTEDGE_WORKLOADURI") // when running on edge device

	c, err := iotdevice.NewModuleFromConnectionString(
		// <transport>, <connection string>, <gateway hostname>, <module gen id>, <iotedge workload uri>, <use iotedge gateway for connection>,
		iotmqtt.NewModuleTransport(), cs, gwhn, mgid, wluri, true, 
	)
	if err != nil {
		log.Fatal(err)
	}

	// connect to the iothub
	if err = c.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}

	// send a device-to-cloud message
	if err = c.SendEvent(context.Background(), []byte("hello")); err != nil {
		log.Fatal(err)
	}
}
