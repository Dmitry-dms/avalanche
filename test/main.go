package main

import (

	//	"fmt"
	"fmt"
	"log"
	"net/http"
	"sync"


	//"github.com/Dmitry-dms/avalanche/internal/core"
	//"github.com/Dmitry-dms/avalanche/pkg/websocket"
)

func main() {
	// cache := core.NewRamCache()
	// cache.AddCompany("test", 10)
	// cache.AddClient("test", &core.Client{UserId: "1",Connection: &websocket.CustomWebsocketTransport{}})
	// cache.AddClient("test", &core.Client{UserId: "2",Connection: &websocket.CustomWebsocketTransport{}})
	// cache.AddClient("test", &core.Client{UserId: "3",Connection: &websocket.CustomWebsocketTransport{}})
	// err, fn := cache.AddClient("test", &core.Client{UserId: "3",Connection: &websocket.CustomWebsocketTransport{}})
	// if err != nil {
	// 	fmt.Println(err)
	// 	fn()
	// }
	// cl, ok := cache.GetClient("test", "3")
	// fmt.Printf("client = %v, bool = %v", cl,ok)
	// fmt.Println(cache.GetCompany("test").GetNumActiveUsers())
	// err := cache.GetCompany("test").AddClient("5", &core.Client{UserId: "5",Connection: &websocket.CustomWebsocketTransport{}})
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//fmt.Println(cache.GetCompany("test").GetUsers())
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		//k:=1
		go requests(fmt.Sprintf("user-%d",i), fmt.Sprintf("Hello to user-%d", i), &wg)
	}
	wg.Wait()
}
func requests(id, payload string, wg *sync.WaitGroup) {
	
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/ws-send", nil)
	if err != nil {
		log.Println(err)
	}
	// Получаем и устанавливаем тип контента
	req.Header.Set("user-id", id)
	req.Header.Set("company-name", "test")
	req.Header.Set("payload", payload)
	// Отправляем запрос
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Printf("error from do: %s", err)
	}
	wg.Done()
}
