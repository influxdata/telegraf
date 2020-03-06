package iotservice

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/amenzhinsky/iothub/common"
	"github.com/amenzhinsky/iothub/eventhub"
	"github.com/amenzhinsky/iothub/logger"
)

// ClientOption is a client configuration option.
type ClientOption func(c *Client)

// WithHTTPClient changes default http rest client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.http = client
	}
}

// WithLogger sets client logger.
func WithLogger(l logger.Logger) ClientOption {
	return func(c *Client) {
		c.logger = l
	}
}

// WithTLSConfig sets TLS config that's used by REST HTTP and AMQP clients.
func WithTLSConfig(config *tls.Config) ClientOption {
	return func(c *Client) {
		c.tls = config
	}
}

const userAgent = "iothub-golang-sdk/dev"

func ParseConnectionString(cs string) (*common.SharedAccessKey, error) {
	m, err := common.ParseConnectionString(
		cs, "HostName", "SharedAccessKeyName", "SharedAccessKey",
	)
	if err != nil {
		return nil, err
	}
	return common.NewSharedAccessKey(
		m["HostName"], m["SharedAccessKeyName"], m["SharedAccessKey"],
	), nil
}

func NewFromConnectionString(cs string, opts ...ClientOption) (*Client, error) {
	sak, err := ParseConnectionString(cs)
	if err != nil {
		return nil, err
	}
	return New(sak, opts...)
}

// New creates new iothub service client.
func New(sak *common.SharedAccessKey, opts ...ClientOption) (*Client, error) {
	c := &Client{
		sak:    sak,
		done:   make(chan struct{}),
		logger: logger.New(logger.LevelWarn, nil),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.tls == nil {
		c.tls = &tls.Config{RootCAs: common.RootCAs()}
	}
	if c.http == nil {
		c.http = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: common.RootCAs(),
				},
			},
		}
	}
	return c, nil
}

// Client is IoT Hub service client.
type Client struct {
	mu     sync.Mutex
	tls    *tls.Config
	conn   *amqp.Client
	done   chan struct{}
	sak    *common.SharedAccessKey
	logger logger.Logger
	http   *http.Client // REST client

	sendMu   sync.Mutex
	sendLink *amqp.Sender

	// TODO: figure out if it makes sense to cache feedback and file notification receivers
}

// newSession connects to IoT Hub's AMQP broker,
// it's needed for sending C2S events and subscribing to events feedback.
//
// It establishes connection only once, subsequent calls return immediately.
func (c *Client) newSession(ctx context.Context) (*amqp.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.NewSession() // already connected
	}
	conn, err := amqp.Dial("amqps://"+c.sak.HostName,
		amqp.ConnTLSConfig(c.tls),
		amqp.ConnProperty("com.microsoft:client-version", userAgent),
	)
	if err != nil {
		return nil, err
	}

	c.logger.Debugf("connected to %s", c.sak.HostName)
	if err = c.putTokenContinuously(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, err
	}

	sess, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	c.conn = conn
	return sess, nil
}

// putTokenContinuously writes token first time in blocking mode and returns
// maintaining token updates in the background until the client is closed.
func (c *Client) putTokenContinuously(ctx context.Context, conn *amqp.Client) error {
	const (
		tokenUpdateInterval = time.Hour

		// we need to update tokens before they expire to prevent disconnects
		// from azure, without interrupting the message flow
		tokenUpdateSpan = 10 * time.Minute
	)

	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close(context.Background())

	if err := c.putToken(ctx, sess, tokenUpdateInterval); err != nil {
		return err
	}

	go func() {
		ticker := time.NewTimer(tokenUpdateInterval - tokenUpdateSpan)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.putToken(context.Background(), sess, tokenUpdateInterval); err != nil {
					c.logger.Errorf("put token error: %s", err)
					return
				}
				ticker.Reset(tokenUpdateInterval - tokenUpdateSpan)
				c.logger.Debugf("token updated")
			case <-c.done:
				return
			}
		}
	}()
	return nil
}

func (c *Client) putToken(
	ctx context.Context, sess *amqp.Session, lifetime time.Duration,
) error {
	send, err := sess.NewSender(
		amqp.LinkTargetAddress("$cbs"),
	)
	if err != nil {
		return err
	}
	defer send.Close(context.Background())

	recv, err := sess.NewReceiver(
		amqp.LinkSourceAddress("$cbs"),
	)
	if err != nil {
		return err
	}
	defer recv.Close(context.Background())

	sas, err := c.sak.Token(c.sak.HostName, lifetime)
	if err != nil {
		return err
	}
	if err = send.Send(ctx, &amqp.Message{
		Value: sas.String(),
		Properties: &amqp.MessageProperties{
			To:      "$cbs",
			ReplyTo: "cbs",
		},
		ApplicationProperties: map[string]interface{}{
			"operation": "put-token",
			"type":      "servicebus.windows.net:sastoken",
			"name":      c.sak.HostName,
		},
	}); err != nil {
		return err
	}

	msg, err := recv.Receive(ctx)
	if err != nil {
		return err
	}
	if err = msg.Accept(); err != nil {
		return err
	}
	return eventhub.CheckMessageResponse(msg)
}

// connectToEventHub connects to IoT Hub endpoint compatible with Eventhub
// for receiving D2C events, it uses different endpoints and authentication
// mechanisms than newSession.
func (c *Client) connectToEventHub(ctx context.Context) (*eventhub.Client, error) {
	sess, err := c.newSession(ctx)
	if err != nil {
		return nil, err
	}

	// iothub broker should redirect us to an eventhub compatible instance
	// straight after subscribing to events stream, for that we need to connect twice
	defer sess.Close(context.Background())

	_, err = sess.NewReceiver(
		amqp.LinkSourceAddress("messages/events/"),
	)
	if err == nil {
		return nil, errorf("expected redirect error")
	}
	rerr, ok := err.(*amqp.Error)
	if !ok || rerr.Condition != amqp.ErrorLinkRedirect {
		return nil, err
	}

	// "amqps://{host}:5671/{consumerGroup}/"
	group := rerr.Info["address"].(string)
	group = group[strings.Index(group, ":5671/")+6 : len(group)-1]

	host := rerr.Info["hostname"].(string)
	c.logger.Debugf("redirected to %s:%s eventhub", host, group)

	tlsCfg := c.tls.Clone()
	tlsCfg.ServerName = host

	eh, err := eventhub.Dial(host, group,
		eventhub.WithTLSConfig(tlsCfg),
		eventhub.WithSASLPlain(c.sak.SharedAccessKeyName, c.sak.SharedAccessKey),
		eventhub.WithConnOption(amqp.ConnProperty("com.microsoft:client-version", userAgent)),
	)
	if err != nil {
		return nil, err
	}
	c.logger.Debugf("connected to %s:%s eventhub", host, group)
	return eh, nil
}

// EventHandler handles incoming cloud-to-device events.
type EventHandler func(e *Event) error

// Event is a device-to-cloud message.
type Event struct {
	*common.Message
}

// SubscribeEvents subscribes to D2C events.
//
// Event handler is blocking, handle asynchronous processing on your own.
func (c *Client) SubscribeEvents(ctx context.Context, fn EventHandler) error {
	// a new connection is established for every invocation,
	// this made on purpose because normally an app calls the method once
	eh, err := c.connectToEventHub(ctx)
	if err != nil {
		return err
	}
	defer eh.Close()

	return eh.Subscribe(ctx, func(msg *eventhub.Event) error {
		if err := fn(&Event{FromAMQPMessage(msg.Message)}); err != nil {
			return err
		}
		return msg.Accept()
	},
		eventhub.WithSubscribeSince(time.Now()),
	)
}

// SendOption is a send option.
type SendOption func(msg *common.Message) error

// WithSendMessageID sets message id.
func WithSendMessageID(mid string) SendOption {
	return func(msg *common.Message) error {
		msg.MessageID = mid
		return nil
	}
}

// WithSendCorrelationID sets correlation id.
func WithSendCorrelationID(cid string) SendOption {
	return func(msg *common.Message) error {
		msg.CorrelationID = cid
		return nil
	}
}

// WithSendUserID sets user id.
func WithSendUserID(uid string) SendOption {
	return func(msg *common.Message) error {
		msg.UserID = uid
		return nil
	}
}

// AckType is event feedback acknowledgement type.
type AckType string

const (
	// AckNone no feedback.
	AckNone AckType = "none"

	// AckPositive receive a feedback message if the message was completed.
	AckPositive AckType = "positive"

	// AckNegative receive a feedback message if the message expired
	// (or maximum delivery count was reached) without being completed by the device.
	AckNegative AckType = "negative"

	// AckFull both positive and negative.
	AckFull AckType = "full"
)

// WithSendAck sets message confirmation type.
func WithSendAck(ack AckType) SendOption {
	return func(msg *common.Message) error {
		if ack == "" {
			return nil
		}
		return WithSendProperty("iothub-ack", string(ack))(msg)
	}
}

// WithSendExpiryTime sets message expiration time.
func WithSendExpiryTime(t time.Time) SendOption {
	return func(msg *common.Message) error {
		msg.ExpiryTime = &t
		return nil
	}
}

// WithSendProperty sets a message property.
func WithSendProperty(k, v string) SendOption {
	return func(msg *common.Message) error {
		if msg.Properties == nil {
			msg.Properties = map[string]string{}
		}
		msg.Properties[k] = v
		return nil
	}
}

// WithSendProperties same as `WithSendProperty` but accepts map of keys and values.
func WithSendProperties(m map[string]string) SendOption {
	return func(msg *common.Message) error {
		if msg.Properties == nil {
			msg.Properties = map[string]string{}
		}
		for k, v := range m {
			msg.Properties[k] = v
		}
		return nil
	}
}

// SendEvent sends the given cloud-to-device message and returns its id.
// Panics when event is nil.
func (c *Client) SendEvent(
	ctx context.Context,
	deviceID string,
	payload []byte,
	opts ...SendOption,
) error {
	if deviceID == "" {
		return errorf("device id is empty")
	}
	msg := &common.Message{
		To:      "/devices/" + deviceID + "/messages/devicebound",
		Payload: payload,
	}
	for _, opt := range opts {
		if err := opt(msg); err != nil {
			return err
		}
	}

	send, err := c.getSendLink(ctx)
	if err != nil {
		return err
	}
	return send.Send(ctx, toAMQPMessage(msg))
}

// getSendLink caches sender link between calls to speed up sending events.
func (c *Client) getSendLink(ctx context.Context) (*amqp.Sender, error) {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.sendLink != nil {
		return c.sendLink, nil
	}
	sess, err := c.newSession(ctx)
	if err != nil {
		return nil, err
	}
	// since the link is cached it's supposed to be closed along with the client itself

	c.sendLink, err = sess.NewSender(
		amqp.LinkTargetAddress("/messages/devicebound"),
	)
	if err != nil {
		_ = sess.Close(context.Background())
		return nil, err
	}
	return c.sendLink, nil
}

// FeedbackHandler handles message feedback.
type FeedbackHandler func(f *Feedback) error

// SubscribeFeedback subscribes to feedback of messages that ack was requested.
func (c *Client) SubscribeFeedback(ctx context.Context, fn FeedbackHandler) error {
	sess, err := c.newSession(ctx)
	if err != nil {
		return err
	}
	defer sess.Close(context.Background())

	recv, err := sess.NewReceiver(
		amqp.LinkSourceAddress("/messages/serviceBound/feedback"),
	)
	if err != nil {
		return err
	}
	defer recv.Close(context.Background())

	for {
		msg, err := recv.Receive(ctx)
		if err != nil {
			return err
		}
		if len(msg.Data) == 0 {
			c.logger.Warnf("zero length data received")
			continue
		}

		var v []*Feedback
		c.logger.Debugf("feedback received: %s", msg.GetData())
		if err = json.Unmarshal(msg.GetData(), &v); err != nil {
			return err
		}
		for _, f := range v {
			if err := fn(f); err != nil {
				return err
			}
		}
		if err = msg.Accept(); err != nil {
			return err
		}
	}
}

// Feedback is message feedback.
type Feedback struct {
	OriginalMessageID  string    `json:"originalMessageId"`
	Description        string    `json:"description"`
	DeviceGenerationID string    `json:"deviceGenerationId"`
	DeviceID           string    `json:"deviceId"`
	EnqueuedTimeUTC    time.Time `json:"enqueuedTimeUtc"`
	StatusCode         string    `json:"statusCode"`
}

// FileNotification is emitted once a blob file is uploaded to the hub.
//
// TODO: structure is yet to define.
type FileNotification struct {
	*amqp.Message
}

// FileNotificationHandler handles file upload notifications.
type FileNotificationHandler func(event *FileNotification) error

// SubscribeFileNotifications subscribes to file notifications.
//
// The feature has to be enabled in the console.
func (c *Client) SubscribeFileNotifications(
	ctx context.Context,
	fn FileNotificationHandler,
) error {
	sess, err := c.newSession(ctx)
	if err != nil {
		return err
	}
	defer sess.Close(context.Background())

	recv, err := sess.NewReceiver(
		amqp.LinkSourceAddress("/messages/serviceBound/filenotifications"),
	)
	if err != nil {
		return err
	}
	defer recv.Close(context.Background())

	for {
		msg, err := recv.Receive(ctx)
		if err != nil {
			return err
		}
		if len(msg.Data) == 0 {
			c.logger.Warnf("zero length data received")
			continue
		}
		if err := fn(&FileNotification{msg}); err != nil {
			return err
		}
		if err = msg.Accept(); err != nil {
			return err
		}
	}
}

// HostName returns service's hostname.
func (c *Client) HostName() string {
	return c.sak.HostName
}

// DeviceConnectionString builds up a connection string for the given device.
func (c *Client) DeviceConnectionString(device *Device, secondary bool) (string, error) {
	key, err := accessKey(device.Authentication, secondary)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("HostName=%s;DeviceId=%s;SharedAccessKey=%s",
		c.sak.HostName, device.DeviceID, key,
	), nil
}

func (c *Client) ModuleConnectionString(module *Module, secondary bool) (string, error) {
	key, err := accessKey(module.Authentication, secondary)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("HostName=%s;DeviceId=%s;ModuleId=%s;SharedAccessKey=%s",
		c.sak.HostName, module.DeviceID, module.ModuleID, key,
	), nil
}

// DeviceSAS generates a GenerateToken token for the named device.
//
// Resource shouldn't include hostname.
func (c *Client) DeviceSAS(
	device *Device, resource string, duration time.Duration, secondary bool,
) (string, error) {
	key, err := accessKey(device.Authentication, secondary)
	if err != nil {
		return "", err
	}
	sas, err := common.NewSharedAccessSignature(
		c.sak.HostName+"/"+strings.TrimLeft(resource, "/"),
		"",
		key,
		time.Now().Add(duration),
	)
	if err != nil {
		return "", err
	}
	return sas.String(), nil
}

func accessKey(auth *Authentication, secondary bool) (string, error) {
	if auth.Type != AuthSAS {
		return "", errorf("invalid authentication type: %s", auth.Type)
	}
	if secondary {
		return auth.SymmetricKey.SecondaryKey, nil
	}
	return auth.SymmetricKey.PrimaryKey, nil
}

func (c *Client) CallDeviceMethod(
	ctx context.Context,
	deviceID string,
	call *MethodCall,
) (*MethodResult, error) {
	return c.callMethod(
		ctx,
		pathf("twins/%s/methods", deviceID),
		call,
	)
}

func (c *Client) CallModuleMethod(
	ctx context.Context,
	deviceID,
	moduleID string,
	call *MethodCall,
) (*MethodResult, error) {
	return c.callMethod(
		ctx,
		pathf("twins/%s/modules/%s/methods", deviceID, moduleID),
		call,
	)
}

func (c *Client) callMethod(ctx context.Context, path string, call *MethodCall) (
	*MethodResult, error,
) {
	var res MethodResult
	if _, err := c.call(
		ctx,
		http.MethodPost,
		path,
		nil,
		nil,
		call,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetDevice retrieves the named device.
func (c *Client) GetDevice(ctx context.Context, deviceID string) (*Device, error) {
	var res Device
	if _, err := c.call(
		ctx,
		http.MethodGet,
		pathf("devices/%s", deviceID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// CreateDevice creates a new device.
func (c *Client) CreateDevice(ctx context.Context, device *Device) (*Device, error) {
	var res Device
	if _, err := c.call(
		ctx,
		http.MethodPut,
		pathf("devices/%s", device.DeviceID),
		nil,
		nil,
		device,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// CreateDevices creates array of devices in bulk mode.
func (c *Client) CreateDevices(
	ctx context.Context, devices []*Device,
) (*BulkResult, error) {
	return c.bulkRequest(ctx, devices, "create")
}

// UpdateDevices updates array of devices in bulk mode.
func (c *Client) UpdateDevices(
	ctx context.Context, devices []*Device, force bool,
) (*BulkResult, error) {
	op := "UpdateIfMatchETag"
	if force {
		op = "Update"
	}
	return c.bulkRequest(ctx, devices, op)
}

// DeleteDevices deletes array of devices in bulk mode.
func (c *Client) DeleteDevices(
	ctx context.Context, devices []*Device, force bool,
) (*BulkResult, error) {
	op := "DeleteIfMatchETag"
	if force {
		op = "Delete"
	}
	return c.bulkRequest(ctx, devices, op)
}

func (c *Client) bulkRequest(
	ctx context.Context, devices []*Device, op string,
) (*BulkResult, error) {
	// convert devices into a variable map and rename deviceId to id
	devs := make([]map[string]interface{}, 0, len(devices))
	for _, dev := range devices {
		m, err := toMap(dev)
		if err != nil {
			return nil, err
		}
		id := m["deviceId"]
		delete(m, "deviceId")
		m["id"] = id
		m["importMode"] = op
		devs = append(devs, m)
	}

	var res BulkResult
	_, err := c.call(
		ctx,
		http.MethodPost,
		"devices",
		nil,
		nil,
		devs,
		&res,
	)
	if err != nil {
		if re, ok := err.(*RequestError); ok && re.Res.StatusCode == http.StatusBadRequest {
			if err = json.Unmarshal(re.Body, &res); err != nil {
				return nil, err
			}
			return &res, nil
		}
		return nil, err
	}
	return &res, err
}

// ridiculous way to convert a structure to a variable map
func toMap(v interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func ifMatchHeader(etag string) http.Header {
	if etag == "" {
		etag = "*"
	} else {
		etag = `"` + etag + `"`
	}
	return http.Header{"If-Match": {etag}}
}

// UpdateDevice updates the named device.
func (c *Client) UpdateDevice(ctx context.Context, device *Device) (*Device, error) {
	var res Device
	if _, err := c.call(
		ctx,
		http.MethodPut,
		pathf("devices/%s", device.DeviceID),
		nil,
		ifMatchHeader(device.ETag),
		device,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// DeleteDevice deletes the named device.
func (c *Client) DeleteDevice(ctx context.Context, device *Device) error {
	_, err := c.call(
		ctx,
		http.MethodDelete,
		pathf("devices/%s", device.DeviceID),
		nil,
		ifMatchHeader(device.ETag),
		nil,
		nil,
	)
	return err
}

// ListDevices lists all registered devices.
func (c *Client) ListDevices(ctx context.Context) ([]*Device, error) {
	var res []*Device
	if _, err := c.call(
		ctx,
		http.MethodGet,
		"devices",
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

// ListModules list all the registered modules on the named device.
func (c *Client) ListModules(ctx context.Context, deviceID string) ([]*Module, error) {
	var res []*Module
	if _, err := c.call(
		ctx,
		http.MethodGet,
		pathf("devices/%s/modules", deviceID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

// CreateModule adds the given module to the registry.
func (c *Client) CreateModule(ctx context.Context, module *Module) (*Module, error) {
	var res Module
	if _, err := c.call(ctx,
		http.MethodPut,
		pathf("devices/%s/modules/%s", module.DeviceID, module.ModuleID),
		nil,
		nil,
		module,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetModule retrieves the named module.
func (c *Client) GetModule(ctx context.Context, deviceID, moduleID string) (
	*Module, error,
) {
	var res Module
	if _, err := c.call(
		ctx,
		http.MethodGet,
		pathf("devices/%s/modules/%s", deviceID, moduleID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// UpdateModule updates the given module.
func (c *Client) UpdateModule(ctx context.Context, module *Module) (*Module, error) {
	var res Module
	if _, err := c.call(
		ctx,
		http.MethodPut,
		pathf("devices/%s/modules/%s", module.DeviceID, module.ModuleID),
		nil,
		ifMatchHeader(module.ETag),
		module,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// DeleteModule removes the named device module.
func (c *Client) DeleteModule(ctx context.Context, module *Module) error {
	_, err := c.call(
		ctx,
		http.MethodDelete,
		pathf("devices/%s/modules/%s", module.DeviceID, module.ModuleID),
		nil,
		ifMatchHeader(module.ETag),
		nil,
		nil,
	)
	return err
}

// GetDeviceTwin retrieves the named twin device from the registry.
func (c *Client) GetDeviceTwin(ctx context.Context, deviceID string) (*Twin, error) {
	var res Twin
	if _, err := c.call(
		ctx,
		http.MethodGet,
		pathf("twins/%s", deviceID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetModuleTwin retrieves the named module's path.
func (c *Client) GetModuleTwin(ctx context.Context, deviceID, moduleID string) (
	*ModuleTwin, error,
) {
	var res ModuleTwin
	if _, err := c.call(
		ctx,
		http.MethodGet,
		pathf("twins/%s/modules/%s", deviceID, moduleID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// UpdateDeviceTwin updates the named twin desired properties.
func (c *Client) UpdateDeviceTwin(ctx context.Context, twin *Twin) (*Twin, error) {
	var res Twin
	if _, err := c.call(
		ctx,
		http.MethodPatch,
		pathf("twins/%s", twin.DeviceID),
		nil,
		ifMatchHeader(twin.ETag),
		twin,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// UpdateModuleTwin updates the named module twin's desired attributes.
func (c *Client) UpdateModuleTwin(ctx context.Context, twin *ModuleTwin) (
	*ModuleTwin, error,
) {
	var res ModuleTwin
	if _, err := c.call(
		ctx,
		http.MethodPatch,
		pathf("twins/%s/modules/%s", twin.DeviceID, twin.ModuleID),
		nil,
		ifMatchHeader(twin.ETag),
		twin,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// ListConfigurations gets all available configurations from the registry.
func (c *Client) ListConfigurations(ctx context.Context) ([]*Configuration, error) {
	var res []*Configuration
	if _, err := c.call(
		ctx,
		http.MethodGet,
		"configurations",
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

// CreateConfiguration adds the given configuration to the registry.
func (c *Client) CreateConfiguration(ctx context.Context, config *Configuration) (
	*Configuration, error,
) {
	var res Configuration
	if _, err := c.call(
		ctx,
		http.MethodPut,
		pathf("configurations/%s", config.ID),
		nil,
		nil,
		config,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetConfiguration gets the named configuration from the registry.
func (c *Client) GetConfiguration(ctx context.Context, configID string) (
	*Configuration, error,
) {
	var res Configuration
	if _, err := c.call(
		ctx,
		http.MethodGet,
		pathf("configurations/%s", configID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// UpdateConfiguration updates the given configuration in the registry.
func (c *Client) UpdateConfiguration(ctx context.Context, config *Configuration) (
	*Configuration, error,
) {
	var res Configuration
	if _, err := c.call(
		ctx,
		http.MethodPut,
		pathf("configurations/%s", config.ID),
		nil,
		ifMatchHeader(config.ETag),
		config,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// DeleteConfiguration removes the given configuration from the registry.
func (c *Client) DeleteConfiguration(ctx context.Context, config *Configuration) error {
	_, err := c.call(
		ctx,
		http.MethodDelete,
		pathf("configurations/%s", config.ID),
		nil,
		ifMatchHeader(config.ETag),
		nil,
		nil,
	)
	return err
}

func (c *Client) ApplyConfigurationContentOnDevice(
	ctx context.Context,
	deviceID string,
	content *ConfigurationContent,
) error {
	_, err := c.call(
		ctx,
		http.MethodPost,
		pathf("devices/%s/applyConfigurationContent", deviceID),
		nil,
		nil,
		content,
		nil,
	)
	return err
}

func (c *Client) QueryDevices(
	ctx context.Context, query string, fn func(v map[string]interface{}) error,
) error {
	var res []map[string]interface{}
	return c.query(
		ctx,
		http.MethodPost,
		"devices/query",
		nil,
		0, // TODO: control page size
		map[string]string{
			"Query": query,
		},
		&res,
		func() error {
			for _, v := range res {
				if err := fn(v); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

func (c *Client) query(
	ctx context.Context,
	method string,
	path string,
	vals url.Values,
	pageSize uint,
	req interface{},
	res interface{},
	fn func() error,
) error {
	var token string
QueryNext:
	h := http.Header{}
	if token != "" {
		h.Add("x-ms-continuation", token)
	}
	if pageSize > 0 {
		h.Add("x-ms-max-item-count", fmt.Sprintf("%d", pageSize))
	}
	header, err := c.call(
		ctx,
		method,
		path,
		vals,
		h,
		req,
		&res,
	)
	if err != nil {
		return err
	}
	if err = fn(); err != nil {
		return err
	}
	if token = header.Get("x-ms-continuation"); token != "" {
		goto QueryNext
	}
	return nil
}

// Stats retrieves the device registry statistic.
func (c *Client) Stats(ctx context.Context) (*Stats, error) {
	var res Stats
	if _, err := c.call(
		ctx,
		http.MethodGet,
		"statistics/devices",
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

// CreateJob creates import / export jobs.
//
// https://docs.microsoft.com/en-us/azure/iot-hub/iot-hub-bulk-identity-mgmt#get-the-container-sas-uri
func (c *Client) CreateJob(ctx context.Context, job *Job) (map[string]interface{}, error) {
	var res map[string]interface{}
	if _, err := c.call(
		ctx,
		http.MethodPost,
		"jobs/create",
		nil,
		nil,
		job,
		&res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

// ListJobs lists all running jobs.
func (c *Client) ListJobs(ctx context.Context) ([]map[string]interface{}, error) {
	var res []map[string]interface{}
	if _, err := c.call(
		ctx,
		http.MethodGet,
		"jobs",
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) GetJob(ctx context.Context, jobID string) (map[string]interface{}, error) {
	var res map[string]interface{}
	if _, err := c.call(
		ctx,
		http.MethodGet,
		pathf("jobs/%s", jobID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) CancelJob(ctx context.Context, jobID string) (map[string]interface{}, error) {
	var res map[string]interface{}
	if _, err := c.call(
		ctx,
		http.MethodDelete,
		pathf("jobs/%s", jobID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return res, nil
}

type JobV2Query struct {
	Type     JobV2Type
	Status   JobV2Status
	PageSize uint
}

func (c *Client) QueryJobsV2(
	ctx context.Context, q *JobV2Query, fn func(*JobV2) error,
) error {
	vals := url.Values{}
	if q.Type != "" {
		vals.Add("jobType", string(q.Type))
	}
	if q.Status != "" {
		vals.Add("jobStatus", string(q.Status))
	}
	var res []*JobV2
	return c.query(
		ctx,
		http.MethodGet,
		"jobs/v2/query",
		vals,
		q.PageSize,
		nil,
		&res,
		func() error {
			for _, v := range res {
				if err := fn(v); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

func (c *Client) GetJobV2(ctx context.Context, jobID string) (*JobV2, error) {
	var res JobV2
	if _, err := c.call(
		ctx,
		http.MethodGet,
		pathf("jobs/v2/%s", jobID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) CancelJobV2(ctx context.Context, jobID string) (*JobV2, error) {
	var res JobV2
	if _, err := c.call(
		ctx,
		http.MethodPost,
		pathf("jobs/v2/%s/cancel", jobID),
		nil,
		nil,
		nil,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) CreateJobV2(ctx context.Context, job *JobV2) (*JobV2, error) {
	var res JobV2
	_, err := c.call(
		ctx,
		http.MethodPut,
		pathf("jobs/v2/%s", job.JobID),
		nil,
		nil,
		job,
		&res,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) call(
	ctx context.Context,
	method string,
	path string,
	vals url.Values,
	headers http.Header,
	r, v interface{}, // request and response objects
) (http.Header, error) {
	var br io.Reader
	if r != nil {
		b, err := json.Marshal(r)
		if err != nil {
			return nil, err
		}
		br = bytes.NewReader(b)
	}
	q := url.Values{"api-version": []string{"2019-03-30"}}
	for k, vv := range vals {
		for _, v := range vv {
			q.Add(k, v)
		}
	}

	uri := "https://" + c.sak.HostName + "/" + path + "?" + q.Encode()
	req, err := http.NewRequest(method, uri, br)
	if err != nil {
		return nil, err
	}
	sas, err := c.sak.Token(c.sak.HostName, time.Hour)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", sas.String())
	req.Header.Set("Request-Id", genID())
	req.Header.Set("User-Agent", userAgent)
	for k, v := range headers {
		for i := range v {
			req.Header.Add(k, v[i])
		}
	}

	c.logger.Debugf("%s", (*requestOutDump)(req))
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	c.logger.Debugf("%s", (*responseDump)(res))

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case http.StatusNoContent:
		return res.Header, nil
	case http.StatusOK:
		return res.Header, json.Unmarshal(body, v)
	case http.StatusBadRequest:
		// try to decode a registry error, because some operations like
		// bulk requests may return the bad request code along with a valid body
		var e BadRequestError
		if err = json.Unmarshal(body, &e); err == nil && e.Message != "" {
			return nil, &e
		}
	}
	return nil, &RequestError{Res: res, Body: body}
}

// RequestError is an API request error.
//
// Response body is already read out to Body attribute,
// so there's no need read it manually and call `e.Res.Body.Close()`
type RequestError struct {
	Res  *http.Response
	Body []byte
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("code = %d, body = %q", e.Res.StatusCode, e.Body)
}

func genID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// Close closes transport.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.done:
		return nil
	default:
		close(c.done)
	}
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func pathf(format string, s ...string) string {
	v := make([]interface{}, len(s))
	for i := range s {
		v[i] = url.PathEscape(s[i])
	}
	return fmt.Sprintf(format, v...)
}

func errorf(format string, v ...interface{}) error {
	return fmt.Errorf("iotservice: "+format, v...)
}
