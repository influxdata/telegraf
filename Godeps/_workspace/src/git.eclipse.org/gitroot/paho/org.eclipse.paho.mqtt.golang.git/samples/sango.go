package main

import (
	"fmt"
	"os"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

var f MQTT.MessageHandler = func(client *MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

var onConnect MQTT.OnConnectHandler = func(client *MQTT.Client) {
	fmt.Println("onConnect")
	if token := client.Subscribe("shirou@github/#", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

var subscribed = "#"

func main() {
	opts := MQTT.NewClientOptions().AddBroker("tcp://lite.mqtt.shiguredo.jp:1883")
	opts.SetDefaultPublishHandler(f)
	opts.SetOnConnectHandler(onConnect)
	opts.SetCleanSession(true)

	opts.SetUsername("shirou@github")
	opts.SetPassword("8Ub6F68kfYlr7RoV")

	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	qos := 0
	retain := false
	payload := "sanple"
	topic := "shirou@github/log"
	token := c.Publish(topic, byte(qos), retain, payload)
	//	token.Wait()
	fmt.Println("%v", token.Error())

	for {
		time.Sleep(1 * time.Second)
	}
}
