package main

import (
	"context"
	"fmt"

	"github.com/Dmitry-dms/avalanche/pkg/serializer/json"
	"github.com/go-redis/redis/v8"
)

func main() {
	red := initRedis("127.0.0.1:6560")
	//ps := red.Subscribe(context.Background(), "ws-2")
	m := &redisMessage{
		CompanyName: "test",
		ClientId: "test",
		Message: "hello",
	}
	ser := &json.CustomJsonSerializer{}
	b, err := ser.Serialize(m)
	if err != nil {
		fmt.Println(err)
		return // TODO: Handle error
	}
	msg, err := red.Publish(context.Background(), "ws-1", b).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(msg)
}

type redisMessage struct {
	CompanyName string
	ClientId    string
	Message     string
}

func initRedis(address string) *redis.Client {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	return r
}
