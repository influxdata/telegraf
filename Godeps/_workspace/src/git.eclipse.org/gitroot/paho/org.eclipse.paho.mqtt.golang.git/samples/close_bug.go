package main

import (
	"fmt"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

func main() {
	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetCleanSession(true)

	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("plz mosquitto goes down now")
	time.Sleep(5 * time.Second)

	c.Disconnect(200)
	time.Sleep(5 * time.Second)
}
