package internal

import (
	"context"
	"time"
)


type Publisher interface {
	Publish(ctx context.Context, channel string, payload []byte) error
}
type Subscriber interface {
	Subscribe(ctx context.Context, channel string, f func(msg string))
}
type Storer interface {
	GetValue(ctx context.Context, k string) (string, error)
	SetValue(ctx context.Context, k string, v []byte, ttl time.Duration) error
	SetArrayValue(ctx context.Context, k string, vals ...interface{}) error
	GetArray(ctx context.Context, k string) ([]string, int)
	DeleteKey(ctx context.Context, k string) error
}

type MessageController interface {
	Publisher
	Subscriber
	Storer
	Ping() error
}