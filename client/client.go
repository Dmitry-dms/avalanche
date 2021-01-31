package main

import (
	"context"
	"log"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	ctx := context.Background()
	conn, _, _, err := ws.Dial(ctx, "ws://localhost:8080")
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = wsutil.WriteMessage(conn, ws.StateClientSide, ws.OpBinary, []byte("Hello"))
	if err != nil {
		log.Fatal(err)

	}
	for {
		msg, err := wsutil.ReadServerBinary(conn)
		if err != nil {
			log.Fatal(err)
	
		}
		log.Printf("Meesage {%s} ", msg)
	}

}
