package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Redis is a struct that contains all the methods for working with Redis.
type Redis struct {
	Client *redis.Client
}

// InitRedis creates Redis object.
func InitRedis(address string) *Redis {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	return &Redis{Client: r}
}

func (r *Redis) Publish(ctx context.Context, channel string, payload []byte) error {
	return r.Client.Publish(ctx, channel, payload).Err()
}
func (r *Redis) Subscribe(ctx context.Context, channel string, f func(msg string)) {
	sub := r.Client.Subscribe(ctx, channel)
	for msg := range sub.Channel() {
		f(msg.Payload)
	}
}
func (r *Redis) Ping() error {
	return r.Client.Ping(r.Client.Context()).Err()
}
func (r *Redis) SetValue(ctx context.Context, k string, v []byte, ttl time.Duration) error {
	return r.Client.Set(ctx, k, v, ttl).Err()
}
func (r *Redis) GetValue(ctx context.Context, k string) (string, error) {
	return r.Client.Get(ctx, k).Result()
}
func (r *Redis) SetArrayValue(ctx context.Context, k string, vals ...interface{}) error {
	return r.Client.SAdd(ctx, k, vals).Err()
}

// sGetMembers looks for any cached messages the Client has.
func (r *Redis) GetArray(ctx context.Context, k string) ([]string, int) {
	res, err := r.Client.SMembers(ctx, k).Result()
	if err != nil {
		return []string{}, 0
	}
	return res, len(res)
}
func (r *Redis) DeleteKey(ctx context.Context, k string) error {
	return r.Client.Del(ctx, k).Err()
}
