package main

import (
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

func main() {
	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883")
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	for {
		time.Sleep(1 * time.Second)
	}
}
