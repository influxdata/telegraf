package cloud_pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
)

type (
	topic interface {
		ID() string
		Stop()
		Publish(ctx context.Context, msg *pubsub.Message) publishResult
		PublishSettings() pubsub.PublishSettings
		SetPublishSettings(settings pubsub.PublishSettings)
	}

	publishResult interface {
		Get(ctx context.Context) (string, error)
	}

	topicWrapper struct {
		topic *pubsub.Topic
	}
)

func (tw *topicWrapper) ID() string {
	return tw.topic.ID()
}

func (tw *topicWrapper) Stop() {
	tw.topic.Stop()
}

func (tw *topicWrapper) Publish(ctx context.Context, msg *pubsub.Message) publishResult {
	return tw.topic.Publish(ctx, msg)
}

func (tw *topicWrapper) PublishSettings() pubsub.PublishSettings {
	return tw.topic.PublishSettings
}

func (tw *topicWrapper) SetPublishSettings(settings pubsub.PublishSettings) {
	tw.topic.PublishSettings = settings
}
