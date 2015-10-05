/*
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package mqtt

import "fmt"
import "time"
import "bytes"

import "io/ioutil"
import "crypto/tls"
import "crypto/x509"
import "testing"

func Test_Start(t *testing.T) {
	ops := NewClientOptions().SetClientID("Start").
		AddBroker(FVTTCP).
		SetStore(NewFileStore("/tmp/fvt/Start"))
	c := NewClient(ops)

	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", token.Error())
	}

	c.Disconnect(250)
}

/* uncomment this if you have connection policy disallowing FailClientID
func Test_InvalidConnRc(t *testing.T) {
	ops := NewClientOptions().SetClientID("FailClientID").
		AddBroker("tcp://" + FVT_IP + ":17003").
		SetStore(NewFileStore("/tmp/fvt/InvalidConnRc"))

	c := NewClient(ops)
	_, err := c.Connect()
	if err != ErrNotAuthorized {
		t.Fatalf("Did not receive error as expected, got %v", err)
	}
	c.Disconnect(250)
}
*/

// Helper function for Test_Start_Ssl
func NewTLSConfig() *tls.Config {
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile("samples/samplecerts/CAfile.pem")
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	cert, err := tls.LoadX509KeyPair("samples/samplecerts/client-crt.pem", "samples/samplecerts/client-key.pem")
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}
}

/* uncomment this if you have ssl setup
func Test_Start_Ssl(t *testing.T) {
	tlsconfig := NewTlsConfig()
	ops := NewClientOptions().SetClientID("StartSsl").
		AddBroker(FVT_SSL).
		SetStore(NewFileStore("/tmp/fvt/Start_Ssl")).
		SetTlsConfig(tlsconfig)

	c := NewClient(ops)

	_, err := c.Connect()
	if err != nil {
		t.Fatalf("Error on Client.Connect(): %v", err)
	}

	c.Disconnect(250)
}
*/

func Test_Publish_1(t *testing.T) {
	ops := NewClientOptions()
	ops.AddBroker(FVTTCP)
	ops.SetClientID("Publish_1")
	ops.SetStore(NewFileStore("/tmp/fvt/Publish_1"))

	c := NewClient(ops)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", token.Error())
	}

	c.Publish("test/Publish", 0, false, "Publish qo0")

	c.Disconnect(250)
}

func Test_Publish_2(t *testing.T) {
	ops := NewClientOptions()
	ops.AddBroker(FVTTCP)
	ops.SetClientID("Publish_2")
	ops.SetStore(NewFileStore("/tmp/fvt/Publish_2"))

	c := NewClient(ops)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", token.Error())
	}

	c.Publish("/test/Publish", 0, false, "Publish1 qos0")
	c.Publish("/test/Publish", 0, false, "Publish2 qos0")

	c.Disconnect(250)
}

func Test_Publish_3(t *testing.T) {
	ops := NewClientOptions()
	ops.AddBroker(FVTTCP)
	ops.SetClientID("Publish_3")
	ops.SetStore(NewFileStore("/tmp/fvt/Publish_3"))

	c := NewClient(ops)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", token.Error())
	}

	c.Publish("/test/Publish", 0, false, "Publish1 qos0")
	c.Publish("/test/Publish", 1, false, "Publish2 qos1")
	c.Publish("/test/Publish", 2, false, "Publish2 qos2")

	c.Disconnect(250)
}

func Test_Subscribe(t *testing.T) {
	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("Subscribe_tx")
	pops.SetStore(NewFileStore("/tmp/fvt/Subscribe/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("Subscribe_rx")
	sops.SetStore(NewFileStore("/tmp/fvt/Subscribe/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
	}
	sops.SetDefaultPublishHandler(f)
	s := NewClient(sops)

	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	s.Subscribe("/test/sub", 0, nil)

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}

	p.Publish("/test/sub", 0, false, "Publish qos0")

	p.Disconnect(250)
	s.Disconnect(250)
}

func Test_Will(t *testing.T) {
	willmsgc := make(chan string)

	sops := NewClientOptions().AddBroker(FVTTCP)
	sops.SetClientID("will-giver")
	sops.SetWill("/wills", "good-byte!", 0, false)
	sops.SetConnectionLostHandler(func(client *Client, err error) {
		fmt.Println("OnConnectionLost!")
	})
	c := NewClient(sops)

	wops := NewClientOptions()
	wops.AddBroker(FVTTCP)
	wops.SetClientID("will-subscriber")
	wops.SetStore(NewFileStore("/tmp/fvt/Will"))
	wops.SetDefaultPublishHandler(func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		willmsgc <- string(msg.Payload())
	})
	wsub := NewClient(wops)

	wToken := wsub.Connect()
	if wToken.Wait() && wToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", wToken.Error())
	}

	wsub.Subscribe("/wills", 0, nil)

	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", token.Error())
	}
	time.Sleep(time.Duration(1) * time.Second)

	c.forceDisconnect()

	wsub.Disconnect(250)

	if <-willmsgc != "good-byte!" {
		t.Fatalf("will message did not have correct payload")
	}
}

func Test_Binary_Will(t *testing.T) {
	willmsgc := make(chan []byte)
	will := []byte{
		0xDE,
		0xAD,
		0xBE,
		0xEF,
	}

	sops := NewClientOptions().AddBroker(FVTTCP)
	sops.SetClientID("will-giver")
	sops.SetBinaryWill("/wills", will, 0, false)
	sops.SetConnectionLostHandler(func(client *Client, err error) {
	})
	c := NewClient(sops)

	wops := NewClientOptions().AddBroker(FVTTCP)
	wops.SetClientID("will-subscriber")
	wops.SetStore(NewFileStore("/tmp/fvt/Binary_Will"))
	wops.SetDefaultPublishHandler(func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %v\n", msg.Payload())
		willmsgc <- msg.Payload()
	})
	wsub := NewClient(wops)

	wToken := wsub.Connect()
	if wToken.Wait() && wToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", wToken.Error())
	}

	wsub.Subscribe("/wills", 0, nil)

	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", token.Error())
	}
	time.Sleep(time.Duration(1) * time.Second)

	c.forceDisconnect()

	wsub.Disconnect(250)

	if !bytes.Equal(<-willmsgc, will) {
		t.Fatalf("will message did not have correct payload")
	}
}

/**
"[...] a publisher is responsible for determining the maximum QoS a
message can be delivered at, but a subscriber is able to downgrade
the QoS to one more suitable for its usage.
The QoS of a message is never upgraded."
**/

/***********************************
 * Tests to cover the 9 QoS combos *
 ***********************************/

func wait(c chan bool) {
	fmt.Println("choke is waiting")
	<-c
}

// Pub 0, Sub 0

func Test_p0s0(t *testing.T) {
	store := "/tmp/fvt/p0s0"
	topic := "/test/p0s0"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("p0s0-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("p0s0-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 0, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}
	p.Publish(topic, 0, false, "p0s0 payload 1")
	p.Publish(topic, 0, false, "p0s0 payload 2")

	wait(choke)
	wait(choke)

	p.Publish(topic, 0, false, "p0s0 payload 3")
	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

// Pub 0, Sub 1

func Test_p0s1(t *testing.T) {
	store := "/tmp/fvt/p0s1"
	topic := "/test/p0s1"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("p0s1-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("p0s1-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 1, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}
	p.Publish(topic, 0, false, "p0s1 payload 1")
	p.Publish(topic, 0, false, "p0s1 payload 2")

	wait(choke)
	wait(choke)

	p.Publish(topic, 0, false, "p0s1 payload 3")
	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

// Pub 0, Sub 2

func Test_p0s2(t *testing.T) {
	store := "/tmp/fvt/p0s2"
	topic := "/test/p0s2"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("p0s2-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("p0s2-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 2, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}
	p.Publish(topic, 0, false, "p0s2 payload 1")
	p.Publish(topic, 0, false, "p0s2 payload 2")

	wait(choke)
	wait(choke)

	p.Publish(topic, 0, false, "p0s2 payload 3")

	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

// Pub 1, Sub 0

func Test_p1s0(t *testing.T) {
	store := "/tmp/fvt/p1s0"
	topic := "/test/p1s0"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("p1s0-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("p1s0-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 0, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}
	p.Publish(topic, 1, false, "p1s0 payload 1")
	p.Publish(topic, 1, false, "p1s0 payload 2")

	wait(choke)
	wait(choke)

	p.Publish(topic, 1, false, "p1s0 payload 3")

	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

// Pub 1, Sub 1

func Test_p1s1(t *testing.T) {
	store := "/tmp/fvt/p1s1"
	topic := "/test/p1s1"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("p1s1-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("p1s1-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 1, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}
	p.Publish(topic, 1, false, "p1s1 payload 1")
	p.Publish(topic, 1, false, "p1s1 payload 2")

	wait(choke)
	wait(choke)

	p.Publish(topic, 1, false, "p1s1 payload 3")
	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

// Pub 1, Sub 2

func Test_p1s2(t *testing.T) {
	store := "/tmp/fvt/p1s2"
	topic := "/test/p1s2"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("p1s2-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("p1s2-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 2, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}
	p.Publish(topic, 1, false, "p1s2 payload 1")
	p.Publish(topic, 1, false, "p1s2 payload 2")

	wait(choke)
	wait(choke)

	p.Publish(topic, 1, false, "p1s2 payload 3")

	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

// Pub 2, Sub 0

func Test_p2s0(t *testing.T) {
	store := "/tmp/fvt/p2s0"
	topic := "/test/p2s0"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("p2s0-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("p2s0-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 0, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}
	p.Publish(topic, 2, false, "p2s0 payload 1")
	p.Publish(topic, 2, false, "p2s0 payload 2")
	wait(choke)
	wait(choke)

	p.Publish(topic, 2, false, "p2s0 payload 3")
	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

// Pub 2, Sub 1

func Test_p2s1(t *testing.T) {
	store := "/tmp/fvt/p2s1"
	topic := "/test/p2s1"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("p2s1-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("p2s1-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 1, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}
	p.Publish(topic, 2, false, "p2s1 payload 1")
	p.Publish(topic, 2, false, "p2s1 payload 2")

	wait(choke)
	wait(choke)

	p.Publish(topic, 2, false, "p2s1 payload 3")

	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

// Pub 2, Sub 2

func Test_p2s2(t *testing.T) {
	store := "/tmp/fvt/p2s2"
	topic := "/test/p2s2"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("p2s2-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("p2s2-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 2, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}
	p.Publish(topic, 2, false, "p2s2 payload 1")
	p.Publish(topic, 2, false, "p2s2 payload 2")

	wait(choke)
	wait(choke)

	p.Publish(topic, 2, false, "p2s2 payload 3")

	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

func Test_PublishMessage(t *testing.T) {
	store := "/tmp/fvt/PublishMessage"
	topic := "/test/pubmsg"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("pubmsg-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("pubmsg-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		if string(msg.Payload()) != "pubmsg payload" {
			fmt.Println("Message payload incorrect", msg.Payload(), len("pubmsg payload"))
			t.Fatalf("Message payload incorrect")
		}
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 2, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}

	text := "pubmsg payload"
	p.Publish(topic, 0, false, text)
	p.Publish(topic, 0, false, text)
	wait(choke)
	wait(choke)

	p.Publish(topic, 0, false, text)
	wait(choke)

	p.Disconnect(250)
	s.Disconnect(250)

	chkcond(isemptydir(store + "/p"))
	chkcond(isemptydir(store + "/s"))
}

func Test_PublishEmptyMessage(t *testing.T) {
	store := "/tmp/fvt/PublishEmptyMessage"
	topic := "/test/pubmsgempty"
	choke := make(chan bool)

	pops := NewClientOptions()
	pops.AddBroker(FVTTCP)
	pops.SetClientID("pubmsgempty-pub")
	pops.SetStore(NewFileStore(store + "/p"))
	p := NewClient(pops)

	sops := NewClientOptions()
	sops.AddBroker(FVTTCP)
	sops.SetClientID("pubmsgempty-sub")
	sops.SetStore(NewFileStore(store + "/s"))
	var f MessageHandler = func(client *Client, msg Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		fmt.Printf("MSG: %s\n", msg.Payload())
		if string(msg.Payload()) != "" {
			t.Fatalf("Message payload incorrect")
		}
		choke <- true
	}
	sops.SetDefaultPublishHandler(f)

	s := NewClient(sops)
	sToken := s.Connect()
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
	}

	sToken = s.Subscribe(topic, 2, nil)
	if sToken.Wait() && sToken.Error() != nil {
		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
	}

	pToken := p.Connect()
	if pToken.Wait() && pToken.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
	}

	p.Publish(topic, 0, false, "")
	p.Publish(topic, 0, false, "")
	wait(choke)
	wait(choke)

	p.Publish(topic, 0, false, "")
	wait(choke)

	p.Disconnect(250)
}

// func Test_Cleanstore(t *testing.T) {
// 	store := "/tmp/fvt/cleanstore"
// 	topic := "/test/cleanstore"

// 	pops := NewClientOptions()
// 	pops.AddBroker(FVTTCP)
// 	pops.SetClientID("cleanstore-pub")
// 	pops.SetStore(NewFileStore(store + "/p"))
// 	p := NewClient(pops)

// 	var s *Client
// 	sops := NewClientOptions()
// 	sops.AddBroker(FVTTCP)
// 	sops.SetClientID("cleanstore-sub")
// 	sops.SetCleanSession(false)
// 	sops.SetStore(NewFileStore(store + "/s"))
// 	var f MessageHandler = func(client *Client, msg Message) {
// 		fmt.Printf("TOPIC: %s\n", msg.Topic())
// 		fmt.Printf("MSG: %s\n", msg.Payload())
// 		// Close the connection after receiving
// 		// the first message so that hopefully
// 		// there is something in the store to be
// 		// cleaned.
// 		s.ForceDisconnect()
// 	}
// 	sops.SetDefaultPublishHandler(f)

// 	s = NewClient(sops)
// 	sToken := s.Connect()
// 	if sToken.Wait() && sToken.Error() != nil {
// 		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
// 	}

// 	sToken = s.Subscribe(topic, 2, nil)
// 	if sToken.Wait() && sToken.Error() != nil {
// 		t.Fatalf("Error on Client.Subscribe(): %v", sToken.Error())
// 	}

// 	pToken := p.Connect()
// 	if pToken.Wait() && pToken.Error() != nil {
// 		t.Fatalf("Error on Client.Connect(): %v", pToken.Error())
// 	}

// 	text := "test message"
// 	p.Publish(topic, 0, false, text)
// 	p.Publish(topic, 0, false, text)
// 	p.Publish(topic, 0, false, text)

// 	p.Disconnect(250)

// 	s2ops := NewClientOptions()
// 	s2ops.AddBroker(FVTTCP)
// 	s2ops.SetClientID("cleanstore-sub")
// 	s2ops.SetCleanSession(true)
// 	s2ops.SetStore(NewFileStore(store + "/s"))
// 	s2ops.SetDefaultPublishHandler(f)

// 	s2 := NewClient(s2ops)
// 	sToken = s2.Connect()
// 	if sToken.Wait() && sToken.Error() != nil {
// 		t.Fatalf("Error on Client.Connect(): %v", sToken.Error())
// 	}

// 	// at this point existing state should be cleared...
// 	// how to check?
// }

func Test_MultipleURLs(t *testing.T) {
	ops := NewClientOptions()
	ops.AddBroker("tcp://127.0.0.1:10000")
	ops.AddBroker(FVTTCP)
	ops.SetClientID("MutliURL")
	ops.SetStore(NewFileStore("/tmp/fvt/MultiURL"))

	c := NewClient(ops)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", token.Error())
	}

	token = c.Publish("/test/MultiURL", 0, false, "Publish qo0")
	token.Wait()

	c.Disconnect(250)
}

/*
// A test to make sure ping mechanism is working
// This test can be left commented out because it's annoying to wait for
func Test_ping3_idle10(t *testing.T) {
	ops := NewClientOptions()
	ops.AddBroker(FVTTCP)
	//ops.AddBroker("tcp://test.mosquitto.org:1883")
	ops.SetClientID("p3i10")
	ops.SetKeepAlive(4)

	c := NewClient(ops)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		t.Fatalf("Error on Client.Connect(): %v", token.Error())
	}
	time.Sleep(time.Duration(10) * time.Second)
	c.Disconnect(250)
}
*/
