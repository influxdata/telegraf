package cloud_pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"time"
)

type (
	subscription interface {
		ID() string
		Receive(ctx context.Context, f func(context.Context, message)) error
	}

	message interface {
		Ack()
		Nack()
		ID() string
		Data() []byte
		Attributes() map[string]string
		PublishTime() time.Time
	}

	gcpSubscription struct {
		sub *pubsub.Subscription
	}

	gcpMessage struct {
		msg *pubsub.Message
	}
)

func (s *gcpSubscription) ID() string {
	if s.sub == nil {
		return ""
	}
	return s.sub.ID()
}

func (s *gcpSubscription) Receive(ctx context.Context, f func(context.Context, message)) error {
	return s.sub.Receive(ctx, func(cctx context.Context, m *pubsub.Message) {
		f(cctx, &gcpMessage{m})
	})
}

func (env *gcpMessage) Ack() {
	env.msg.Ack()
}

func (env *gcpMessage) Nack() {
	env.msg.Nack()
}

func (env *gcpMessage) ID() string {
	return env.msg.ID
}

func (env *gcpMessage) Data() []byte {
	return env.msg.Data
}

func (env *gcpMessage) Attributes() map[string]string {
	return env.msg.Attributes
}

func (env *gcpMessage) PublishTime() time.Time {
	return env.msg.PublishTime
}
