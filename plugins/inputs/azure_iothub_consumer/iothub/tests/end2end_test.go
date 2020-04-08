package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/amenzhinsky/iothub/iotdevice"
	"github.com/amenzhinsky/iothub/iotdevice/transport"
	"github.com/amenzhinsky/iothub/iotdevice/transport/mqtt"
	"github.com/amenzhinsky/iothub/iotservice"
)

func TestEnd2End(t *testing.T) {
	cs := os.Getenv("TEST_IOTHUB_SERVICE_CONNECTION_STRING")
	if cs == "" {
		t.Fatal("$TEST_IOTHUB_SERVICE_CONNECTION_STRING is empty")
	}
	sc, err := iotservice.NewFromConnectionString(cs) //iotservice.WithLogger(logger.New(logger.LevelDebug, nil)),

	if err != nil {
		t.Fatal(err)
	}
	defer sc.Close()

	// create devices with all possible authentication types
	_, err = sc.DeleteDevices(context.Background(), []*iotservice.Device{
		{DeviceID: "golang-iothub-sas"},
		{DeviceID: "golang-iothub-self-signed"},
		{DeviceID: "golang-iothub-ca"},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	result, err := sc.CreateDevices(context.Background(), []*iotservice.Device{{
		DeviceID: "golang-iothub-sas",
		Authentication: &iotservice.Authentication{
			Type: iotservice.AuthSAS,
		},
	}, {
		DeviceID: "golang-iothub-self-signed",
		Authentication: &iotservice.Authentication{
			Type: iotservice.AuthSelfSigned,
			X509Thumbprint: &iotservice.X509Thumbprint{
				PrimaryThumbprint:   "443ABB6DEA8F93D5987D31D2607BE2931217752C",
				SecondaryThumbprint: "443ABB6DEA8F93D5987D31D2607BE2931217752C",
			},
		},
	}, {
		DeviceID: "golang-iothub-ca",
		Authentication: &iotservice.Authentication{
			Type: iotservice.AuthCA,
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsSuccessful {
		t.Fatalf("couldn't create devices: %v", result.Errors)
	}

	for name, mktransport := range map[string]func() transport.Transport{
		"mqtt":    func() transport.Transport { return mqtt.New() },
		"mqtt-ws": func() transport.Transport { return mqtt.New(mqtt.WithWebSocket(true)) },
		// TODO: "amqp": func() transport.Transport { return amqp.New() },
		// TODO: "http": func() transport.Transport { return http.New() },
	} {
		mktransport := mktransport
		t.Run(name, func(t *testing.T) {
			for auth, suite := range map[string]struct {
				init func(transport transport.Transport) (*iotdevice.Client, error)
				only string
			}{
				// TODO: ca authentication
				"x509": {
					func(transport transport.Transport) (*iotdevice.Client, error) {
						return iotdevice.NewFromX509FromFile(
							transport,
							"golang-iothub-self-signed",
							sc.HostName(),
							"testdata/device.crt",
							"testdata/device.key",
						)
					},
					"DeviceToCloud", // just need to check access
				},
				"sas": {
					func(transport transport.Transport) (*iotdevice.Client, error) {
						device, err := sc.GetDevice(context.Background(), "golang-iothub-sas")
						if err != nil {
							return nil, err
						}
						dcs, err := sc.DeviceConnectionString(device, false)
						if err != nil {
							t.Fatal(err)
						}
						return iotdevice.NewFromConnectionString(transport, dcs)
					},
					"*",
				},
			} {
				for name, test := range map[string]func(*testing.T, *iotservice.Client, *iotdevice.Client){
					"DeviceToCloud":    testDeviceToCloud,
					"CloudToDevice":    testCloudToDevice,
					"DirectMethod":     testDirectMethod,
					"UpdateDeviceTwin": testUpdateTwin,
					"SubscribeTwin":    testSubscribeTwin,
				} {
					if suite.only != "*" && suite.only != name {
						continue
					}
					test, suite, mktransport := test, suite, mktransport
					t.Run(auth+"/"+name, func(t *testing.T) {
						dc, err := suite.init(mktransport())
						if err != nil {
							t.Fatal(err)
						}
						defer dc.Close()
						if err := dc.Connect(context.Background()); err != nil {
							t.Fatal(err)
						}
						test(t, sc, dc)
					})
				}
			}
		})
	}
}

func testDeviceToCloud(t *testing.T, sc *iotservice.Client, dc *iotdevice.Client) {
	evsc := make(chan *iotservice.Event, 1)
	errc := make(chan error, 2)
	go func() {
		errc <- sc.SubscribeEvents(context.Background(), func(ev *iotservice.Event) error {
			if ev.ConnectionDeviceID == dc.DeviceID() {
				evsc <- ev
			}
			return nil
		})
	}()

	payload := []byte("hello")
	props := map[string]string{"a": "a", "b": "b"}
	done := make(chan struct{})
	defer close(done)

	// send events until one of them is received
	go func() {
		for {
			if err := dc.SendEvent(context.Background(), payload,
				iotdevice.WithSendMessageID(genID()),
				iotdevice.WithSendCorrelationID(genID()),
				iotdevice.WithSendProperties(props),
			); err != nil {
				errc <- err
				break
			}
			select {
			case <-done:
			case <-time.After(500 * time.Millisecond):
			}
		}
	}()

	select {
	case msg := <-evsc:
		if msg.MessageID == "" {
			t.Error("MessageID is empty")
		}
		if msg.CorrelationID == "" {
			t.Error("CorrelationID is empty")
		}
		if msg.ConnectionDeviceID != dc.DeviceID() {
			t.Errorf("ConnectionDeviceID = %q, want %q", msg.ConnectionDeviceID, dc.DeviceID())
		}
		if msg.ConnectionDeviceGenerationID == "" {
			t.Error("ConnectionDeviceGenerationID is empty")
		}
		if msg.ConnectionAuthMethod == nil {
			t.Error("ConnectionAuthMethod is nil")
		}
		if msg.MessageSource == "" {
			t.Error("MessageSource is empty")
		}
		if msg.EnqueuedTime.IsZero() {
			t.Error("EnqueuedTime is zero")
		}
		if !bytes.Equal(msg.Payload, payload) {
			t.Errorf("Payload = %v, want %v", msg.Payload, payload)
		}
		testProperties(t, msg.Properties, props)
	case err := <-errc:
		t.Fatal(err)
	case <-time.After(10 * time.Second):
		t.Fatal("d2c timed out")
	}
}

func testProperties(t *testing.T, got, want map[string]string) {
	t.Helper()
	for k, v := range want {
		x, ok := got[k]
		if !ok || x != v {
			t.Errorf("Properties = %v, want %v", got, want)
			return
		}
	}
}

func testCloudToDevice(t *testing.T, sc *iotservice.Client, dc *iotdevice.Client) {
	fbsc := make(chan *iotservice.Feedback, 1)
	errc := make(chan error, 3)

	sub, err := dc.SubscribeEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// subscribe to feedback and report first registered message id
	go func() {
		errc <- sc.SubscribeFeedback(context.Background(), func(fb *iotservice.Feedback) error {
			fbsc <- fb
			return nil
		})
	}()

	payload := []byte("hello")
	props := map[string]string{"a": "a", "b": "b"}
	uid := "golang-iothub"

	mid := genID()
	if err := sc.SendEvent(context.Background(), dc.DeviceID(), payload,
		iotservice.WithSendAck(iotservice.AckFull),
		iotservice.WithSendProperties(props),
		iotservice.WithSendUserID(uid),
		iotservice.WithSendMessageID(mid),
		iotservice.WithSendCorrelationID(genID()),
		iotservice.WithSendExpiryTime(time.Now().Add(5*time.Second)),
	); err != nil {
		errc <- err
		return
	}

	for {
		select {
		case msg := <-sub.C():
			if msg.MessageID != mid {
				continue
			}

			// validate event feedback
		Outer:
			for {
				select {
				case fb := <-fbsc:
					if fb.OriginalMessageID != mid {
						continue
					}
					if fb.StatusCode != "Success" {
						t.Errorf("feedback status = %q, want %q", fb.StatusCode, "Success")
					}
					break Outer
				case <-time.After(30 * time.Second):
					t.Fatal("feedback timed out")
				}
			}

			// validate message content
			if msg.To == "" {
				t.Error("To is empty")
			}
			if msg.UserID != uid {
				t.Errorf("UserID = %q, want %q", msg.UserID, uid)
			}
			if !bytes.Equal(msg.Payload, payload) {
				t.Errorf("Payload = %v, want %v", msg.Payload, payload)
			}
			if msg.MessageID == "" {
				t.Error("MessageID is empty")
			}
			if msg.CorrelationID == "" {
				t.Error("CorrelationID is empty")
			}
			if msg.ExpiryTime.IsZero() {
				t.Error("ExpiryTime is zero")
			}
			testProperties(t, msg.Properties, props)
			return
		case err := <-errc:
			t.Fatal(err)
		case <-time.After(10 * time.Second):
			t.Fatal("c2d timed out")
		}
	}
}

func testUpdateTwin(t *testing.T, sc *iotservice.Client, dc *iotdevice.Client) {
	// update state and keep track of version
	s := fmt.Sprintf("%d", time.Now().UnixNano())
	v, err := dc.UpdateTwinState(context.Background(), map[string]interface{}{
		"ts": s,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, r, err := dc.RetrieveTwinState(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v != r.Version() {
		t.Errorf("update-twin version = %d, want %d", r.Version(), v)
	}
	if r["ts"] != s {
		t.Errorf("update-twin parameter = %q, want %q", r["ts"], s)
	}
}

// TODO: very flaky
func testSubscribeTwin(t *testing.T, sc *iotservice.Client, dc *iotdevice.Client) {
	sub, err := dc.SubscribeTwinUpdates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer dc.UnsubscribeTwinUpdates(sub)

	// TODO: hacky, but reduces flakiness
	time.Sleep(time.Second)

	twin, err := sc.UpdateDeviceTwin(context.Background(), &iotservice.Twin{
		DeviceID: dc.DeviceID(),
		Tags: map[string]interface{}{
			"test-device": true,
		},
		Properties: &iotservice.Properties{
			Desired: map[string]interface{}{
				"test-prop": time.Now().UnixNano() / 1000,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case state := <-sub.C():
		if state["$version"] != twin.Properties.Desired["$version"] {
			t.Errorf("version = %d, want %d", state["$version"], twin.Properties.Desired["$version"])
		}
		if state["test-prop"] != twin.Properties.Desired["test-prop"] {
			t.Errorf("test-prop = %q, want %q", state["test-prop"], twin.Properties.Desired["test-prop"])
		}
	case <-time.After(10 * time.Second):
		t.Fatal("SubscribeTwinUpdates timed out")
	}
}

func testDirectMethod(t *testing.T, sc *iotservice.Client, dc *iotdevice.Client) {
	if err := dc.RegisterMethod(
		context.Background(),
		"sum",
		func(v map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": v["a"].(float64) + v["b"].(float64),
			}, nil
		},
	); err != nil {
		t.Fatal(err)
	}

	resc := make(chan *iotservice.MethodResult, 1)
	errc := make(chan error, 2)
	go func() {
		v, err := sc.CallDeviceMethod(context.Background(), dc.DeviceID(), &iotservice.MethodCall{
			MethodName:      "sum",
			ConnectTimeout:  5,
			ResponseTimeout: 5,
			Payload: map[string]interface{}{
				"a": 1.5,
				"b": 3,
			},
		})
		if err != nil {
			errc <- err
		}
		resc <- v
	}()

	select {
	case v := <-resc:
		w := &iotservice.MethodResult{
			Status: 200,
			Payload: map[string]interface{}{
				"result": 4.5,
			},
		}
		if !reflect.DeepEqual(v, w) {
			t.Errorf("direct-method result = %v, want %v", v, w)
		}
	case err := <-errc:
		t.Fatal(err)
	}
}

func genID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
