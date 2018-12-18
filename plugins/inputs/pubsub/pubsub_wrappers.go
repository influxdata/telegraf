package pubsub

import (
	"time"
	"cloud.google.com/go/pubsub"
	"context"
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

	subWrapper struct {
		sub *pubsub.Subscription
	}

	msgWrapper struct {
		msg *pubsub.Message
	}
)

func (s* subWrapper) ID() string {
	if s.sub == nil {
		return ""
	}
	return s.sub.ID()
}

func (s* subWrapper) Receive(ctx context.Context, f func(context.Context, message)) error {
	return s.sub.Receive(ctx, func(cctx context.Context, m *pubsub.Message) {
		f(cctx, &msgWrapper{m})
	})
}

func (env *msgWrapper) Ack() {
	env.msg.Ack()
}

func (env *msgWrapper) Nack() {
	env.msg.Nack()
}

func (env *msgWrapper) ID() string {
	return env.msg.ID
}

func (env *msgWrapper) Data() []byte {
	return env.msg.Data
}

func (env *msgWrapper) Attributes() map[string]string {
	return env.msg.Attributes
}

func (env *msgWrapper) PublishTime() time.Time {
	return env.msg.PublishTime
}
