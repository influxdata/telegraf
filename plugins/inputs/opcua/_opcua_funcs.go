package opcua

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/ua"
)

// READ VALUE

func readValue(address, node string) {
	var (
		endpoint = flag.String("endpoint", address, "OPC UA Endpoint URL")
		nodeID   = flag.String("node", node, "NodeID to read")
		policy   = flag.String("sec-policy", "Basic256Sha256", "Security Policy URL or one of None, Basic128Rsa15, Basic256, Basic256Sha256")
		mode     = flag.String("sec-mode", "SignAndEncrypt", "Security Mode: one of None, Sign, SignAndEncrypt")
		certFile = flag.String("cert", "./trusted/cert.pem", "Path to ./trusted/cert.pem. Required for security mode/policy != None")
		keyFile  = flag.String("key", "./trusted/key.pem", "Path to private ./trusted/key.pem. Required for security mode/policy != None")
	)
	flag.BoolVar(&debug.Enable, "debug", false, "enable debug logging")
	flag.Parse()
	log.SetFlags(0)

	ctx := context.Background()

	endpoints, err := opcua.GetEndpoints(*endpoint)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("\nCalled function: opcua.GetEndpoints()\n")

	ep := opcua.SelectEndpoint(endpoints, *policy, ua.MessageSecurityModeFromString(*mode))
	if ep == nil {
		log.Fatal("Failed to find suitable endpoint")
	}

	log.Print("\nCalled function: opcua.SelectEndpoint()\n")

	log.Printf("\nENDPOINTS: \n%s", endpoints)

	opts := []opcua.Option{
		opcua.SecurityPolicy(*policy),
		opcua.SecurityModeString(*mode),
		opcua.CertificateFile(*certFile),
		opcua.PrivateKeyFile(*keyFile),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c := opcua.NewClient(*endpoint, opts...)
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	log.Print("\nCalled function: opcua.NewClient()\n")

	id, err := ua.ParseNodeID(*nodeID)
	if err != nil {
		log.Fatalf("invalid node id: %v", err)
	}

	log.Print("\nCalled function: ua.ParseNodeID()\n")

	req := &ua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*ua.ReadValueID{
			&ua.ReadValueID{NodeID: id},
		},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	log.Print("\nCalled function: ua.ReadRequest()\n")

	resp, err := c.Read(req)
	if err != nil {
		log.Fatalf("Read failed: %s", err)
	}
	if resp.Results[0].Status != ua.StatusOK {
		log.Fatalf("Status not OK: %v", resp.Results[0].Status)
	}
	log.Printf("%#v", resp.Results[0].Value.Value())
}

// WRITE

func write(address, node string, newValue interface{}) {
	var (
		endpoint = flag.String("endpoint", address, "OPC UA Endpoint URL")
		nodeID   = flag.String("node", node, "NodeID to read")
		//policy   = flag.String("policy", "Basic256Sha256", "Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto")
		//mode     = flag.String("mode", "SignAndEncrypt", "Security mode: None, Sign, SignAndEncrypt. Default: auto")
		//certFile = flag.String("cert", "./trusted/cert.pem", "Path to cert.pem. Required for security mode/policy != None")
		//keyFile  = flag.String("key", "./trusted/key.pem", "Path to private key.pem. Required for security mode/policy != None")
		policy = flag.String("policy", "None", "Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto")
		mode   = flag.String("mode", "None", "Security mode: None, Sign, SignAndEncrypt. Default: auto")
	)
	flag.BoolVar(&debug.Enable, "debug", false, "enable debug logging")
	flag.Parse()
	log.SetFlags(0)

	ctx := context.Background()

	endpoints, err := opcua.GetEndpoints(*endpoint)
	if err != nil {
		log.Fatal(err)
	}
	ep := opcua.SelectEndpoint(endpoints, *policy, ua.MessageSecurityModeFromString(*mode))
	if ep == nil {
		log.Fatal("Failed to find suitable endpoint")
	}

	opts := []opcua.Option{
		opcua.SecurityPolicy(*policy),
		opcua.SecurityModeString(*mode),
		//opcua.CertificateFile(*certFile),
		//opcua.PrivateKeyFile(*keyFile),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c := opcua.NewClient(*endpoint, opts...)
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	id, err := ua.ParseNodeID(*nodeID)
	if err != nil {
		log.Fatalf("invalid node id: %v", err)
	}

	v, err := ua.NewVariant(newValue)
	if err != nil {
		log.Fatalf("invalid value: %v", err)
	}

	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			&ua.WriteValue{
				NodeID:      id,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        v,
				},
			},
		},
	}

	resp, err := c.Write(req)
	if err != nil {
		log.Fatalf("Read failed: %s", err)
	}
	log.Printf("%v", resp.Results[0])
}

// SUBSCRIBE

func subscribe(address, node string) {
	var (
		endpoint = flag.String("endpoint", address, "OPC UA Endpoint URL")
		policy   = flag.String("policy", "Basic256Sha256", "Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto")
		mode     = flag.String("mode", "SignAndEncrypt", "Security mode: None, Sign, SignAndEncrypt. Default: auto")
		certFile = flag.String("cert", "./trusted/cert.pem", "Path to cert.pem. Required for security mode/policy != None")
		keyFile  = flag.String("key", "./trusted/key.pem", "Path to private key.pem. Required for security mode/policy != None")
		nodeID   = flag.String("node", node, "node id to subscribe to")
		interval = flag.String("interval", opcua.DefaultSubscriptionInterval.String(), "subscription interval")
	)
	flag.BoolVar(&debug.Enable, "debug", false, "enable debug logging")
	flag.Parse()
	log.SetFlags(0)

	subInterval, err := time.ParseDuration(*interval)
	if err != nil {
		log.Fatal(err)
	}

	// add an arbitrary timeout to demonstrate how to stop a subscription
	// with a context.
	//d := 30 * time.Second
	//ctx, cancel := context.WithTimeout(context.Background(), d)
	ctx := context.Background()
	//defer cancel()

	log.Print("\nCalled function: opcua.GetEndpoints()\n")

	endpoints, err := opcua.GetEndpoints(*endpoint)
	if err != nil {
		log.Fatal(err)
	}
	ep := opcua.SelectEndpoint(endpoints, *policy, ua.MessageSecurityModeFromString(*mode))
	if ep == nil {
		log.Fatal("Failed to find suitable endpoint")
	}

	log.Print("\nCalled function: opcua.SelectEndpoint()\n")

	log.Printf("\nENDPOINTS: \n%s", endpoints)

	fmt.Println("*", ep.SecurityPolicyURI, ep.SecurityMode)

	opts := []opcua.Option{
		opcua.SecurityPolicy(*policy),
		opcua.SecurityModeString(*mode),
		opcua.CertificateFile(*certFile),
		opcua.PrivateKeyFile(*keyFile),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c := opcua.NewClient(*endpoint, opts...)
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	log.Print("\nCalled function: opcua.NewClient()\n")

	notifyCh := make(chan *opcua.PublishNotificationData)

	sub, err := c.Subscribe(&opcua.SubscriptionParameters{
		Interval: subInterval, //500 * time.Millisecond,
	}, notifyCh)

	log.Print("\nCalled function: c.Subscribe()\n")

	if err != nil {
		log.Fatal(err)
	}
	defer sub.Cancel()
	log.Printf("Created subscription with id %v", sub.SubscriptionID)

	id, err := ua.ParseNodeID(*nodeID)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("\nCalled function: ua.ParseNodeID()\n")

	// arbitrary client handle for the monitoring item
	handle := uint32(42)
	miCreateRequest := opcua.NewMonitoredItemCreateRequestWithDefaults(id, ua.AttributeIDValue, handle)
	res, err := sub.Monitor(ua.TimestampsToReturnBoth, miCreateRequest)
	if err != nil || res.Results[0].StatusCode != ua.StatusOK {
		log.Fatal(err)
	}

	log.Print("\nCalled function: opcua.NewMonitoredItemCreateRequestWithDefaults()\n")

	go sub.Run(ctx) // start Publish loop

	// read from subscription's notification channel until ctx is cancelled
	for {
		select {
		case <-ctx.Done():
			log.Print("Received ctx.Done()")
			return
		case res := <-sub.Notifs:
			if res.Error != nil {
				log.Print(res.Error)
				continue
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				for _, item := range x.MonitoredItems {
					data := item.Value.Value.Value()
					log.Printf("MonitoredItem with client handle %v = %v", item.ClientHandle, data)
				}

			default:
				log.Printf("what's this publish result? %T", res.Value)
			}
		}
	}
}

func subscribeMany(address string, nodes map[string]string) {
	var (
		endpoint = flag.String("endpoint", address, "OPC UA Endpoint URL")
		policy   = flag.String("policy", "Basic256Sha256", "Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto")
		mode     = flag.String("mode", "SignAndEncrypt", "Security mode: None, Sign, SignAndEncrypt. Default: auto")
		certFile = flag.String("cert", "./trusted/cert.pem", "Path to cert.pem. Required for security mode/policy != None")
		keyFile  = flag.String("key", "./trusted/key.pem", "Path to private key.pem. Required for security mode/policy != None")
		interval = flag.String("interval", opcua.DefaultSubscriptionInterval.String(), "subscription interval")
		//nodeID   = flag.String("node", "", "node id to subscribe to")
	)
	flag.BoolVar(&debug.Enable, "debug", false, "enable debug logging")
	flag.Parse()
	log.SetFlags(0)

	subInterval, err := time.ParseDuration(*interval)
	if err != nil {
		log.Fatal(err)
	}

	// add an arbitrary timeout to demonstrate how to stop a subscription
	// with a context.
	//d := 30 * time.Second
	//ctx, cancel := context.WithTimeout(context.Background(), d)
	ctx := context.Background()
	//defer cancel()

	endpoints, err := opcua.GetEndpoints(*endpoint)
	if err != nil {
		log.Fatal(err)
	}
	ep := opcua.SelectEndpoint(endpoints, *policy, ua.MessageSecurityModeFromString(*mode))
	if ep == nil {
		log.Fatal("Failed to find suitable endpoint")
	}

	fmt.Println("*", ep.SecurityPolicyURI, ep.SecurityMode)

	opts := []opcua.Option{
		opcua.SecurityPolicy(*policy),
		opcua.SecurityModeString(*mode),
		opcua.CertificateFile(*certFile),
		opcua.PrivateKeyFile(*keyFile),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c := opcua.NewClient(*endpoint, opts...)
	if err := c.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	log.Print("\nCalled function: opcua.NewClient()\n")

	notifyCh := make(chan *opcua.PublishNotificationData)

	sub, err := c.Subscribe(&opcua.SubscriptionParameters{
		Interval: subInterval, //500 * time.Millisecond,
	}, notifyCh)

	log.Print("\nCalled function: c.Subscribe()\n")

	if err != nil {
		log.Fatal(err)
	}
	defer sub.Cancel()
	log.Printf("Created subscription with id %v", sub.SubscriptionID)

	allRequests := []*ua.MonitoredItemCreateRequest{}
	handleCounter := uint32(10)
	handleLookup := map[uint32]string{}

	for handleName, nodeID := range nodes {

		id, err := ua.ParseNodeID(nodeID)
		if err != nil {
			log.Fatal(err)
		}

		log.Print("\nCalled function: ua.ParseNodeID()\n")

		// arbitrary client handle for the monitoring item
		handle := handleCounter
		handleLookup[handle] = handleName
		miCreateRequest := opcua.NewMonitoredItemCreateRequestWithDefaults(id, ua.AttributeIDValue, handle)
		allRequests = append(allRequests, miCreateRequest)

		log.Print("\nCalled function: opcua.NewMonitoredItemCreateRequestWithDefaults()\n")
		handleCounter++

	}

	res, err := sub.Monitor(ua.TimestampsToReturnBoth, allRequests...)
	if err != nil || res.Results[0].StatusCode != ua.StatusOK {
		log.Fatal(err)
	}

	go sub.Run(ctx) // start Publish loop

	// read from subscription's notification channel until ctx is cancelled
	for {
		select {
		case <-ctx.Done():
			log.Print("Received ctx.Done()")
			return
		case res := <-sub.Notifs:
			if res.Error != nil {
				log.Print(res.Error)
				continue
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				for _, item := range x.MonitoredItems {
					data := item.Value.Value.Value()
					log.Printf("MonitoredItem %s (%v) = %v", handleLookup[item.ClientHandle], item.ClientHandle, data)
				}

			default:
				log.Printf("what's this publish result? %T", res.Value)
			}
		}
	}
}
