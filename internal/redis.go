package internal

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


func (r *Redis) publish(ctx context.Context, channel string, payload []byte) error {
	return r.Client.Publish(ctx, channel, payload).Err()
}
func (r *Redis) subscribe(ctx context.Context, channel string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channel)
}
func (r *Redis) ping() error {
	return r.Client.Ping(r.Client.Context()).Err()
}
func (r *Redis) setKV(ctx context.Context, k string, v []byte, ttl time.Duration) error {
	return r.Client.Set(ctx, k, v, ttl).Err()
}
func (r *Redis) getV(ctx context.Context, k string) (string, error) {
	return r.Client.Get(ctx, k).Result()
}
func (r *Redis) sAdd(ctx context.Context, k string, vals ...interface{}) {
	r.Client.SAdd(ctx, k, vals)
}
// sGetMembers looks for any cached messages the Client has.
func (r *Redis) sGetMembers(ctx context.Context, k string) ([]string, int) {
	res, err := r.Client.SMembers(ctx, k).Result()
	if err != nil {
		return []string{}, 0
	}
	return res, len(res)
}
func (r *Redis) deleteK(ctx context.Context, k string) string {
	return r.Client.Del(ctx, k).String()
}
