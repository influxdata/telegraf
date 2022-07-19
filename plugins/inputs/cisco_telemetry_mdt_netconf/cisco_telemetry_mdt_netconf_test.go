package cisco_telemetry_mdt_netconf

import (
	"context"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"

	"github.com/cisco-ie/netgonf/netconf"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const (
	goodConn = iota
	badConnPort
	badConnKey
	badConnUser
	badConnPass
	goodMainConfig
	goodMainConfigNoSubs
	goodMainConfigNoGets
	badMainConfigNoRequests
	badMainConfigNoXpaths
	goodMainConfigNoTags
	dummySession
	badSession
	goodDialinService
	goodDialinTelemetryService
	goodDialinEventNotificationService
	badDialinServiceNoRequest
	goodDialinServiceCancelImmediately
	goodGetService
	badGetServiceNoRequest
	goodGetServiceCancelImmediately
	badDialinRequestInvalidPath
	badDialinRequestPeriodicMissingPeriod
	badDialinRequestMissingUpdateTrigger
	goodDialinRequestOnChange
	badDialinRequestOnChange
	badDialinRequestOnChangeMissingPeriod
	badGetRequestInvalidPath
	goodEventNotificationRequest
	goodDialinReceive
	badDialinReceive
	goodTelemetryTree
	goodEmptyTelemetryTree
	badTelemetryTree
)

const privateKey = "-----BEGIN OPENSSH PRIVATE KEY-----\n" +
	"b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAlwAAAAdzc2gtcn\n" +
	"NhAAAAAwEAAQAAAIEAok+C0A2WtrE6vCy53cRzvRMcP4wY3Ygiktv1PEXHWVCV13rGZfb+\n" +
	"SOUbIbIvEdhqtSf+FOyJKlzEovotxRwcn/uSf1KetlCToAnD02rB6bZ/4om6W+U6qE7FPn\n" +
	"qVj9J5Nxtrra69FZuQOk4MnAUVRklJB9Zh405vxAj5qwEYPVUAAAIQ5/Vl8+f1ZfMAAAAH\n" +
	"c3NoLXJzYQAAAIEAok+C0A2WtrE6vCy53cRzvRMcP4wY3Ygiktv1PEXHWVCV13rGZfb+SO\n" +
	"UbIbIvEdhqtSf+FOyJKlzEovotxRwcn/uSf1KetlCToAnD02rB6bZ/4om6W+U6qE7FPnqV\n" +
	"j9J5Nxtrra69FZuQOk4MnAUVRklJB9Zh405vxAj5qwEYPVUAAAADAQABAAAAgHmHKxz4b7\n" +
	"ZOsPmgS3J+22HgYzA5h4ynl6t6Qf5lCMQZEHiMluxVqUOPN2ddcNzdu9f0H8wu5uzvFNQq\n" +
	"mgaR6+Oz6qrhKAm/OejNZz+7IadhUUZoH3Z/jS93nawzhGPfaQPSf4vR6b6uHi1s0dce//\n" +
	"P7pdrY21O8OlUyd9OiVFdpAAAAQF9iVgSzH4XOr9RdWkIF7gqaulTsNjdLXtcWQXfQsfZ6\n" +
	"KQYuP5Ybs9RAVvmbl/LEip9dq+YG6cAiQEUi5UFJWiIAAABBANY2QG4623X8RTT4CQMzoZ\n" +
	"ErXIh791lnsnMQZcu3LKhX7esZ+1t/Jek65dXc5h/xQ6rpp8oTqpstyR5CrIJNJLsAAABB\n" +
	"AMH5Tj3Y1kbzbMgLnCX1wPz6ao5S0cZ0bB/QagixS2LoWeVv5CE701Znl4uA8Kp0LjbbYa\n" +
	"64sAwn2zFuf2WjDS8AAAAWY3ByZWN1cEBDUFJFQ1VQLU0tMzBIRAECAwQF\n" +
	"-----END OPENSSH PRIVATE KEY-----"

const publicKey = "[127.0.0.1]:2200 ssh-rsa AAAAB3NzaC1yc2EAAAA" +
	"DAQABAAAAgQCiT4LQDZa2sTq8LLndxHO9Exw/jBjdiCKS2/U8RcdZUJXX" +
	"esZl9v5I5Rshsi8R2Gq1J/4U7IkqXMSi+i3FHByf+5J/Up62UJOgCcPTa" +
	"sHptn/iibpb5TqoTsU+epWP0nk3G2utrr0Vm5A6TgycBRVGSUkH1mHjTm" +
	"/ECPmrARg9VQ=="

var goodTelemetrySampleXML = []byte(`<interfaces-state xmlns="urn:ietf:params:` +
	`xml:ns:yang:ietf-interfaces"><interface><name>GigabitEthernet1</name>` +
	`<admin-status>up</admin-status><oper-status>up</oper-status><last-change>` +
	`1970-01-01T00:00:00.00001+00:00</last-change><if-index>1</if-index>` +
	`<phys-address>00:42:42:42:42:42</phys-address><speed>1024000000</speed>` +
	`</interface></interfaces-state>`)

var goodEventNotificationSampleXML = []byte(`<eventTime>1970-01-01T00:00:00.00001+00:00` +
	`</eventTime><alarm-notification xmlns='http://tail-f.com/ns/ncs-alarms'>` +
	`<alarm-class>changed-alarm</alarm-class><device>xrv</device>` +
	`<type xmlns:al="http://tail-f.com/ns/ncs-alarms">al:out-of-sync</type>` +
	`<managed-object xmlns:ncs="http://tail-f.com/ns/ncs">/ncs:devices/ncs:` +
	`device[ncs:name='xrv']</managed-object>` +
	`<event-type>operationalViolation</event-type><has-clear>true</has-clear>` +
	`<kind-of-alarm>root-cause</kind-of-alarm><probable-cause>0</probable-cause>` +
	`<event-time>1970-01-01T00:00:00.00001+00:00</event-time><perceived-severity>` +
	`major</perceived-severity><alarm-text>got: 1000000444 expected: 1000000443` +
	`</alarm-text></alarm-notification>`)

var badSampleXML = []byte(`<asdf><asdf-node><asdf-name>asdf</asdf-name></asdf>`)

// A mock of a request to create a telemetry subscription through dial-in
var mockDTSRequest = dialinSubscriptionRequest{
	XPathFilter:   "/if:interfaces-state/interface",
	Tags:          []string{"/if:interfaces-state/interface/name"},
	UpdateTrigger: "periodic",
	Period:        config.Duration(1 * time.Second),
}

// A mock of a request to create an event notification subscription
var mockDENSRequest = notificationSubscriptionRequest{
	Stream: "ncs-alarms",
	Tags:   []string{"ncs-alarms:alarm-notification/alarm-class"},
}

// A mock of a request to create a get operation
var mockGRequest = getRequest{
	SelectFilter: "/memory-statistics/memory-statistic",
	Tags:         []string{"/memory-statistics/memory-statistic/name"},
	Period:       config.Duration(1 * time.Second),
}

// sessionMockGoodTelemetryReceive embeds the netconf.Session in order to
// override the Receive() method: returns a good telemetry example in XML format
type sessionMockGoodTelemetryReceive struct {
	netconf.Session
}

// sessionMockBadTelemetryReceive embeds the netconf.Session in order to
// override the Receive() method: returns a bad telemetry example in XML format
type sessionMockBadTelemetryReceive struct {
	sessionMockGoodTelemetryReceive
}

// sessionMockGoodEventNotificationReceive embeds the netconf.Session in order to
// override the Receive() method: returns a good event notification example in XML format
type sessionMockGoodEventNotificationReceive struct {
	netconf.Session
}

// sessionMockBadEventNotificationReceive embeds the netconf.Session in order to
// override the Receive() method: returns a bad event notification example in XML format
type sessionMockBadEventNotificationReceive struct {
	sessionMockGoodEventNotificationReceive
}

// Read user input because we require a connection to a real NETCONF server.
// Run this test with optional arguments tXpathf, tXpathk, enStream, enXpathk:
// go test -args -tServer=X.X.X.X:830 -tUser="asdf" -tPassword="asdf"
// -tKey=":830 ssh-rsa ..." -tXpathf="asdfpath:asdf2/asdf3" -tXpathk="asdfkey"
// -enServer=X.X.X.X:2022 -enUser="asdf" -enPassword="asdf"
// -enKey=":2022 ssh-rsa ..." -enStream="ncs-alarms"

var tServer, tUser, tPassword, tKey, tXpathf, tXpathk,
	enServer, enUser, enPassword, enKey, enStream *string

func init() {
	// Read parameters of the NETCONF server to be used for tests of the
	// dial-in telemetrysubscriptions and get operations
	tServer = flag.String("tServer", "", "127.0.0.0:830")
	tUser = flag.String("tUser", "user", "user")
	tPassword = flag.String("tPassword", "password", "password")
	tKey = flag.String("tKey", "", publicKey)
	tXpathf = flag.String("tXpathf", "", mockDTSRequest.XPathFilter)
	tXpathk = flag.String("tXpathk", "", mockDTSRequest.Tags[0])

	// Read parameters of the NETCONF server to be used for test of an
	// event notification subscription
	enServer = flag.String("enServer", "", "127.0.0.0:830")
	enUser = flag.String("enUser", "user", "user")
	enPassword = flag.String("enPassword", "password", "password")
	enKey = flag.String("enKey", "", publicKey)
	enStream = flag.String("enStream", "", mockDENSRequest.Stream)
}

// Receive is a good mock wrapper around netconf.Session's Receive() method
func (s *sessionMockGoodTelemetryReceive) Receive(response interface{}) error {
	response.(*netconfYangPush).Notification = netconf.Notification{XMLName: xml.Name{
		Local: "interfaces-state", Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
	}}
	response.(*netconfYangPush).EventTime = time.Now()
	response.(*netconfYangPush).PushUpdate.Content.InnerXML = goodTelemetrySampleXML
	return nil
}

// Receive is a bad mock wrapper around netconf.Session's Receive() method
func (s *sessionMockBadTelemetryReceive) Receive(response interface{}) error {
	response.(*netconfYangPush).Notification = netconf.Notification{XMLName: xml.Name{
		Local: "interfaces-state", Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
	}}
	response.(*netconfYangPush).EventTime = time.Now()
	response.(*netconfYangPush).PushUpdate.Content.InnerXML = badSampleXML
	return nil
}

// Receive is a good mock wrapper around netconf.Session's Receive() method
func (s *sessionMockGoodEventNotificationReceive) Receive(response interface{}) error {
	response.(*netconfYangEvent).Notification = netconf.Notification{XMLName: xml.Name{
		Local: "alarm-notification", Space: "http://tail-f.com/ns/ncs-alarms"}}
	response.(*netconfYangEvent).EventTime = time.Now()
	response.(*netconfYangEvent).InnerXML = goodEventNotificationSampleXML
	return nil
}

// Receive is a bad mock wrapper around netconf.Session's Receive() method
func (s *sessionMockBadEventNotificationReceive) Receive(response interface{}) error {
	response.(*netconfYangEvent).Notification = netconf.Notification{XMLName: xml.Name{
		Local: "alarm-notification", Space: "http://tail-f.com/ns/ncs-alarms"}}
	response.(*netconfYangEvent).EventTime = time.Now()
	response.(*netconfYangEvent).InnerXML = badSampleXML
	return nil
}

func mockMetric() testutil.Metric {
	return testutil.Metric{
		Measurement: "ietf-interfaces:interfaces-state/interface",
		Tags: map[string]string{
			"source": "127.0.0.1",
			"ietf-interfaces:interfaces-state/interface/name": "GigabitEthernet1",
		},
		Fields: map[string]interface{}{
			"admin-status": "up", "if-index": uint64(1),
			"last-change": "1970-01-01T00:00:00.00001+00:00",
			"oper-status": "up", "phys-address": "00:42:42:42:42:42",
			"speed": uint64(1024000000),
		},
	}
}

func mockEventNotification() testutil.Metric {
	return testutil.Metric{
		Measurement: "ncs-alarms:alarm-notification",
		Tags: map[string]string{
			"source":                "127.0.0.1",
			mockDENSRequest.Tags[0]: "changed-alarm",
		},
		Fields: map[string]interface{}{
			"type": "al:out-of-sync", "has-clear": true,
			"event-type": "operationalViolation", "perceived-severity": "major",
			"managed-object": "/ncs:devices/ncs:device[ncs:name='xrv']",
			"alarm-text":     "got: 1000000444 expected: 1000000443", "device": "xrv",
			"kind-of-alarm": "root-cause", "probable-cause": uint64(0),
			"event-time": "1970-01-01T00:00:00.00001+00:00",
		},
	}
}

func mockCiscoTelemetryNETCONF(dsrs *dialinSubscriptionRequestsService,
	grs *getRequestsService, isa bool, options ...string) *CiscoTelemetryNETCONF {

	// Default
	c := &CiscoTelemetryNETCONF{ServerAddress: "127.0.0.1:2200",
		Username: "user", Password: "password",
		IgnoreServerAuthenticity: false,
		ServerPublicKey:          publicKey,
		Redial:                   config.Duration(10 * time.Second),
		Dsrs: &dialinSubscriptionRequestsService{
			Subscriptions: []dialinSubscriptionRequest{mockDTSRequest},
			Notifications: []notificationSubscriptionRequest{mockDENSRequest},
			setting:       &setting{},
		},
		Grs: &getRequestsService{
			Gets:    []getRequest{mockGRequest},
			setting: &setting{},
		},
	}

	// Overwrite default
	for i, o := range options {
		if o != "" {
			switch i {
			case 0:
				c.ServerAddress = o
				c.ServerPublicKey = strings.Replace(
					c.ServerPublicKey, "127.0.0.1:2200", o, 1)
			case 1:
				c.ServerPublicKey = o
			case 2:
				c.Username = o
			case 3:
				c.Password = o
			}
		}
	}
	if isa == true {
		c.IgnoreServerAuthenticity = isa
		c.ServerPublicKey = ""
	}

	if dsrs != nil {
		c.Dsrs = dsrs
	}

	if grs != nil {
		c.Grs = grs
	}

	return c
}

func mockSetting(options ...interface{}) *setting {
	// Default
	s := &setting{Mutex: new(sync.Mutex)}

	// Overwrite default
	for i, o := range options {
		if o != nil {
			switch i {
			case 0:
				s.Client = o.(netconf.Client)
			case 1:
				s.Session = o.(Session)
			}
		}
	}

	return s
}

// TestCiscoTelemetryNETCONF_connectClient tests good and bad connection configurations
// The current tests assume there will be an error thrown when the connection fails and
// checks for these errors accordingly.
func TestCiscoTelemetryNETCONF_connectClient(t *testing.T) {
	authenticationErrorMessage := "failed to dial server: ssh: " +
		"handshake failed: ssh: unable to authenticate"
	type args struct {
		ctx context.Context
		s   *setting
	}
	tests := []struct {
		name       string
		c          *CiscoTelemetryNETCONF
		args       args
		wantErrors []error
		testType   int
	}{
		{
			name:     "GoodConnectionSecure",
			c:        mockCiscoTelemetryNETCONF(nil, nil, false),
			testType: goodConn,
		},
		{
			name:     "GoodConnectionInsecure",
			c:        mockCiscoTelemetryNETCONF(nil, nil, true),
			testType: goodConn,
		},
		{
			name: "BadConnectionPort",
			c: mockCiscoTelemetryNETCONF(nil, nil, false,
				"127.0.0.1:2300"),
			wantErrors: []error{
				errors.New("failed to dial server: dial tcp 127.0.0.1:2300: " +
					"connect: connection refused"),
				errors.New("failed to dial server: dial tcp 127.0.0.1:2300: " +
					"connectex: No connection could be made because the " +
					"target machine actively refused it.")},
			testType: badConnPort,
		},
		{
			name: "BadConnectionKey",
			c: mockCiscoTelemetryNETCONF(nil, nil, false,
				"", "[127.0.0.1]:2200 ssh-rsa asdasd"),
			wantErrors: []error{errors.New(
				"failed to dial server: cannot parse public key for server" +
					" 127.0.0.1:2200 - illegal base64 data at input byte 4")},
			testType: badConnKey,
		},
		{
			name: "BadConnectionUsername",
			c: mockCiscoTelemetryNETCONF(nil, nil, false,
				"", "", "asdasd"),
			wantErrors: []error{errors.New(authenticationErrorMessage)},
			testType:   badConnUser,
		},
		{
			name: "BadConnectionPassword",
			c: mockCiscoTelemetryNETCONF(nil, nil, false,
				"", "", "", "asdasd"),
			wantErrors: []error{errors.New(authenticationErrorMessage)},
			testType:   badConnPass,
		},
	}

	// Create SSH server
	listener, err := net.Listen("tcp", "127.0.0.1:2200")
	if err != nil {
		t.Error("Mock server failed to listen for SSH connections:", err)
	} else {
		log.Println("Mock server listening...")

		waitgroup := new(sync.WaitGroup)
		waitgroup.Add(1)
		mainCtx, mainCancel := context.WithCancel(context.Background())

		go func(listener net.Listener, waitgroup *sync.WaitGroup) {

			config := &ssh.ServerConfig{
				PasswordCallback: func(c ssh.ConnMetadata,
					pass []byte) (*ssh.Permissions, error) {
					if c.User() == "user" && string(pass) == "password" {
						return nil, nil
					}
					return nil, fmt.Errorf("Wrong password for %q", c.User())
				},
			}

			parsedKey, err := ssh.ParsePrivateKey([]byte(privateKey))
			if err != nil {
			}
			config.AddHostKey(parsedKey)

			for mainCtx.Err() == nil {
				conn, err := listener.Accept()
				if err != nil {
				} else {
					_, _, _, err = ssh.NewServerConn(conn, config)
					if err != nil {
					} else {
						log.Println("Client connected to server.")
					}
				}
			}
			waitgroup.Done()
		}(listener, waitgroup)

		for _, tt := range tests {
			acc := &testutil.Accumulator{}
			ctx, cancel := context.WithCancel(context.Background())
			tt.args.s = &setting{}
			tt.args.ctx = ctx
			tt.c.cancel = cancel
			tt.c.acc = acc

			w := new(sync.WaitGroup)
			w.Add(1)
			go func() {
				t.Run(tt.name, func(t *testing.T) {
					tt.c.connectClient(tt.args.ctx, tt.args.s)
				})
				w.Done()
			}()

			time.Sleep(100 * time.Millisecond)
			cancel()
			switch tt.testType {
			case goodConn:
				assert.NotEqual(t, tt.args.s.Client, nil)
				tt.args.s.Client.Close()
			case badConnPort:
				log.Println("Context's error: ", tt.args.ctx.Err())
				log.Println("Received error: ", acc.Errors)
				log.Println("Wanted error: ", tt.wantErrors)
				assert.Condition(t, func() bool {
					for _, e := range acc.Errors {
						if e.Error() == tt.wantErrors[0].Error() ||
							e.Error() == tt.wantErrors[1].Error() {
							return true
						}
					}

					// Windows specific fix
					if tt.args.ctx.Err().Error() == "context canceled" {
						return true
					}
					return false
				})
			case badConnKey:
				assert.Contains(t, acc.Errors, tt.wantErrors[0])
			case badConnUser, badConnPass:
				assert.Contains(t, acc.FirstError().Error(), tt.wantErrors[0].Error())
			}
			tt.c.Stop()
			w.Wait()
		}
		mainCancel()
		listener.Close()
		waitgroup.Wait()
	}
}

// TestCiscoTelemetryNETCONF_handleTelemetry tests if xpaths are trimmed correctly.
func TestCiscoTelemetryNETCONF_handleTelemetry(t *testing.T) {
	type args struct {
		tt TelemetryTree
		t  time.Time
	}
	tests := []struct {
		name       string
		c          *CiscoTelemetryNETCONF
		wantMetric testutil.Metric
		wantError  string
		args       args
		testType   int
	}{
		{
			name:       "GoodTelemetryTree",
			c:          mockCiscoTelemetryNETCONF(nil, nil, false),
			wantMetric: mockMetric(),
			args: args{
				tt: TelemetryTree{
					XMLName: xml.Name{
						Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
						Local: "interfaces-state",
					},
					Children: []TelemetryTree{
						{
							XMLName: xml.Name{
								Local: "interface",
							},
							Children: []TelemetryTree{
								{
									XMLName: xml.Name{
										Local: "name",
									},
									Value: "GigabitEthernet1",
								},
								{
									XMLName: xml.Name{
										Local: "admin-status",
									},
									Value: "up",
								},
								{
									XMLName: xml.Name{
										Local: "oper-status",
									},
									Value: "up",
								},
								{
									XMLName: xml.Name{
										Local: "last-change",
									},
									Value: "1970-01-01T00:00:00.00001+00:00",
								},
								{
									XMLName: xml.Name{
										Local: "if-index",
									},
									Value: uint64(1),
								},
								{
									XMLName: xml.Name{
										Local: "phys-address",
									},
									Value: "00:42:42:42:42:42",
								},
								{
									XMLName: xml.Name{
										Local: "speed",
									},
									Value: uint64(1024000000),
								},
							},
						},
					},
				},
				t: time.Now(),
			},
			testType: goodTelemetryTree,
		},
		{
			name: "GoodEmptyTelemetryTree",
			c:    mockCiscoTelemetryNETCONF(nil, nil, false),
			args: args{
				tt: TelemetryTree{},
				t:  time.Now(),
			},
			testType: goodEmptyTelemetryTree,
		},
		{
			name: "BadTelemetryTree",
			c:    mockCiscoTelemetryNETCONF(nil, nil, false),
			wantError: "failed to traverse telemetry tree: field name has " +
				"value 1 with unsupported tag data type uint32",
			args: args{
				tt: TelemetryTree{XMLName: xml.Name{
					Space: "urn:ietf:params:xml:ns:yang:ietf-interfaces",
					Local: "interfaces-state",
				},
					Children: []TelemetryTree{
						{
							XMLName: xml.Name{
								Local: "interface",
							},
							Children: []TelemetryTree{
								{
									XMLName: xml.Name{
										Local: "name",
									},
									Value: uint32(1),
								},
								{
									XMLName: xml.Name{
										Local: "admin-status",
									},
									Value: "up",
								},
							},
						},
					},
				},
				t: time.Now(),
			},
			testType: badTelemetryTree,
		},
	}
	for _, tt := range tests {
		acc := &testutil.Accumulator{}
		tt.c.acc = acc

		t.Run(tt.name, func(t *testing.T) {
			tt.c.preparePaths()
			tt.c.handleTelemetry(tt.args.tt, tt.args.t)

			switch tt.testType {
			case goodTelemetryTree:
				assert.NoError(t, acc.FirstError())
				m, ok := acc.Get(tt.wantMetric.Measurement)
				if assert.Exactly(t, ok, true) {
					assert.Exactly(t, m.Tags, tt.wantMetric.Tags)
					assert.Exactly(t, m.Fields, tt.wantMetric.Fields)
				} else {
					t.Fail()
				}
			case goodEmptyTelemetryTree:
				assert.NoError(t, acc.FirstError())
				assert.Zero(t, len(acc.Metrics))
			case badTelemetryTree:
				if assert.Error(t, acc.FirstError()) {
					assert.EqualError(t, acc.FirstError(), tt.wantError)
				}
				assert.Zero(t, len(acc.Metrics))
			}
		})
	}
}

// TestCiscoTelemetryNETCONF_trimXpath tests if xpaths are trimmed correctly.
func TestCiscoTelemetryNETCONF_trimXpath(t *testing.T) {
	type args struct {
		xpath string
	}
	tests := []struct {
		name      string
		c         *CiscoTelemetryNETCONF
		args      args
		wantXpath string
	}{
		{
			name:      "GoodXpath1",
			args:      args{xpath: "/asdf:gh-jk/qwerty"},
			wantXpath: "gh-jk/qwerty",
		},
		{
			name:      "GoodXpath2",
			args:      args{xpath: "/gh-jk/qwerty/uiop"},
			wantXpath: "gh-jk/qwerty/uiop",
		},
		{
			name:      "GoodXpath3",
			args:      args{xpath: "/gh-jk/qwerty/bnm[uiop=zxcv]"},
			wantXpath: "gh-jk/qwerty/bnm",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if trimmedXpath := tt.c.trimXpath(tt.args.xpath); trimmedXpath != tt.wantXpath {
				t.Errorf("test trimXpath() failed: got %v, but expected %v",
					trimmedXpath, tt.wantXpath)
			}
		})
	}
}

// TestCiscoTelemetryNETCONF_Start tests the start of the CiscoTelemetryNETCONF plugin
// Start() will always return nil.
func TestCiscoTelemetryNETCONF_Start(t *testing.T) {
	type args struct {
		acc telegraf.Accumulator
	}
	tests := []struct {
		name        string
		c           *CiscoTelemetryNETCONF
		args        args
		wantLengths []int
		testType    int
	}{
		{
			name:        "GoodMainConfig",
			c:           mockCiscoTelemetryNETCONF(nil, nil, false),
			args:        args{&testutil.Accumulator{}},
			wantLengths: []int{3, 3},
			testType:    goodMainConfig,
		},
		{
			name: "GoodMainConfigNoSubs",
			c: mockCiscoTelemetryNETCONF(
				&dialinSubscriptionRequestsService{}, nil, false),
			args:        args{&testutil.Accumulator{}},
			wantLengths: []int{1, 1},
			testType:    goodMainConfigNoSubs,
		},
		{
			name:        "GoodMainConfigNoGets",
			c:           mockCiscoTelemetryNETCONF(nil, &getRequestsService{}, false),
			args:        args{&testutil.Accumulator{}},
			wantLengths: []int{2, 2},
			testType:    goodMainConfigNoGets,
		},
		{
			name: "BadMainConfigNoRequests",
			c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{},
				&getRequestsService{}, false),
			wantLengths: []int{0, 0},
			testType:    badMainConfigNoRequests,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, tt.c.Start(tt.args.acc))
		})

		time.Sleep(100 * time.Millisecond)
		tt.c.Stop()
		assert.Exactly(t, len(tt.c.userXpaths), tt.wantLengths[0])
		assert.Exactly(t, len(tt.c.userTags), tt.wantLengths[1])

		switch tt.testType {
		case goodMainConfig:
			assert.NotZero(t, tt.c.Dsrs.setting)
			assert.NotZero(t, tt.c.Grs.setting)
		case goodMainConfigNoSubs:
			assert.Zero(t, tt.c.Dsrs.setting)
		case goodMainConfigNoGets:
			assert.Zero(t, tt.c.Grs.setting)
		case badMainConfigNoRequests:
			assert.Zero(t, tt.c.Dsrs.setting)
			assert.Zero(t, tt.c.Grs.setting)
		}
		tt.c.Stop()
	}
}

// TestCiscoTelemetryNETCONF_preparePaths tests the preparePaths() method
func TestCiscoTelemetryNETCONF_preparePaths(t *testing.T) {
	tests := []struct {
		name        string
		c           *CiscoTelemetryNETCONF
		wantLengths []int
		wantXpath   []string
		wantTags    []string
		testType    int
	}{
		{
			name:        "GoodMainConfig",
			c:           mockCiscoTelemetryNETCONF(nil, nil, false),
			wantLengths: []int{3, 3},
			wantXpath: []string{"interfaces-state/interface",
				mockDENSRequest.Stream, "memory-statistics/memory-statistic"},
			wantTags: []string{"interfaces-state/interface/name",
				"alarm-notification/alarm-class",
				"memory-statistics/memory-statistic/name"},
			testType: goodMainConfig,
		},
		{
			name: "GoodMainConfigNoSubs",
			c: mockCiscoTelemetryNETCONF(
				&dialinSubscriptionRequestsService{}, nil, false),
			wantLengths: []int{1, 1},
			wantXpath:   []string{"memory-statistics/memory-statistic"},
			wantTags:    []string{"memory-statistics/memory-statistic/name"},
			testType:    goodMainConfigNoSubs,
		},
		{
			name:        "GoodMainConfigNoGets",
			c:           mockCiscoTelemetryNETCONF(nil, &getRequestsService{}, false),
			wantLengths: []int{2, 2},
			wantXpath: []string{"interfaces-state/interface",
				mockDENSRequest.Stream},
			wantTags: []string{"interfaces-state/interface/name",
				"alarm-notification/alarm-class"},
			testType: goodMainConfigNoGets,
		},
		{
			name: "BadMainConfigNoRequests",
			c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{},
				&getRequestsService{}, false),
			wantLengths: []int{0, 0},
			testType:    badMainConfigNoRequests,
		},
		{
			name: "BadMainConfigNoXpaths",
			c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
				Subscriptions: []dialinSubscriptionRequest{{
					UpdateTrigger: "periodic",
					Tags:          []string{"/if:interfaces-state/interface/name"},
					Period:        config.Duration(10 * time.Second),
				}},
				setting: &setting{},
			}, &getRequestsService{
				Gets: []getRequest{{
					Tags:   []string{"/memory-statistics/memory-statistic/name"},
					Period: config.Duration(10 * time.Second),
				}},
				setting: &setting{},
			}, false),
			wantLengths: []int{0, 0},
			testType:    badMainConfigNoXpaths,
		},
		{
			name: "GoodMainConfigNoTags",
			c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
				Subscriptions: []dialinSubscriptionRequest{{
					XPathFilter:   mockDTSRequest.XPathFilter,
					UpdateTrigger: mockDTSRequest.UpdateTrigger,
					Period:        config.Duration(10 * time.Second),
				}},
				setting: &setting{},
			}, &getRequestsService{
				Gets: []getRequest{{
					SelectFilter: mockGRequest.SelectFilter,
					Period:       config.Duration(10 * time.Second),
				}},
				setting: &setting{},
			}, false),
			wantLengths: []int{2, 0},
			wantXpath: []string{"interfaces-state/interface",
				"memory-statistics/memory-statistic"},
			wantTags: []string{},
			testType: goodMainConfigNoTags,
		},
	}
	for _, tt := range tests {
		acc := &testutil.Accumulator{}
		tt.c.acc = acc
		t.Run(tt.name, func(t *testing.T) {
			tt.c.preparePaths()
		})

		// Check that preparePaths() built the xpath maps
		assert.Exactly(t, len(tt.c.userXpaths), tt.wantLengths[0])
		assert.Exactly(t, len(tt.c.userTags), tt.wantLengths[1])
		for _, s := range tt.c.Dsrs.Subscriptions {
			if tt.wantLengths[0] > 0 {
				assert.Contains(t, tt.c.userXpaths, tt.wantXpath[0])
				for range s.Tags {
					assert.Contains(t, tt.c.userTags, tt.wantTags[0])
				}
			}
		}
		for _, g := range tt.c.Grs.Gets {
			if tt.wantLengths[0] >= 1 {
				var idx uint
				if tt.wantLengths[0] > 1 {
					idx = 1
				}
				assert.Contains(t, tt.c.userXpaths, tt.wantXpath[idx])
				for range g.Tags {
					assert.Contains(t, tt.c.userTags, tt.wantTags[idx])
				}
			}
		}
	}
}

// Test_setting_cleanup tests the cleanup() method
func Test_setting_cleanup(t *testing.T) {
	tests := []struct {
		name     string
		s        *setting
		testType int
	}{
		{
			name:     "DummySession",
			s:        mockSetting(),
			testType: dummySession,
		},
		{
			name: "BadSession",
			s: mockSetting(netconf.NewClientSSH(&ssh.Client{}),
				&netconf.Session{}),
			testType: badSession,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.testType {
			case dummySession:
				assert.NoError(t, tt.s.cleanup())
			case badSession:
				assert.Panics(t, func() { tt.s.cleanup() })
			}
		})
	}
}

// TestCiscoTelemetryNETCONF_Stop tests the Stop() method
func TestCiscoTelemetryNETCONF_Stop(t *testing.T) {
	tests := []struct {
		name string
		c    *CiscoTelemetryNETCONF
	}{
		{
			name: "GoodMainConfig",
			c:    mockCiscoTelemetryNETCONF(nil, nil, false),
		},
	}

	for _, tt := range tests {
		_, cancel := context.WithCancel(context.Background())
		tt.c.cancel = cancel
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() { tt.c.Stop() })
		})
	}
}

// Test_dialinSubscriptionRequestsService_createService tests the createService()
// method of the dial-in subscription service
func Test_dialinSubscriptionRequestsService_createService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test. Run this test separately with IOS-XE connection details given as t* arguments and NSO connection details given as en* arguments.")
	}

	type args struct {
		ctx context.Context
		c   *CiscoTelemetryNETCONF
	}
	tests := []struct {
		name      string
		dsrs      *dialinSubscriptionRequestsService
		args      args
		wantState int
		testType  int
	}{
		{
			name: "GoodDialinTelemetryService",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
					Subscriptions: []dialinSubscriptionRequest{mockDTSRequest}},
					&getRequestsService{}, false),
			},
			wantState: ended,
			testType:  goodDialinTelemetryService,
		},
		{
			name: "GoodDialinEventNotficationService",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
					Notifications: []notificationSubscriptionRequest{
						mockDENSRequest}}, &getRequestsService{}, false),
			},
			wantState: ended,
			testType:  goodDialinEventNotificationService,
		},
		{
			name: "BadDialinServiceNoRequest",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{},
					nil, false),
			},
			wantState: connected,
			testType:  badDialinServiceNoRequest,
		},
		{
			name: "GoodDialinServiceCancelImmediately",
			args: args{
				c: mockCiscoTelemetryNETCONF(nil, nil, false),
			},
			wantState: 0,
			testType:  goodDialinServiceCancelImmediately,
		},
	}
	for _, tt := range tests {
		// Read user input because we require a connection to a real NETCONF server.
		switch tt.testType {
		case goodDialinTelemetryService, badDialinServiceNoRequest:
			tt.args.c.ServerAddress = *tServer
			tt.args.c.Username = *tUser
			tt.args.c.Password = *tPassword
			tt.args.c.ServerPublicKey = *tKey
			if tt.testType == goodDialinTelemetryService {
				tt.args.c.Dsrs.Subscriptions[0].XPathFilter = *tXpathf
				tt.args.c.Dsrs.Subscriptions[0].Tags = []string{*tXpathk}
			}
		case goodDialinEventNotificationService:
			tt.args.c.ServerAddress = *enServer
			tt.args.c.Username = *enUser
			tt.args.c.Password = *enPassword
			tt.args.c.ServerPublicKey = *enKey
			tt.args.c.Dsrs.Notifications[0].Stream = *enStream
		}

		acc := &testutil.Accumulator{}
		tt.args.c.acc = acc
		tt.dsrs = tt.args.c.Dsrs
		ctx, cancel := context.WithCancel(context.Background())
		tt.args.ctx, tt.args.c.cancel = ctx, cancel
		tt.dsrs.setting = mockSetting()

		w := new(sync.WaitGroup)
		w.Add(1)
		go func() {
			t.Run(tt.name, func(t *testing.T) {
				tt.args.c.waitgroup.Add(1)
				tt.dsrs.createService(tt.args.ctx, tt.args.c)
			})
			w.Done()
		}()

		switch tt.testType {
		case goodDialinTelemetryService, goodDialinEventNotificationService:
			fallthrough
		case badDialinServiceNoRequest:
			time.Sleep(2 * time.Second)
			fallthrough
		default:
			// Terminate service immediately
			cancel()
			tt.args.c.Stop()
			w.Wait()
		}

		assert.Exactly(t, tt.args.c.Dsrs.state, tt.wantState)
	}
}

// Test_dialinSubscriptionRequestsService_createRequests tests the createRequests()
// method of the dial-in subscription service
func Test_dialinSubscriptionRequestsService_createRequests(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *CiscoTelemetryNETCONF
	}

	tests := []struct {
		name      string
		dsrs      *dialinSubscriptionRequestsService
		args      args
		wantError string
		wantState int
		testType  int
	}{
		{
			name: "GoodDialinRequestPeriodic",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
					Subscriptions: []dialinSubscriptionRequest{mockDTSRequest}},
					&getRequestsService{}, false),
			},
			testType: goodDialinService,
		},
		{
			name: "BadDialinRequest",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{},
					&getRequestsService{}, false),
			},
			wantError: "missing subscription requests",
			wantState: connected,
			testType:  badDialinServiceNoRequest,
		},
		{
			name: "BadDialinRequestPeriodicMissingPeriod",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
					Subscriptions: []dialinSubscriptionRequest{{
						XPathFilter:   mockDTSRequest.XPathFilter,
						UpdateTrigger: mockDTSRequest.UpdateTrigger,
					}},
				}, &getRequestsService{}, false),
			},
			wantError: "failed to create telemetry subscription: missing field " +
				"in subscription 1: period",
			wantState: connected,
			testType:  badDialinRequestPeriodicMissingPeriod,
		},
		{
			name: "BadDialinRequestMissingUpdateTrigger",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
					Subscriptions: []dialinSubscriptionRequest{{
						XPathFilter: mockDTSRequest.XPathFilter,
						Period:      config.Duration(1 * time.Second),
					}},
				}, &getRequestsService{}, false),
			},

			wantError: "failed to create telemetry subscription: " +
				"bad / missing field update_trigger (options are: periodic, on-change)",
			wantState: connected,
			testType:  badDialinRequestMissingUpdateTrigger,
		},
		{
			name: "GoodDialinRequestOnChange",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
					Subscriptions: []dialinSubscriptionRequest{{
						XPathFilter:   "/cdp-ios-xe-oper:cdp-neighbor-details/cdp-neighbor-detail",
						UpdateTrigger: "on-change",
						Period:        config.Duration(0 * time.Second),
					}},
				}, &getRequestsService{}, false),
			},
			wantState: created,
			testType:  goodDialinRequestOnChange,
		},
		{
			name: "BadDialinRequestOnChange",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
					Subscriptions: []dialinSubscriptionRequest{{
						XPathFilter:   "/cdp-ios-xe-oper:cdp-neighbor-details/cdp-neighbor-detail",
						UpdateTrigger: "on-change",
						Period:        config.Duration(1 * time.Second),
					}},
				}, &getRequestsService{}, false),
			},
			wantError: "NETCONF RPC error invalid-value: " +
				"/rpc/notif-bis:establish-subscription/yp:dampening-period: " +
				"\"100\" is out of range.",
			wantState: connected,
			testType:  badDialinRequestOnChange,
		},
		{
			name: "BadDialinRequestOnChangeMissingPeriod",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
					Subscriptions: []dialinSubscriptionRequest{{
						XPathFilter:   "/cdp-ios-xe-oper:cdp-neighbor-details/cdp-neighbor-detail",
						UpdateTrigger: "on-change",
					}},
				}, &getRequestsService{}, false),
			},
			wantError: "failed to create telemetry subscription: " +
				"missing field in subscription 1: period",
			wantState: connected,
			testType:  badDialinRequestOnChangeMissingPeriod,
		},
		{
			name: "GoodEventNotificationRequest",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{
					Notifications: []notificationSubscriptionRequest{mockDENSRequest}},
					&getRequestsService{}, false),
			},
			testType: goodEventNotificationRequest,
		},
	}

	for _, tt := range tests {
		tt.args.c.acc = &testutil.Accumulator{}
		tt.dsrs = tt.args.c.Dsrs
		ctx, cancel := context.WithCancel(context.Background())
		tt.args.ctx, tt.args.c.cancel = ctx, cancel

		// Read user input because we require a connection to a real NETCONF server.
		switch tt.testType {
		case goodDialinService:
			if testing.Short() {
				t.Skip("Skipping integration test. Run this test separately with IOS-XE connection details given as t* arguments.")
			}
		case goodEventNotificationRequest:
			tt.args.c.ServerAddress = *enServer
			tt.args.c.Username = *enUser
			tt.args.c.Password = *enPassword
			tt.args.c.ServerPublicKey = *enKey
			tt.args.c.Dsrs.Notifications[0].Stream = *enStream
		default:
			tt.args.c.ServerAddress = *tServer
			tt.args.c.Username = *tUser
			tt.args.c.Password = *tPassword
			tt.args.c.ServerPublicKey = *tKey
			if len(tt.args.c.Dsrs.Subscriptions) == 1 {
				tt.args.c.Dsrs.Subscriptions[0].XPathFilter = *tXpathf
				tt.args.c.Dsrs.Subscriptions[0].Tags = []string{*tXpathk}
			}
		}

		tt.dsrs.setting = mockSetting()

		w := new(sync.WaitGroup)
		w.Add(1)
		go func() {
			t.Run(tt.name, func(t *testing.T) {
				// Prerequisites: establish connection
				tt.args.c.connectClient(tt.args.ctx, tt.dsrs.setting)

				// Main test
				err := tt.dsrs.createRequests(tt.args.ctx, tt.args.c)

				switch tt.testType {
				case badDialinServiceNoRequest, badDialinRequestPeriodicMissingPeriod,
					badDialinRequestMissingUpdateTrigger, badDialinRequestOnChange,
					badDialinRequestOnChangeMissingPeriod:
					if assert.Error(t, err) {
						assert.EqualError(t, err, tt.wantError)
						assert.Equal(t, tt.dsrs.setting.state, tt.wantState)
					}
				default:
					if assert.NoError(t, err) {
						assert.Equal(t, tt.dsrs.setting.state, created)
					}
				}
			})

			w.Done()
		}()

		// Terminate service
		time.Sleep(5 * time.Second)
		cancel()
		tt.args.c.Stop()
		w.Wait()
	}
}

// Test_dialinSubscriptionRequestsService_receiveTelemetry tests the receiveTelemetry()
// method of the dial-in subscription service
func Test_dialinSubscriptionRequestsService_receiveTelemetry(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *CiscoTelemetryNETCONF
	}
	tests := []struct {
		name       string
		dsrs       *dialinSubscriptionRequestsService
		args       args
		wantMetric testutil.Metric
		wantError  error
		wantState  int
		testType   int
	}{
		{
			name: "GoodDialinReceive",
			args: args{
				c: mockCiscoTelemetryNETCONF(nil, &getRequestsService{}, false),
			},
			wantMetric: mockMetric(),
			wantState:  ended,
			testType:   goodDialinReceive,
		},
		{
			name: "BadDialinReceive",
			args: args{
				c: mockCiscoTelemetryNETCONF(nil, &getRequestsService{}, false),
			},
			wantError: errors.New("failed to unmarshal XML: XML syntax error " +
				"on line 1: element <asdf-node> closed by </asdf>"),
			wantState: ended,
			testType:  badDialinReceive,
		},
	}
	for _, tt := range tests {
		acc := &testutil.Accumulator{}
		tt.args.c.acc = acc
		tt.dsrs = tt.args.c.Dsrs
		ctx, cancel := context.WithCancel(context.Background())
		tt.args.ctx, tt.args.c.cancel = ctx, cancel

		switch tt.testType {
		case goodDialinReceive:
			tt.dsrs.setting = mockSetting(netconf.NewClientSSH(&ssh.Client{}),
				&sessionMockGoodTelemetryReceive{})
		case badDialinReceive:
			tt.dsrs.setting = mockSetting(netconf.NewClientSSH(&ssh.Client{}),
				&sessionMockBadTelemetryReceive{})
		}

		w := new(sync.WaitGroup)
		w.Add(1)

		go func() {
			t.Run(tt.name, func(t *testing.T) {
				// Extract the tags
				tt.args.c.preparePaths()

				tt.dsrs.receiveTelemetry(tt.args.ctx, tt.args.c)
			})

			switch tt.testType {
			case goodDialinReceive:
				assert.NoError(t, acc.FirstError())
				m, ok := acc.Get(tt.wantMetric.Measurement)
				if assert.Exactly(t, ok, true) {
					assert.Exactly(t, m.Tags, tt.wantMetric.Tags)
					assert.Exactly(t, m.Fields, tt.wantMetric.Fields)
				} else {
					t.Fail()
				}
			case badDialinReceive:
				assert.Error(t, acc.FirstError())
				assert.Zero(t, len(acc.Metrics))
				assert.Contains(t, acc.Errors, tt.wantError)
			}
			assert.Exactly(t, tt.dsrs.state, tt.wantState)

			w.Done()
		}()

		// Terminate service
		time.Sleep(10 * time.Millisecond)
		cancel()
		w.Wait()
	}
}

// Test_dialinSubscriptionRequestsService_receiveNotifications tests the receiveNotifications()
// method of the dial-in subscription service
func Test_dialinSubscriptionRequestsService_receiveNotifications(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *CiscoTelemetryNETCONF
	}
	tests := []struct {
		name       string
		dsrs       *dialinSubscriptionRequestsService
		args       args
		wantMetric testutil.Metric
		wantError  error
		wantState  int
		testType   int
	}{
		{
			name: "GoodDialinReceive",
			args: args{
				c: mockCiscoTelemetryNETCONF(nil, &getRequestsService{}, false),
			},
			wantMetric: mockEventNotification(),
			wantState:  ended,
			testType:   goodDialinReceive,
		},
		{
			name: "BadDialinReceive",
			args: args{
				c: mockCiscoTelemetryNETCONF(nil, &getRequestsService{}, false),
			},
			wantError: errors.New("failed to unmarshal XML: XML syntax error " +
				"on line 1: element <asdf-node> closed by </asdf>"),
			wantState: ended,
			testType:  badDialinReceive,
		},
	}
	for _, tt := range tests {
		acc := &testutil.Accumulator{}
		tt.args.c.acc = acc
		tt.dsrs = tt.args.c.Dsrs
		ctx, cancel := context.WithCancel(context.Background())
		tt.args.ctx, tt.args.c.cancel = ctx, cancel

		switch tt.testType {
		case goodDialinReceive:
			tt.dsrs.setting = mockSetting(netconf.NewClientSSH(&ssh.Client{}),
				&sessionMockGoodEventNotificationReceive{})
		case badDialinReceive:
			tt.dsrs.setting = mockSetting(netconf.NewClientSSH(&ssh.Client{}),
				&sessionMockBadEventNotificationReceive{})
		}

		w := new(sync.WaitGroup)
		w.Add(1)

		go func() {
			t.Run(tt.name, func(t *testing.T) {
				// Extract the tags
				tt.args.c.preparePaths()

				tt.dsrs.receiveNotifications(tt.args.ctx, tt.args.c)
			})

			switch tt.testType {
			case goodDialinReceive:
				assert.NoError(t, acc.FirstError())
				m, ok := acc.Get(tt.wantMetric.Measurement)
				if assert.Exactly(t, ok, true) {
					assert.Exactly(t, m.Tags, tt.wantMetric.Tags)
					assert.Exactly(t, m.Fields, tt.wantMetric.Fields)
				} else {
					t.Fail()
				}
			case badDialinReceive:
				assert.Error(t, acc.FirstError())
				assert.Zero(t, len(acc.Metrics))
				assert.Contains(t, acc.Errors, tt.wantError)
			}
			assert.Exactly(t, tt.dsrs.state, tt.wantState)

			w.Done()
		}()

		// Terminate service
		time.Sleep(10 * time.Millisecond)
		cancel()
		w.Wait()
	}
}

// Test_getRequestsService_createService tests the createService()
// method of the get service
func Test_getRequestsService_createService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test. Run this test separately with IOS-XE connection details given as t* arguments.")
	}

	type args struct {
		ctx context.Context
		c   *CiscoTelemetryNETCONF
	}
	tests := []struct {
		name      string
		grs       *getRequestsService
		args      args
		wantState int
		testType  int
	}{
		{
			name: "GoodGetService",
			args: args{
				c: mockCiscoTelemetryNETCONF(nil, nil, false),
			},
			wantState: ended,
			testType:  goodGetService,
		},
		{
			name: "BadGetServiceNoRequest",
			args: args{
				c: mockCiscoTelemetryNETCONF(nil, &getRequestsService{}, false),
			},
			wantState: ended,
			testType:  badGetServiceNoRequest,
		},
		{
			name: "GoodGetServiceCancelImmediately",
			args: args{
				c: mockCiscoTelemetryNETCONF(nil, nil, false),
			},
			wantState: 0,
			testType:  goodGetServiceCancelImmediately,
		},
	}
	for _, tt := range tests {
		// Read user input because we require a connection to a real NETCONF server.
		tt.args.c.ServerAddress = *tServer
		tt.args.c.Username = *tUser
		tt.args.c.Password = *tPassword
		tt.args.c.ServerPublicKey = *tKey

		if tt.testType == goodGetService {
			tt.args.c.Grs.Gets[0].SelectFilter = *tXpathf
			tt.args.c.Grs.Gets[0].Tags = []string{*tXpathk}
		}

		acc := &testutil.Accumulator{}
		tt.args.c.acc = acc
		tt.grs = tt.args.c.Grs
		ctx, cancel := context.WithCancel(context.Background())
		tt.args.ctx, tt.args.c.cancel = ctx, cancel

		w := new(sync.WaitGroup)
		w.Add(1)
		go func() {
			t.Run(tt.name, func(t *testing.T) {
				tt.args.c.waitgroup.Add(1)
				tt.grs.createService(tt.args.ctx, tt.args.c)
			})
			w.Done()
		}()

		switch tt.testType {
		case goodGetService:
			fallthrough
		case badGetServiceNoRequest:
			time.Sleep(2 * time.Second)
			fallthrough
		default:
			// Terminate service immediately
			cancel()
			tt.args.c.Stop()
			w.Wait()
		}

		assert.Exactly(t, tt.args.c.Grs.state, tt.wantState)
	}
}

// Test_getRequestsService_createRequests tests the createRequests()
// method of the get service
func Test_getRequestsService_createRequests(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *CiscoTelemetryNETCONF
	}
	tests := []struct {
		name      string
		grs       *getRequestsService
		args      args
		wantError error
		wantState int
		testType  int
	}{
		{
			name: "GoodGetRequest",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{},
					nil, false),
			},
			wantState: ended,
			testType:  goodGetService,
		},
		{
			name: "GoodGetRequestCancelImmediately",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{},
					nil, false),
			},
			wantState: ended,
			testType:  goodGetServiceCancelImmediately,
		},
		{
			name: "BadGetRequest",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{},
					&getRequestsService{}, false),
			},
			wantError: errors.New("missing get requests"),
			wantState: ended,
			testType:  badGetServiceNoRequest,
		},
		{
			name: "BadGetRequestInvalidPath",
			args: args{
				c: mockCiscoTelemetryNETCONF(&dialinSubscriptionRequestsService{},
					&getRequestsService{
						Gets: []getRequest{{
							SelectFilter: "/asdf/asdf",
							Period:       config.Duration(10 * time.Second),
						}},
					}, false),
			},
			wantError: errors.New("failed to unmarshal XML: EOF"),
			wantState: ended,
			testType:  badGetRequestInvalidPath,
		},
	}

	for _, tt := range tests {
		acc := &testutil.Accumulator{}
		tt.args.c.acc = acc
		tt.grs = tt.args.c.Grs
		ctx, cancel := context.WithCancel(context.Background())
		tt.args.ctx, tt.args.c.cancel = ctx, cancel

		// Read user input because we require a connection to a real NETCONF server.
		tt.args.c.ServerAddress = *tServer
		tt.args.c.Username = *tUser
		tt.args.c.Password = *tPassword
		tt.args.c.ServerPublicKey = *tKey

		tt.args.c.Grs.setting = mockSetting()

		w := new(sync.WaitGroup)
		w.Add(1)

		switch tt.testType {
		case badGetRequestInvalidPath:
			if testing.Short() {
				t.Skip("Skipping integration test. Run this test separately with IOS-XE connection details given as t* arguments.")
			}
		}

		go func() {
			t.Run(tt.name, func(t *testing.T) {
				// Prerequisites: establish connection
				tt.args.c.connectClient(tt.args.ctx, tt.grs.setting)

				// Main test
				err := tt.grs.createRequests(tt.args.ctx, tt.args.c)

				switch tt.testType {
				case badGetServiceNoRequest:
					assert.Error(t, err)
					assert.Equal(t, err, tt.wantError)
				case badGetRequestInvalidPath:
					assert.Contains(t, acc.Errors, tt.wantError)
					fallthrough
				default:
					assert.NoError(t, err)
				}
				assert.Equal(t, tt.grs.setting.state, ended)
			})

			w.Done()
		}()

		// Terminate service
		switch tt.testType {
		case goodGetService, badGetServiceNoRequest, badGetRequestInvalidPath:
			time.Sleep(3 * time.Second)
		}
		cancel()
		tt.args.c.Stop()
		w.Wait()
	}

}
