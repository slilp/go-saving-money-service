package kafka

import "context"

type KafkaPubSub interface {
	Subscribe(eventName string, handlerFunc func(payload []byte)) error
	Publish(ctx context.Context, eventName string, payload any) error
	Start()
	Stop()
}

type SubscribeFunc = func(payload []byte)
