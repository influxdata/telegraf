package cisco_telemetry_mdt_netconf

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cisco-ie/netgonf/netconf"
	"golang.org/x/crypto/ssh"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Sample configuration
const sampleConfig = `
  ### NETCONF over SSH connection
  ## Address and port
  server_address = "10.10.10.10:830"

  ## Credentials
  username = "cisco"
  password = "cisco"

  ## Enable check for authenticity of the NETCONF server
  ## Unknown servers are ignored by default. Set ignore_server_authenticity
  ## to true to disable the check for authenticity of a server's public key.
  ## Optional, default value: false.
  ignore_server_authenticity = false

  ## Public key of the NETCONF server
  ## Mandatory if ignore_server_authenticity is set to false.
  ## The public key should follow the format of the known_hosts file,
  ## as documented in sshd(8) manual page.
  server_public_key = "[10.10.10.10]:830 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDXxWHGjcEcyEDw/YbJeB824husNnchKKbRtR5i9s+Y712kckQpkWScgwRJJsvneUg4Ztu4ZS8PPzlfiaoHAzOiKjuE7Ns+zklaPSwTj6hf6Sl0FuChWMXi/EchfPcUREQ9mlKL10oMD37W+m3vRUtmj/LM1gNHUSjp3Q1RsyfhLfxYw7I2RQXDfindwxxrX32iWWJdPMfY7PDRYpvh/xmyQVb9RdOhZ7qA/xkDc+SS1hZrzCkh2kaKTd4Glh76K58fEuQ2NFCRYztezWa7D61OiXIeWZJ4x2Utb8xH6wsGA5T0vBt89DB7EvF8xsnEdDtlMsI8L99JtGlNO3MXasdf"

  ## Redial interval if the client fails to connect to the server
  ## Optional, default value: 10s.
  redial = "10s"

  ### Telemetry streaming
  ## IOS-XE Subscription - periodic
  [[inputs.cisco_telemetry_mdt_netconf.subscription]]
    xpath_filter = "/memory-ios-xe-oper:memory-statistics/memory-statistic"
    update_trigger = "periodic"
    period = "2s"

    # Leaves to be marked as keys in Influx LINE format.
    # They are valid throughout all other defined operations.
    keys = ["/memory-ios-xe-oper:memory-statistics/memory-statistic/name"]

  ## IOS-XE Subscription - periodic
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/mdt-oper:mdt-oper-data/mdt-subscriptions"
    update_trigger = "periodic"
	period = "1s"
    keys = ["/mdt-oper:mdt-oper-data/mdt-subscriptions/subscription-id"]

  ## IOS-XE Subscription - Xpath union for multiple subtrees
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
    xpath_filter = "/interfaces-ios-xe-oper:interfaces/interface/statistics/in-octets|/interfaces-ios-xe-oper:interfaces/interface/statistics/out-octets"
    update_trigger = "periodic"
    period = "5s"
    keys = ["/interfaces-ios-xe-oper:interfaces/interface/name"]

  ## IOS-XE Subscription - on-change
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.subscription]]
	xpath_filter = "/cdp-ios-xe-oper:cdp-neighbor-details/cdp-neighbor-detail"
    update_trigger = "on-change"
    period = "0s"

  ### Get operations
  ## IOS-XE Get Request with filter
  [[inputs.cisco_telemetry_mdt_netconf.get_service.get]]
    xpath_filter = "/interfaces-state/interface[name='GigabitEthernet1']/oper-status"
    period = "10s"
    keys = ["/interfaces-state/interface/name"]

  ## IOS-XE Get Request with filter and multiple keys
  [[inputs.cisco_telemetry_mdt_netconf.get_service.get]]
    xpath_filter = "/interfaces-state/interface[name='GigabitEthernet2']"
    period = "10s"
    keys = ["/interfaces-state/interface/name", "/interfaces-state/interface/if-index"]

  ### Event notification subscription
  # NSO Event notfication subscription with a key
  [[inputs.cisco_telemetry_mdt_netconf.subscription_service.notification]]
    stream = "ncs-alarms"
    keys = ["alarm-notification/alarm-class"]
 `

const pluginName = "[inputs.cisco_telemetry_mdt_netconf]"

// Flags that track the state of a service
const (
	connected = iota + 1 // Service is connected to NETCONF server
	created              // Service's requests are created
	ended                // Service has ended
)

// setting represents the internal context of NETCONF client service
type setting struct {
	netconf.Client
	Session
	*sync.Mutex
	// Possible service states: none, connected, created, ended
	state int
}

// Session represents an interface for NETCONF session
type Session interface {
	Call(request interface{}, response interface{}) error
	CallSimple(request interface{}) error
	Receive(response interface{}) error
	Close() error
}

// service represents an interface for NETCONF client services
type service interface {
	createService(context.Context, *CiscoTelemetryNETCONF)
	createRequests(context.Context, *CiscoTelemetryNETCONF) error
}

// dialinSubscriptionRequestsService represents the service for
// dial-in subscriptions established by the NETCONF client
type dialinSubscriptionRequestsService struct {
	Subscriptions []dialinSubscriptionRequest       `toml:"subscription"`
	Notifications []notificationSubscriptionRequest `toml:"notification"`
	*setting
}

// getRequestsService represents the service for get requests
// made by the NETCONF client
type getRequestsService struct {
	Gets []getRequest `toml:"get"`
	*setting
}

// dialinSubscriptionRequest given as input to the plugin
type dialinSubscriptionRequest struct {
	XPathFilter   string `toml:"xpath_filter"`
	UpdateTrigger string `toml:"update_trigger"`
	Period        config.Duration
	Keys          []string
}

// notificationSubscriptionRequest given as input to the plugin
type notificationSubscriptionRequest struct {
	Stream string `toml:"stream"`
	Keys   []string
}

// getRequest given as input to the plugin
type getRequest struct {
	SelectFilter string `toml:"xpath_filter"`
	Period       config.Duration
	Keys         []string
}

// netconfSubscriptionRequest sent through NETCONF, as per specification
type netconfSubscriptionRequest struct {
	XMLName         xml.Name `xml:"urn:ietf:params:xml:ns:yang:ietf-event-notifications establish-subscription"`
	YangPush        string   `xml:"xmlns:yp,attr"`
	Stream          string   `xml:"stream"`
	XPathFilter     string   `xml:"urn:ietf:params:xml:ns:yang:ietf-yang-push xpath-filter"`
	Period          uint64   `xml:"urn:ietf:params:xml:ns:yang:ietf-yang-push period,omitempty"`           // hundreds of a second 100=1s
	DampeningPeriod *uint64  `xml:"urn:ietf:params:xml:ns:yang:ietf-yang-push dampening-period,omitempty"` // on-change
}

// netconfYangPush notification received through NETCONF, as per specification
type netconfYangPush struct {
	netconf.Notification
	PushUpdate struct {
		Content struct {
			InnerXML []byte `xml:",innerxml"`
		} `xml:"datastore-contents-xml"`
	} `xml:"urn:ietf:params:xml:ns:yang:ietf-yang-push push-update"`
}

// netconfYangEvent notification received through NETCONF, as per specification
type netconfYangEvent struct {
	netconf.Notification
	InnerXML []byte `xml:",innerxml"`
}

// CiscoTelemetryNETCONF plugin
type CiscoTelemetryNETCONF struct {
	ServerAddress string `toml:"server_address"`

	Dsrs *dialinSubscriptionRequestsService `toml:"subscription_service"`
	Grs  *getRequestsService                `toml:"get_service"`

	// SSH connection - credentials
	Username string
	Password string

	// SSH connection - security
	IgnoreServerAuthenticity bool
	ServerPublicKey          string

	// Xpaths and keys of all operations
	userXpaths map[string]interface{}
	userKeys   map[string]interface{}

	Redial config.Duration

	acc    telegraf.Accumulator
	cancel context.CancelFunc

	// Waitgroup for monitoring the main services
	waitgroup sync.WaitGroup
}

// createService calls functionalities for the dial-in subscriptions service
func (dsrs *dialinSubscriptionRequestsService) createService(
	ctx context.Context, c *CiscoTelemetryNETCONF) {

	dsrs.setting = &setting{Client: nil, Session: nil,
		Mutex: new(sync.Mutex), state: 0}

	defer c.waitgroup.Done()

	for ctx.Err() == nil {
		c.connectClient(ctx, dsrs.setting)
		if dsrs.setting.state == connected {
			dsrs.createRequests(ctx, c)
		}
		if dsrs.setting.state == created {
			// Launch 2 goroutines: one for reception of telemetry data,
			//  and one for reception of event notifications
			var waitgroup sync.WaitGroup

			if len(dsrs.Subscriptions) > 0 {
				waitgroup.Add(1)
				go func() {
					dsrs.receiveTelemetry(ctx, c)
					waitgroup.Done()
				}()
			}

			if len(dsrs.Notifications) > 0 {
				waitgroup.Add(1)
				go func() {
					dsrs.receiveNotifications(ctx, c)
					waitgroup.Done()
				}()
			}

			waitgroup.Wait()
		}

		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(c.Redial)):
		}
	}
}

// createRequests establishes dynamic dial-in subscriptions over NETCONF
func (dsrs *dialinSubscriptionRequestsService) createRequests(
	ctx context.Context, c *CiscoTelemetryNETCONF) error {
	var err error

	if len(dsrs.Subscriptions) > 0 {
		// Initialization of subscriptions
		subscriptions := make([]*netconfSubscriptionRequest, len(dsrs.Subscriptions))
		for i, subscription := range dsrs.Subscriptions {
			if subscription.Period != 0 {
				if time.Duration(subscription.Period) >= 0 {
					// Check type of subscription
					switch subscription.UpdateTrigger {
					case "periodic":
						if time.Duration(subscription.Period) > 0 {
							subscriptions[i] = &netconfSubscriptionRequest{
								YangPush:    "urn:ietf:params:xml:ns:yang:ietf-yang-push",
								Stream:      "yp:yang-push",
								XPathFilter: subscription.XPathFilter,
								Period: uint64(time.Duration(subscription.Period).Nanoseconds() /
									(int64(time.Millisecond) * 10)),
							}
						} else {
							err = fmt.Errorf(
								"failed to create telemetry subscription: "+
									"time period for subscription %d has to be "+
									"strictly positive but is %d",
								i+1,
								time.Duration(subscription.Period),
							)
							c.acc.AddError(err)
							continue
						}
					case "on-change":
						p := new(uint64)
						*p = uint64(time.Duration(subscription.Period).Nanoseconds() /
							(int64(time.Millisecond) * 10))
						// on-change
						subscriptions[i] = &netconfSubscriptionRequest{
							YangPush:        "urn:ietf:params:xml:ns:yang:ietf-yang-push",
							Stream:          "yp:yang-push",
							XPathFilter:     subscription.XPathFilter,
							DampeningPeriod: p,
						}
					default:
						err = fmt.Errorf(
							"failed to create telemetry subscription: " +
								"bad / missing field update_trigger " +
								"(options are: periodic, on-change)",
						)
						c.acc.AddError(err)
						continue
					}

					// Create dynamic subscription over NETCONF
					log.Printf(
						"%s: establishing subscription %s %s %s key(s)=%v...",
						pluginName, subscription.XPathFilter,
						subscription.UpdateTrigger, time.Duration(subscription.Period),
						subscription.Keys,
					)

					dsrs.Mutex.Lock()
					err = dsrs.Session.CallSimple(subscriptions[i])
					dsrs.Mutex.Unlock()

					if err != nil {
						c.reportError(ctx, err, "create telemetry subscription")
					} else {
						log.Printf(
							"%s: established subscription", pluginName,
						)
						// Announce state:created to the next stage of the pipeline
						dsrs.setting.state = created
					}
				} else {
					err = fmt.Errorf(
						"failed to create telemetry subscription: "+
							"time period for subscription %d has to be positive but is %d",
						i+1,
						time.Duration(subscription.Period),
					)
					c.acc.AddError(err)
				}
			} else {
				err = fmt.Errorf(
					"failed to create telemetry subscription: "+
						"missing field in subscription %d: period",
					i+1,
				)
				c.acc.AddError(err)
			}
		}
	}

	if len(dsrs.Notifications) > 0 {
		// Initialization of notifications
		notifications := make([]*netconf.CreateSubscription, len(dsrs.Notifications))

		// Create notification subscriptions over NETCONF
		for i, notification := range dsrs.Notifications {
			if notification.Stream != "" {
				notifications[i] = &netconf.CreateSubscription{
					Stream: &notification.Stream,
				}
			} else {
				// Allow subscription without any stream, as per specification
				notifications[i] = &netconf.CreateSubscription{}
			}

			log.Printf(
				"%s: establishing notification subscription %s...",
				pluginName, notification,
			)

			dsrs.Mutex.Lock()
			err = dsrs.Session.CallSimple(notifications[i])
			dsrs.Mutex.Unlock()

			if err != nil {
				c.reportError(ctx, err, "create notification subscription")
			} else {
				log.Printf(
					"%s: established subscription", pluginName,
				)
				// Announce state:created to the next stage of the pipeline
				dsrs.setting.state = created
			}
		}
	}

	if len(dsrs.Subscriptions) == 0 && len(dsrs.Notifications) == 0 {
		err = fmt.Errorf("missing subscription requests")
		c.reportError(ctx, err, "create subscription")
	}

	return err
}

// receiveTelemetry for subscriptions as yang-push notifications over NETCONF
func (dsrs *dialinSubscriptionRequestsService) receiveTelemetry(
	ctx context.Context, c *CiscoTelemetryNETCONF) {
	push := &netconfYangPush{}

	// Receive telemetry data
	for ctx.Err() == nil {
		dsrs.Mutex.Lock()
		sessionExists := dsrs.Session != nil
		dsrs.Mutex.Unlock()

		if sessionExists {
			log.Printf(
				"%s: receiving yang-push messages ...", pluginName,
			)

			err := dsrs.Session.Receive(push)

			if err != nil {
				c.reportError(ctx, err, "receive telemetry subscription data")
			} else {
				// Unmarshal XML data
				var tt TelemetryTree

				err = xml.Unmarshal(push.PushUpdate.Content.InnerXML, &tt)
				if err != nil && err != io.EOF {
					// Don't throw an error if error is EOF due to empty xml
					c.acc.AddError(fmt.Errorf(
						"failed to unmarshal XML: %s",
						err,
					))
				} else if err == nil {
					// Send data to Influx accumulator
					c.handleTelemetry(tt, push.EventTime)
				}
			}
		}
	}
	dsrs.state = ended
	log.Printf(
		"%s: stopped Cisco NETCONF dial-in telemetry subscription service on %s",
		pluginName, c.ServerAddress,
	)
}

// receiveNotifications for event subscriptions over NETCONF
func (dsrs *dialinSubscriptionRequestsService) receiveNotifications(
	ctx context.Context, c *CiscoTelemetryNETCONF) {
	event := &netconfYangEvent{}

	// Receive data
	for ctx.Err() == nil {
		dsrs.Mutex.Lock()
		sessionExists := dsrs.Session != nil
		dsrs.Mutex.Unlock()

		if sessionExists {
			log.Printf(
				"%s: receiving event messages ...", pluginName,
			)

			err := dsrs.Session.Receive(event)

			if err != nil {
				c.reportError(ctx, err, "receive notification data")
			} else {
				// Unmarshal XML data
				var tt TelemetryTree
				err = xml.Unmarshal(append(append([]byte("<notification>"),
					event.InnerXML...), []byte("</notification>")...), &tt)

				if err != nil && err != io.EOF {
					// Don't throw an error if error is EOF due to empty xml
					c.acc.AddError(fmt.Errorf(
						"failed to unmarshal XML: %s",
						err,
					))
				} else if err == nil {
					// Take the alarm content only, without the <eventTime> entry (child 0)
					// Send data to Influx accumulator
					c.handleTelemetry(tt.Children[1], event.EventTime)
				}
			}
		}
	}
	dsrs.state = ended
	log.Printf(
		"%s: stopped Cisco NETCONF event notification subscription service on %s",
		pluginName, c.ServerAddress,
	)
}

// createService calls functionalities for the get service
func (grs *getRequestsService) createService(
	ctx context.Context, c *CiscoTelemetryNETCONF) {
	grs.setting = &setting{Client: nil, Session: nil,
		Mutex: new(sync.Mutex), state: 0}

	defer c.waitgroup.Done()

	for ctx.Err() == nil {
		c.connectClient(ctx, grs.setting)
		if grs.setting.state == connected {
			grs.createRequests(ctx, c)
		}

		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(c.Redial)):
		}
	}
}

// createRequests makes get requests over NETCONF, at a given polling interval
func (grs *getRequestsService) createRequests(
	ctx context.Context, c *CiscoTelemetryNETCONF) error {
	var err error
	// Waitgroup for goroutines allocated to each get request
	var waitgroup sync.WaitGroup
	defer waitgroup.Wait()

	if len(grs.Gets) > 0 {

		// Initialization of get operations
		gets := make([]*netconf.Get, len(grs.Gets))

		for i, g := range grs.Gets {
			// Select filter
			filter := netconf.Filter{
				Type:   "xpath",
				Select: g.SelectFilter,
			}

			gets[i] = &netconf.Get{
				Filter: &filter,
			}

			// Add the next goroutine to the waitlist
			waitgroup.Add(1)

			go func(get *netconf.Get, period config.Duration) {
				// Create a ticker that ticks periodically until it is stopped
				ticker := time.NewTicker(time.Duration(
					time.Duration(period).Nanoseconds()) * time.Nanosecond)

				for ctx.Err() == nil {
					grs.Mutex.Lock()
					sessionExists := grs.Session != nil
					grs.Mutex.Unlock()

					if sessionExists {
						var err error
						getReply := new(netconf.RPCReplyData)

						grs.Mutex.Lock()
						sessionExists := grs.Session != nil
						if sessionExists {
							// Get operation
							log.Printf(
								"%s: performing <get> operation ...",
								pluginName,
							)

							err = grs.Session.Call(get, getReply)
							grs.Mutex.Unlock()

							// Get the current timestamp because
							// no timestamp is delivered with NETCONF get
							timestamp := time.Now()

							if err != nil {
								c.reportError(ctx, err, "<get> operation")
							} else {
								// Unmarshal XML data
								var tt TelemetryTree
								err = xml.Unmarshal(getReply.Data.InnerXML, &tt)

								if err != nil {
									c.acc.AddError(fmt.Errorf(
										"failed to unmarshal XML: %s",
										err))
								} else if err == nil {
									// Send data to Influx accumulator
									c.handleTelemetry(tt, timestamp)
								}
							}
						} else {
							grs.Mutex.Unlock()
						}

						// Block waiting on ticker to tick or context to be cancelled
						select {
						case <-ctx.Done():
						case <-ticker.C:
						}
					}
				}

				ticker.Stop()
				waitgroup.Done()

				log.Printf(
					"%s: stopped Cisco NETCONF <get> service on %s",
					pluginName, c.ServerAddress,
				)
			}(gets[i], g.Period)
		}
	} else {
		err = fmt.Errorf("missing get requests")
		c.reportError(ctx, err, "<get> operation")
	}
	grs.state = ended
	return err
}

// connectClient to the NETCONF server for all types of services
func (c *CiscoTelemetryNETCONF) connectClient(ctx context.Context, s *setting) {
	var err error
	var client netconf.Client
	var serverPublicKey ssh.PublicKey

	// Create NETCONF client
	for ctx.Err() == nil {
		// Dial unknown server
		if c.IgnoreServerAuthenticity {
			log.Printf(
				"%s: dialling unknown NETCONF server %s ...",
				pluginName, c.ServerAddress,
			)
			client, err = netconf.DialSSHWithPassword(
				c.ServerAddress,
				c.Username, c.Password,
				ssh.InsecureIgnoreHostKey(),
			)
		} else {
			// Dial known server
			log.Printf(
				"%s: dialling known NETCONF server %s ...",
				pluginName, c.ServerAddress,
			)
			if c.ServerPublicKey != "" {
				if _, _, serverPublicKey, _, _, err =
					ssh.ParseKnownHosts(
						[]byte(c.ServerPublicKey),
					); err == nil {
					client, err = netconf.DialSSHWithPassword(
						c.ServerAddress,
						c.Username, c.Password,
						ssh.FixedHostKey(serverPublicKey),
					)
				} else {
					err = fmt.Errorf(
						"cannot parse public key for server %s - %s",
						c.ServerAddress,
						err,
					)
				}
			} else {
				err = fmt.Errorf(
					"missing public key for server %s",
					c.ServerAddress,
				)
			}
		}

		if err != nil {
			c.reportError(ctx, err, "dial server")
		} else {
			s.Client = client

			log.Printf(
				"%s: dialled server %s",
				pluginName, c.ServerAddress,
			)
			break
		}

		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(c.Redial)):
		}
	}

	// Create NETCONF session
	for ctx.Err() == nil && s.Client != nil {
		log.Printf(
			"%s: creating new NETCONF session ...", pluginName,
		)

		ss, err := s.Client.NewSession()

		if err != nil {
			c.reportError(ctx, err, "create NETCONF session")
		} else {
			log.Printf(
				"%s: created new NETCONF session", pluginName,
			)
			s.Session = ss
			break
		}

		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(c.Redial)):
		}
	}

	if ctx.Err() == nil {
		log.Printf(
			"%s: dialled server %s and created new NETCONF session",
			pluginName, c.ServerAddress,
		)
		s.state = connected
	}
}

// handleTelemetry converts incoming telemetry
// to Influx LINE format and sends it to the accumulator
func (c *CiscoTelemetryNETCONF) handleTelemetry(tt TelemetryTree, t time.Time) {
	// Transform XML tree in Influx LINE format
	grouper, err := tt.TraverseTree(c.userKeys, c.userXpaths,
		strings.Split(c.ServerAddress, ":")[0], t)

	if err != nil {
		c.acc.AddError(fmt.Errorf("%s", err))
	} else {
		// Send lines to accumulator
		for _, m := range grouper.Metrics() {
			c.acc.AddMetric(m)
		}
	}
}

// trimXpath strips an xpath string of "namespace:", leading or trailing '/', and "[*]"
func (c *CiscoTelemetryNETCONF) trimXpath(xpath string) string {
	var newXpath string

	// Remove substrings of the form "namespace:"
	r := regexp.MustCompile(".*:")
	splits := strings.SplitAfter(xpath, "/")
	for _, s := range splits {
		newXpath += r.ReplaceAllString(s, "")
	}

	// Remove leading and trailing '/' characters
	newXpath = strings.Trim(newXpath, "/")

	// Remove substrings of the form "[xyz]"
	r = regexp.MustCompile("\\[.*\\]")
	newXpath = r.ReplaceAllString(newXpath, "")

	return newXpath
}

// preparePaths prepares the user x-paths and the user keys for tree traversals
func (c *CiscoTelemetryNETCONF) preparePaths() {
	var noXpaths int
	var noKeys int

	sExist := c.Dsrs != nil && len(c.Dsrs.Subscriptions) > 0
	nExist := c.Dsrs != nil && len(c.Dsrs.Notifications) > 0
	gExist := c.Grs != nil && len(c.Grs.Gets) > 0

	// Initialize user xpaths and keys
	if sExist {
		for _, s := range c.Dsrs.Subscriptions {
			if s.XPathFilter != "" {
				noXpaths++
			}
			noKeys += len(s.Keys)
		}
	}
	if nExist {
		for _, n := range c.Dsrs.Notifications {
			if n.Stream != "" {
				noXpaths++
			}
			noKeys += len(n.Keys)
		}

	}
	if gExist {
		for _, g := range c.Grs.Gets {
			if g.SelectFilter != "" {
				noXpaths++
			}
			noKeys += len(g.Keys)
		}

	}

	c.userXpaths = make(map[string]interface{}, noXpaths)
	c.userKeys = make(map[string]interface{}, noKeys)

	// Store x-paths and keys from Subscriptions
	if sExist {
		for _, s := range c.Dsrs.Subscriptions {
			if s.XPathFilter != "" {
				c.userXpaths[c.trimXpath(s.XPathFilter)] = nil
				for _, k := range s.Keys {
					c.userKeys[c.trimXpath(k)] = nil
				}
			}
		}
	}
	// Store x-paths and keys from Notifications
	if nExist {
		for _, n := range c.Dsrs.Notifications {
			if n.Stream != "" {
				c.userXpaths[c.trimXpath(n.Stream)] = nil
				for _, k := range n.Keys {
					c.userKeys[c.trimXpath(k)] = nil
				}
			}
		}
	}
	// Store x-paths and keys from GetOperations
	if gExist {
		for _, g := range c.Grs.Gets {
			if g.SelectFilter != "" {
				c.userXpaths[c.trimXpath(g.SelectFilter)] = nil
				for _, k := range g.Keys {
					c.userKeys[c.trimXpath(k)] = nil
				}
			}
		}
	}
}

//reportError abstracts the report of error messages based on their root cause
func (c *CiscoTelemetryNETCONF) reportError(
	ctx context.Context, err error, processName string) {
	logMessage := pluginName + ": NETCONF %s stopped"
	errorMessage := "failed to %s: %s"

	if ctx.Err() != nil && ctx.Err() == context.Canceled {
		// Acknowledge errors due to context cancellation
		log.Printf(
			logMessage, processName,
		)
		log.Println("Context cancelled. Adding error to accumulator: ", err, "!")
		c.acc.AddError(fmt.Errorf(
			errorMessage, processName, err,
		))
	} else if ctx.Err() != nil {
		// Acknowledge other context erro
		log.Println("Context error: Adding error to accumulator: ", err, "!")
		c.acc.AddError(fmt.Errorf(
			errorMessage, processName, ctx.Err().Error(),
		))

	} else {
		// Report base error
		log.Println("Adding error to accumulator: ", err, "!")
		c.acc.AddError(fmt.Errorf(
			errorMessage, processName, err,
		))
	}
}

// cleanup should be called before the termination of the plugin
func (s *setting) cleanup() error {
	// Close open session and client before stopping the plugin
	log.Printf(
		"%s: cleaning up ...", pluginName,
	)
	var err error
	if s.Mutex != nil {
		s.Mutex.Lock()
		if s.Session != nil {
			err = s.Session.Close()
			// Override error if it is due to receive interruption, and hence
			// invalid message framing
			if err == netconf.ErrFraming {
				err = nil
			}
		}
		s.Mutex.Unlock()
	}
	if s.Client != nil {
		if errorClient := s.Client.Close(); err == nil && errorClient != nil {
			err = errorClient
		}
	}
	return err
}

// Start the Cisco NETCONF service
func (c *CiscoTelemetryNETCONF) Start(acc telegraf.Accumulator) error {
	// Plugin context
	var ctx context.Context
	ctx, c.cancel = context.WithCancel(context.Background())

	c.acc = acc

	c.preparePaths()

	// Create services for dial-in subscriptions and/or get requests
	c.waitgroup.Add(1)
	defer c.waitgroup.Done()
	go func() {
		if c.Dsrs != nil && len(c.Dsrs.Subscriptions) > 0 {
			log.Printf(
				"%s: starting Cisco NETCONF dial-in subscription service on %s",
				pluginName, c.ServerAddress)
			c.waitgroup.Add(1)
			go c.Dsrs.createService(ctx, c)
		}
		if c.Grs != nil && len(c.Grs.Gets) > 0 {
			log.Printf(
				"%s: starting Cisco NETCONF get service on %s",
				pluginName, c.ServerAddress,
			)
			c.waitgroup.Add(1)
			go c.Grs.createService(ctx, c)
		}
		if c.Dsrs != nil && len(c.Dsrs.Notifications) > 0 {
			log.Printf(
				"%s: starting Cisco NETCONF notification subscription service on %s",
				pluginName, c.ServerAddress)
			// Create the Dsrs service only if it was not already created
			if len(c.Dsrs.Subscriptions) == 0 {
				c.waitgroup.Add(1)
				go c.Dsrs.createService(ctx, c)
			}
		}
	}()

	return nil
}

// Stop function executed when the plugin is stopped
func (c *CiscoTelemetryNETCONF) Stop() {
	log.Printf(
		"%s: stopping channels ...", pluginName,
	)

	// Send process cancellation to goroutines
	c.cancel()
	// Proceed with cleanup but treat Close() errors as warning
	if c.Dsrs != nil && len(c.Dsrs.Subscriptions) > 0 {
		if err := c.Dsrs.cleanup(); err != nil {
			c.acc.AddError(fmt.Errorf(
				"failed to close NETCONF dial-in subscription session: %s", err))
		}
	}
	if c.Dsrs != nil && len(c.Dsrs.Notifications) > 0 {
		if err := c.Dsrs.cleanup(); err != nil {
			c.acc.AddError(fmt.Errorf(
				"failed to close NETCONF notification subscription session: %s", err))
		}
	}
	if c.Grs != nil && len(c.Grs.Gets) > 0 {
		if err := c.Grs.cleanup(); err != nil {
			c.acc.AddError(fmt.Errorf(
				"failed to close NETCONF <get> session: %s", err))
		}
	}

	// Wait for goroutines to finish their execution
	c.waitgroup.Wait()

	log.Printf(
		"%s: stopped Cisco NETCONF telemetry plugin on %s",
		pluginName, c.ServerAddress,
	)
}

// Description of plugin
func (c *CiscoTelemetryNETCONF) Description() string {
	return "Cisco dial-in NETCONF telemetry input plugin"
}

// SampleConfig of plugin
func (c *CiscoTelemetryNETCONF) SampleConfig() string {
	return sampleConfig
}

// Gather plugin measurements (unused)
func (c *CiscoTelemetryNETCONF) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Initialization function for plugin
func init() {
	inputs.Add("cisco_telemetry_mdt_netconf", func() telegraf.Input {
		return &CiscoTelemetryNETCONF{
			Redial: config.Duration(10 * time.Second),
		}
	})
}
