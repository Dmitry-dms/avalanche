package main

import (
	"log"
	"net/http"
	"os"
	"time"

//	"github.com/gobwas/ws"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	k:=2
	chd := make(chan int)
	head := http.Header{}
	head.Add("Cookie", uuid.NewString())
	// dialer := ws.Dialer{
	// 	Timeout: time.Duration(100),
	// 	WriteBufferSize: 1024,
	// 	ReadBufferSize: 1024,
	// 	Header: ws.HandshakeHeaderHTTP{head},
	// }
	
	for i := 0; i < 10000; i++ {
		time.Sleep(time.Millisecond*time.Duration(k))
		go func(i int) {
			head := http.Header{}
			head.Add("Cookie", uuid.NewString())
			c, _, err := websocket.DefaultDialer.Dial("ws://host.docker.internal:8080/", head)
			if err != nil {
				log.Printf("format string %s", err)
				//c.Close()
				return
			}
			defer c.Close()
			done := make(chan struct{})

			go func() {
				defer close(done)
				for {
					
					_, message, err := c.ReadMessage()
					if err != nil {
						log.Println("read:", err)
						log.Printf("err from read %s", err)
						return
					}
					log.Printf("recv: %s", message)
				}
			}()

			ticker := time.NewTicker(time.Second*1000000)
			defer ticker.Stop()

			for {

				select {
				case <-done:

					return
				case t := <-ticker.C:
					err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
					if err != nil {
						log.Println("write:", err)
						log.Printf("err fro write %s", err)
						return
					}
				case <-interrupt:
					log.Println("interrupt")

					// Cleanly close the connection by sending a close message and then
					// waiting (with timeout) for the server to close the connection.
					err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						log.Println("write close:", err)
						return
					}
					select {
					case <-done:
					case <-time.After(time.Second):
					}
					return
				 }
			}
			
		}(i)
	}
	<-chd
}

