package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	ctx := context.Background()
	conn, _, _, err := ws.Dial(ctx, "ws://localhost:8080/ws")
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}
	go delay(conn)
	err = wsutil.WriteMessage(conn, ws.StateClientSide, ws.OpBinary, []byte("Hello"))
	if err != nil {
		log.Fatal(err)

	}
	for {
		h, r, err := wsutil.NextReader(conn, ws.StateClientSide)
		if err != nil {
			fmt.Println(err)
		}
		if h.OpCode == ws.OpPong {
			fmt.Println("called pong")
		}
		if h.OpCode.IsControl() {
			fmt.Println("is control")
		}

		msg, err := ioutil.ReadAll(r)
		if err != nil {
			fmt.Println(err)
			return 
		}
		// msg, err := wsutil.ReadServerBinary(conn)
		// if err != nil {
		// 	log.Fatal(err)

		// }
		log.Printf("Meesage {%s} ", msg)
	}

}
func delay(conn net.Conn) {
	time.Sleep(1 * time.Second)

	//_ = wsutil.WriteClientMessage(conn, ws.OpClose, []byte("close connection bhrjrjrjrjh"))
	_ = wsutil.WriteClientMessage(conn, ws.OpPing, []byte("ping"))
}
