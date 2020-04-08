// The package implements a minimal set of Azure Event Hubs functionality.
//
// We could use https://github.com/Azure/azure-event-hubs-go but it's too huge
// and also it has a number of dependencies that aren't desired in the project.
package eventhub

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/go-amqp"
)

// Credentials is an evenhub connection string representation.
type Credentials struct {
	Endpoint            string
	SharedAccessKeyName string
	SharedAccessKey     string
	EntityPath          string
}

// ParseConnectionString parses the given connection string into Credentials structure.
func ParseConnectionString(cs string) (*Credentials, error) {
	var c Credentials
	for _, s := range strings.Split(cs, ";") {
		kv := strings.SplitN(s, "=", 2)
		if len(kv) != 2 {
			return nil, errors.New("malformed connection string")
		}

		switch kv[0] {
		case "Endpoint":
			if !strings.HasPrefix(kv[1], "sb://") {
				return nil, errors.New("only sb:// schema supported")
			}
			c.Endpoint = strings.TrimRight(kv[1][5:], "/")
		case "SharedAccessKeyName":
			c.SharedAccessKeyName = kv[1]
		case "SharedAccessKey":
			c.SharedAccessKey = kv[1]
		case "EntityPath":
			c.EntityPath = kv[1]
		}
	}
	return &c, nil
}

// Option is a client configuration option.
type Option func(c *Client)

// WithTLSConfig sets connection TLS configuration.
func WithTLSConfig(tc *tls.Config) Option {
	return WithConnOption(amqp.ConnTLSConfig(tc))
}

// WithSASLPlain configures connection username and password.
func WithSASLPlain(username, password string) Option {
	return WithConnOption(amqp.ConnSASLPlain(username, password))
}

// WithConnOption sets a low-level connection option.
func WithConnOption(opt amqp.ConnOption) Option {
	return func(c *Client) {
		c.opts = append(c.opts, opt)
	}
}

// Dial connects to the named EventHub and returns a client instance.
func Dial(host, name string, opts ...Option) (*Client, error) {
	c := &Client{name: name}
	for _, opt := range opts {
		opt(c)
	}

	var err error
	c.conn, err = amqp.Dial("amqps://"+host, c.opts...)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// DialConnectionString dials an EventHub instance using the given connection string.
func DialConnectionString(cs string, opts ...Option) (*Client, error) {
	creds, err := ParseConnectionString(cs)
	if err != nil {
		return nil, err
	}
	return Dial(creds.Endpoint, creds.EntityPath, append([]Option{
		WithSASLPlain(creds.SharedAccessKeyName, creds.SharedAccessKey),
	}, opts...)...)
}

// Client is an EventHub client.
type Client struct {
	name string
	conn *amqp.Client
	opts []amqp.ConnOption
}

// SubscribeOption is a Subscribe option.
type SubscribeOption func(r *sub)

// WithSubscribeConsumerGroup overrides default consumer group, default is `$Default`.
func WithSubscribeConsumerGroup(name string) SubscribeOption {
	return func(s *sub) {
		s.group = name
	}
}

// WithSubscribeSince requests events that occurred after the given time.
func WithSubscribeSince(t time.Time) SubscribeOption {
	return WithSubscribeLinkOption(amqp.LinkSelectorFilter(
		fmt.Sprintf("amqp.annotation.x-opt-enqueuedtimeutc > '%d'",
			t.UnixNano()/int64(time.Millisecond)),
	))
}

// WithSubscribeLinkOption is a low-level subscription configuration option.
func WithSubscribeLinkOption(opt amqp.LinkOption) SubscribeOption {
	return func(s *sub) {
		s.opts = append(s.opts, opt)
	}
}

type sub struct {
	group string
	opts  []amqp.LinkOption
}

// Event is an Event Hub event, simply wraps an AMQP message.
type Event struct {
	*amqp.Message
}

// Subscribe subscribes to all hub's partitions and registers the given
// handler and blocks until it encounters an error or the context is cancelled.
//
// It's client's responsibility to accept/reject/release events.
func (c *Client) Subscribe(
	ctx context.Context,
	fn func(event *Event) error,
	opts ...SubscribeOption,
) error {
	var s sub
	for _, opt := range opts {
		opt(&s)
	}
	if s.group == "" {
		s.group = "$Default"
	}

	// initialize new session for each subscribe session
	sess, err := c.conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close(context.Background())

	ids, err := c.getPartitionIDs(ctx, sess)
	if err != nil {
		return err
	}

	// stop all goroutines at return
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msgc := make(chan *amqp.Message)
	errc := make(chan error)

	for _, id := range ids {
		addr := fmt.Sprintf("/%s/ConsumerGroups/%s/Partitions/%s", c.name, s.group, id)
		recv, err := sess.NewReceiver(
			append([]amqp.LinkOption{amqp.LinkSourceAddress(addr)}, s.opts...)...,
		)
		if err != nil {
			return err
		}

		go func(recv *amqp.Receiver) {
			defer recv.Close(context.Background())
			for {
				msg, err := recv.Receive(ctx)
				if err != nil {
					select {
					case errc <- err:
					case <-ctx.Done():
					}
					return
				}
				select {
				case msgc <- msg:
				case <-ctx.Done():
				}
			}
		}(recv)
	}

	for {
		select {
		case msg := <-msgc:
			if err := fn(&Event{msg}); err != nil {
				return err
			}
		case err := <-errc:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// getPartitionIDs returns partition ids of the hub.
func (c *Client) getPartitionIDs(ctx context.Context, sess *amqp.Session) ([]string, error) {
	replyTo := genID()
	recv, err := sess.NewReceiver(
		amqp.LinkSourceAddress("$management"),
		amqp.LinkTargetAddress(replyTo),
	)
	if err != nil {
		return nil, err
	}
	defer recv.Close(context.Background())

	send, err := sess.NewSender(
		amqp.LinkTargetAddress("$management"),
		amqp.LinkSourceAddress(replyTo),
	)
	if err != nil {
		return nil, err
	}
	defer send.Close(context.Background())

	mid := genID()
	if err := send.Send(ctx, &amqp.Message{
		Properties: &amqp.MessageProperties{
			MessageID: mid,
			ReplyTo:   replyTo,
		},
		ApplicationProperties: map[string]interface{}{
			"operation": "READ",
			"name":      c.name,
			"type":      "com.microsoft:eventhub",
		},
	}); err != nil {
		return nil, err
	}

	msg, err := recv.Receive(ctx)
	if err != nil {
		return nil, err
	}
	if err = CheckMessageResponse(msg); err != nil {
		return nil, err
	}
	if msg.Properties.CorrelationID != mid {
		return nil, errors.New("message-id mismatch")
	}
	if err := msg.Accept(); err != nil {
		return nil, err
	}

	val, ok := msg.Value.(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to typecast value")
	}
	ids, ok := val["partition_ids"].([]string)
	if !ok {
		return nil, errors.New("unable to typecast partition_ids")
	}
	return ids, nil
}

// Close closes underlying AMQP connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// CheckMessageResponse checks for 200 response code otherwise returns an error.
func CheckMessageResponse(msg *amqp.Message) error {
	rc, ok := msg.ApplicationProperties["status-code"].(int32)
	if !ok {
		return errors.New("unable to typecast status-code")
	}
	if rc == 200 {
		return nil
	}
	rd, _ := msg.ApplicationProperties["status-description"].(string)
	return fmt.Errorf("code = %d, description = %q", rc, rd)
}

func genID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
