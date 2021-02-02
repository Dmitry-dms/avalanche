package main

import (
	"fmt"

	"github.com/Dmitry-dms/avalanche/internal/core"
	"github.com/Dmitry-dms/avalanche/internal/repository/in_memory_cache"
	"github.com/Dmitry-dms/avalanche/pkg/websocket"
)



func main() {
	cache := in_memory_cache.NewRamCache()
	cache.AddCompany("test", 10)
	cache.AddClient("test", &core.Client{UserId: "1",Connection: &websocket.CustomWebsocketTransport{}})
	cache.AddClient("test", &core.Client{UserId: "2",Connection: &websocket.CustomWebsocketTransport{}})
	cache.AddClient("test", &core.Client{UserId: "3",Connection: &websocket.CustomWebsocketTransport{}})
	err := cache.AddClient("test", &core.Client{UserId: "3",Connection: &websocket.CustomWebsocketTransport{}})
	if err != nil {
		fmt.Println(err)
	}
	cl, ok := cache.GetClient("test", "3")
	fmt.Printf("client = %v, bool = %v", cl,ok)
	fmt.Println(cache.GetCompany("test").GetNumActiveUsers())
	// err := cache.GetCompany("test").AddClient("5", &core.Client{UserId: "5",Connection: &websocket.CustomWebsocketTransport{}})
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//fmt.Println(cache.GetCompany("test").GetUsers())
}

