package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Dmitry-dms/avalanche/pkg/auth"
	"github.com/gobwas/ws"
)

func main() {
	ln, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	var userId, companyName string
	manager, err := auth.NewManager("test")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(manager.NewJWT("test", time.Hour*10))
	u := ws.Upgrader{
		ReadBufferSize:  256,
		WriteBufferSize: 1024,
		OnHeader: func(key, value []byte) error {
			if string(key) == "User" {
				userId = string(value)
			}
			if string(key) == "Token" {
				var err error
				companyName, err = manager.Parse(string(value))
				if err != nil {
					return ws.RejectConnectionError(
						ws.RejectionReason(fmt.Sprintf("bad token: %s", err)),
						ws.RejectionStatus(400))
				}
			}
			return nil
		},
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		_, err = u.Upgrade(conn)
		if err != nil {
			log.Println(err)
		}
		log.Printf("user-id = {%s} token = {%s}", userId, companyName)
	}
}
