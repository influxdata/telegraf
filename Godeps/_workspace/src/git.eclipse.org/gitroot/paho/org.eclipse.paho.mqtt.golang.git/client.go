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

// Package mqtt provides an MQTT v3.1.1 client library.
package mqtt

import (
	"errors"
	"fmt"
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git/packets"
	"net"
	"sync"
	"time"
)

// ClientInt is the interface definition for a Client as used by this
// library, the interface is primarily to allow mocking tests.
type ClientInt interface {
	IsConnected() bool
	Connect() Token
	Disconnect(uint)
	disconnect()
	Publish(string, byte, bool, interface{}) Token
	Subscribe(string, byte, MessageHandler) Token
	SubscribeMultiple(map[string]byte, MessageHandler) Token
	Unsubscribe(...string) Token
}

// Client is an MQTT v3.1.1 client for communicating
// with an MQTT server using non-blocking methods that allow work
// to be done in the background.
// An application may connect to an MQTT server using:
//   A plain TCP socket
//   A secure SSL/TLS socket
//   A websocket
// To enable ensured message delivery at Quality of Service (QoS) levels
// described in the MQTT spec, a message persistence mechanism must be
// used. This is done by providing a type which implements the Store
// interface. For convenience, FileStore and MemoryStore are provided
// implementations that should be sufficient for most use cases. More
// information can be found in their respective documentation.
// Numerous connection options may be specified by configuring a
// and then supplying a ClientOptions type.
type Client struct {
	sync.RWMutex
	messageIds
	conn            net.Conn
	ibound          chan packets.ControlPacket
	obound          chan *PacketAndToken
	oboundP         chan *PacketAndToken
	msgRouter       *router
	stopRouter      chan bool
	incomingPubChan chan *packets.PublishPacket
	errors          chan error
	stop            chan struct{}
	persist         Store
	options         ClientOptions
	lastContact     lastcontact
	pingOutstanding bool
	connected       bool
	workers         sync.WaitGroup
}

// NewClient will create an MQTT v3.1.1 client with all of the options specified
// in the provided ClientOptions. The client must have the Start method called
// on it before it may be used. This is to make sure resources (such as a net
// connection) are created before the application is actually ready.
func NewClient(o *ClientOptions) *Client {
	c := &Client{}
	c.options = *o

	if c.options.Store == nil {
		c.options.Store = NewMemoryStore()
	}
	switch c.options.ProtocolVersion {
	case 3, 4:
		c.options.protocolVersionExplicit = true
	default:
		c.options.ProtocolVersion = 4
		c.options.protocolVersionExplicit = false
	}
	c.persist = c.options.Store
	c.connected = false
	c.messageIds = messageIds{index: make(map[uint16]Token)}
	c.msgRouter, c.stopRouter = newRouter()
	c.msgRouter.setDefaultHandler(c.options.DefaultPublishHander)
	return c
}

// IsConnected returns a bool signifying whether
// the client is connected or not.
func (c *Client) IsConnected() bool {
	c.RLock()
	defer c.RUnlock()
	return c.connected
}

func (c *Client) setConnected(status bool) {
	c.Lock()
	defer c.Unlock()
	c.connected = status
}

//ErrNotConnected is the error returned from function calls that are
//made when the client is not connected to a broker
var ErrNotConnected = errors.New("Not Connected")

// Connect will create a connection to the message broker
// If clean session is false, then a slice will
// be returned containing Receipts for all messages
// that were in-flight at the last disconnect.
// If clean session is true, then any existing client
// state will be removed.
func (c *Client) Connect() Token {
	var err error
	t := newToken(packets.Connect).(*ConnectToken)
	DEBUG.Println(CLI, "Connect()")

	go func() {
		var rc byte
		cm := newConnectMsgFromOptions(&c.options)

		for _, broker := range c.options.Servers {
		CONN:
			DEBUG.Println(CLI, "about to write new connect msg")
			c.conn, err = openConnection(broker, &c.options.TLSConfig, c.options.ConnectTimeout)
			if err == nil {
				DEBUG.Println(CLI, "socket connected to broker")
				switch c.options.ProtocolVersion {
				case 3:
					DEBUG.Println(CLI, "Using MQTT 3.1 protocol")
					cm.ProtocolName = "MQIsdp"
					cm.ProtocolVersion = 3
				default:
					DEBUG.Println(CLI, "Using MQTT 3.1.1 protocol")
					c.options.ProtocolVersion = 4
					cm.ProtocolName = "MQTT"
					cm.ProtocolVersion = 4
				}
				cm.Write(c.conn)

				rc = c.connect()
				if rc != packets.Accepted {
					c.conn.Close()
					c.conn = nil
					//if the protocol version was explicitly set don't do any fallback
					if c.options.protocolVersionExplicit {
						ERROR.Println(CLI, "Connecting to", broker, "CONNACK was not CONN_ACCEPTED, but rather", packets.ConnackReturnCodes[rc])
						continue
					}
					if c.options.ProtocolVersion == 4 {
						DEBUG.Println(CLI, "Trying reconnect using MQTT 3.1 protocol")
						c.options.ProtocolVersion = 3
						goto CONN
					}
				}
				break
			} else {
				ERROR.Println(CLI, err.Error())
				WARN.Println(CLI, "failed to connect to broker, trying next")
				rc = packets.ErrNetworkError
			}
		}

		if c.conn == nil {
			ERROR.Println(CLI, "Failed to connect to a broker")
			t.returnCode = rc
			if rc != packets.ErrNetworkError {
				t.err = packets.ConnErrors[rc]
			} else {
				t.err = fmt.Errorf("%s : %s", packets.ConnErrors[rc], err)
			}
			t.flowComplete()
			return
		}

		c.lastContact.update()
		c.persist.Open()

		c.obound = make(chan *PacketAndToken, 100)
		c.oboundP = make(chan *PacketAndToken, 100)
		c.ibound = make(chan packets.ControlPacket)
		c.errors = make(chan error)
		c.stop = make(chan struct{})

		c.incomingPubChan = make(chan *packets.PublishPacket, 100)
		c.msgRouter.matchAndDispatch(c.incomingPubChan, c.options.Order, c)

		c.workers.Add(1)
		go outgoing(c)
		go alllogic(c)

		c.connected = true
		DEBUG.Println(CLI, "client is connected")
		if c.options.OnConnect != nil {
			go c.options.OnConnect(c)
		}

		if c.options.KeepAlive != 0 {
			c.workers.Add(1)
			go keepalive(c)
		}

		// Take care of any messages in the store
		//var leftovers []Receipt
		if c.options.CleanSession == false {
			//leftovers = c.resume()
		} else {
			c.persist.Reset()
		}

		// Do not start incoming until resume has completed
		c.workers.Add(1)
		go incoming(c)

		DEBUG.Println(CLI, "exit startClient")
		t.flowComplete()
	}()
	return t
}

// internal function used to reconnect the client when it loses its connection
func (c *Client) reconnect() {
	DEBUG.Println(CLI, "enter reconnect")
	var rc byte = 1
	var sleep uint = 1
	var err error

	for rc != 0 {
		cm := newConnectMsgFromOptions(&c.options)

		for _, broker := range c.options.Servers {
		CONN:
			DEBUG.Println(CLI, "about to write new connect msg")
			c.conn, err = openConnection(broker, &c.options.TLSConfig, c.options.ConnectTimeout)
			if err == nil {
				DEBUG.Println(CLI, "socket connected to broker")
				switch c.options.ProtocolVersion {
				case 3:
					DEBUG.Println(CLI, "Using MQTT 3.1 protocol")
					cm.ProtocolName = "MQIsdp"
					cm.ProtocolVersion = 3
				default:
					DEBUG.Println(CLI, "Using MQTT 3.1.1 protocol")
					c.options.ProtocolVersion = 4
					cm.ProtocolName = "MQTT"
					cm.ProtocolVersion = 4
				}
				cm.Write(c.conn)

				rc = c.connect()
				if rc != packets.Accepted {
					c.conn.Close()
					c.conn = nil
					//if the protocol version was explicitly set don't do any fallback
					if c.options.protocolVersionExplicit {
						ERROR.Println(CLI, "Connecting to", broker, "CONNACK was not Accepted, but rather", packets.ConnackReturnCodes[rc])
						continue
					}
					if c.options.ProtocolVersion == 4 {
						DEBUG.Println(CLI, "Trying reconnect using MQTT 3.1 protocol")
						c.options.ProtocolVersion = 3
						goto CONN
					}
				}
				break
			} else {
				ERROR.Println(CLI, err.Error())
				WARN.Println(CLI, "failed to connect to broker, trying next")
				rc = packets.ErrNetworkError
			}
		}
		if rc != 0 {
			DEBUG.Println(CLI, "Reconnect failed, sleeping for", sleep, "seconds")
			time.Sleep(time.Duration(sleep) * time.Second)
			if sleep <= uint(c.options.MaxReconnectInterval.Seconds()) {
				sleep *= 2
			}
		}
	}

	c.lastContact.update()
	c.stop = make(chan struct{})

	c.workers.Add(1)
	go outgoing(c)
	go alllogic(c)

	c.setConnected(true)
	DEBUG.Println(CLI, "client is reconnected")
	if c.options.OnConnect != nil {
		go c.options.OnConnect(c)
	}

	if c.options.KeepAlive != 0 {
		c.workers.Add(1)
		go keepalive(c)
	}
	c.workers.Add(1)
	go incoming(c)
}

// This function is only used for receiving a connack
// when the connection is first started.
// This prevents receiving incoming data while resume
// is in progress if clean session is false.
func (c *Client) connect() byte {
	DEBUG.Println(NET, "connect started")

	ca, err := packets.ReadPacket(c.conn)
	if err != nil {
		ERROR.Println(NET, "connect got error", err)
		//c.errors <- err
		return packets.ErrNetworkError
	}
	msg := ca.(*packets.ConnackPacket)

	if msg == nil || msg.FixedHeader.MessageType != packets.Connack {
		ERROR.Println(NET, "received msg that was nil or not CONNACK")
	} else {
		DEBUG.Println(NET, "received connack")
	}
	return msg.ReturnCode
}

// Disconnect will end the connection with the server, but not before waiting
// the specified number of milliseconds to wait for existing work to be
// completed.
func (c *Client) Disconnect(quiesce uint) {
	if !c.IsConnected() {
		WARN.Println(CLI, "already disconnected")
		return
	}
	DEBUG.Println(CLI, "disconnecting")
	c.setConnected(false)

	dm := packets.NewControlPacket(packets.Disconnect).(*packets.DisconnectPacket)
	dt := newToken(packets.Disconnect)
	c.oboundP <- &PacketAndToken{p: dm, t: dt}

	// wait for work to finish, or quiesce time consumed
	dt.WaitTimeout(time.Duration(quiesce) * time.Millisecond)
	c.disconnect()
}

// ForceDisconnect will end the connection with the mqtt broker immediately.
func (c *Client) forceDisconnect() {
	if !c.IsConnected() {
		WARN.Println(CLI, "already disconnected")
		return
	}
	c.setConnected(false)
	c.conn.Close()
	DEBUG.Println(CLI, "forcefully disconnecting")
	c.disconnect()
}

func (c *Client) internalConnLost(err error) {
	close(c.stop)
	c.conn.Close()
	c.workers.Wait()
	if c.IsConnected() {
		if c.options.OnConnectionLost != nil {
			go c.options.OnConnectionLost(c, err)
		}
		if c.options.AutoReconnect {
			go c.reconnect()
		} else {
			c.setConnected(false)
		}
	}
}

func (c *Client) disconnect() {
	select {
	case <-c.stop:
		//someone else has already closed the channel, must be error
	default:
		close(c.stop)
	}
	c.conn.Close()
	c.workers.Wait()
	close(c.stopRouter)
	DEBUG.Println(CLI, "disconnected")
	c.persist.Close()
}

// Publish will publish a message with the specified QoS
// and content to the specified topic.
// Returns a read only channel used to track
// the delivery of the message.
func (c *Client) Publish(topic string, qos byte, retained bool, payload interface{}) Token {
	token := newToken(packets.Publish).(*PublishToken)
	DEBUG.Println(CLI, "enter Publish")
	if !c.IsConnected() {
		token.err = ErrNotConnected
		token.flowComplete()
		return token
	}
	pub := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	pub.Qos = qos
	pub.TopicName = topic
	pub.Retain = retained
	switch payload.(type) {
	case string:
		pub.Payload = []byte(payload.(string))
	case []byte:
		pub.Payload = payload.([]byte)
	default:
		token.err = errors.New("Unknown payload type")
		token.flowComplete()
		return token
	}

	DEBUG.Println(CLI, "sending publish message, topic:", topic)
	c.obound <- &PacketAndToken{p: pub, t: token}
	return token
}

// Subscribe starts a new subscription. Provide a MessageHandler to be executed when
// a message is published on the topic provided.
func (c *Client) Subscribe(topic string, qos byte, callback MessageHandler) Token {
	token := newToken(packets.Subscribe).(*SubscribeToken)
	DEBUG.Println(CLI, "enter Subscribe")
	if !c.IsConnected() {
		token.err = ErrNotConnected
		token.flowComplete()
		return token
	}
	sub := packets.NewControlPacket(packets.Subscribe).(*packets.SubscribePacket)
	if err := validateTopicAndQos(topic, qos); err != nil {
		token.err = err
		return token
	}
	sub.Topics = append(sub.Topics, topic)
	sub.Qoss = append(sub.Qoss, qos)
	DEBUG.Println(sub.String())

	if callback != nil {
		c.msgRouter.addRoute(topic, callback)
	}

	token.subs = append(token.subs, topic)
	c.oboundP <- &PacketAndToken{p: sub, t: token}
	DEBUG.Println(CLI, "exit Subscribe")
	return token
}

// SubscribeMultiple starts a new subscription for multiple topics. Provide a MessageHandler to
// be executed when a message is published on one of the topics provided.
func (c *Client) SubscribeMultiple(filters map[string]byte, callback MessageHandler) Token {
	var err error
	token := newToken(packets.Subscribe).(*SubscribeToken)
	DEBUG.Println(CLI, "enter SubscribeMultiple")
	if !c.IsConnected() {
		token.err = ErrNotConnected
		token.flowComplete()
		return token
	}
	sub := packets.NewControlPacket(packets.Subscribe).(*packets.SubscribePacket)
	if sub.Topics, sub.Qoss, err = validateSubscribeMap(filters); err != nil {
		token.err = err
		return token
	}

	if callback != nil {
		for topic := range filters {
			c.msgRouter.addRoute(topic, callback)
		}
	}
	token.subs = make([]string, len(sub.Topics))
	copy(token.subs, sub.Topics)
	c.oboundP <- &PacketAndToken{p: sub, t: token}
	DEBUG.Println(CLI, "exit SubscribeMultiple")
	return token
}

// Unsubscribe will end the subscription from each of the topics provided.
// Messages published to those topics from other clients will no longer be
// received.
func (c *Client) Unsubscribe(topics ...string) Token {
	token := newToken(packets.Unsubscribe).(*UnsubscribeToken)
	DEBUG.Println(CLI, "enter Unsubscribe")
	if !c.IsConnected() {
		token.err = ErrNotConnected
		token.flowComplete()
		return token
	}
	unsub := packets.NewControlPacket(packets.Unsubscribe).(*packets.UnsubscribePacket)
	unsub.Topics = make([]string, len(topics))
	copy(unsub.Topics, topics)

	c.oboundP <- &PacketAndToken{p: unsub, t: token}
	for _, topic := range topics {
		c.msgRouter.deleteRoute(topic)
	}

	DEBUG.Println(CLI, "exit Unsubscribe")
	return token
}

//DefaultConnectionLostHandler is a definition of a function that simply
//reports to the DEBUG log the reason for the client losing a connection.
func DefaultConnectionLostHandler(client *Client, reason error) {
	DEBUG.Println("Connection lost:", reason.Error())
}
