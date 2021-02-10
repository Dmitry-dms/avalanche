package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)


func main() {
	red := initRedis("127.0.0.1:6560")
	ps := red.Subscribe(context.Background(), "ws-1")
	maf, err := red.Ping(red.Context()).Result()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(maf)
	for n := range ps.Channel() {
		fmt.Println(n)
	}
}

func initRedis(address string) *redis.Client {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	return r
}