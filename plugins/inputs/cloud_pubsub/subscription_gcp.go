package cloud_pubsub

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub/v2"
)

type (
	subscription interface {
		// ID returns the unique identifier of the subscription.
		ID() string
		// Receive starts receiving messages from the subscription and processes them using the provided function.
		Receive(ctx context.Context, f func(context.Context, message)) error
	}

	message interface {
		// Ack acknowledges the message, indicating successful processing.
		Ack()
		// Nack negatively acknowledges the message, indicating it should be redelivered.
		Nack()
		// ID returns the unique identifier of the message.
		ID() string
		// Data returns the payload of the message.
		Data() []byte
		// Attributes returns the attributes of the message as a key-value map.
		Attributes() map[string]string
		// PublishTime returns the time when the message was published.
		PublishTime() time.Time
	}

	gcpSubscription struct {
		sub *pubsub.Subscriber
	}

	gcpMessage struct {
		msg *pubsub.Message
	}
)

// ID returns the unique identifier of the subscription.
func (s *gcpSubscription) ID() string {
	if s.sub == nil {
		return ""
	}
	return s.sub.ID()
}

// Receive starts receiving messages from the subscription and processes them using the provided function.
func (s *gcpSubscription) Receive(ctx context.Context, f func(context.Context, message)) error {
	return s.sub.Receive(ctx, func(cctx context.Context, m *pubsub.Message) {
		f(cctx, &gcpMessage{m})
	})
}

// Ack acknowledges the message, indicating successful processing.
func (env *gcpMessage) Ack() {
	env.msg.Ack()
}

// Nack negatively acknowledges the message, indicating it should be redelivered.
func (env *gcpMessage) Nack() {
	env.msg.Nack()
}

// ID returns the unique identifier of the message.
func (env *gcpMessage) ID() string {
	return env.msg.ID
}

// Data returns the payload of the message.
func (env *gcpMessage) Data() []byte {
	return env.msg.Data
}

// Attributes returns the attributes of the message as a key-value map.
func (env *gcpMessage) Attributes() map[string]string {
	return env.msg.Attributes
}

// PublishTime returns the time when the message was published.
func (env *gcpMessage) PublishTime() time.Time {
	return env.msg.PublishTime
}
