package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Dmitry-dms/websockets/internal/core"
)

func main() {
	//chat := make(chan string, 1)
	// _, err := netpoll.New(nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	conn, _, _, err := ws.UpgradeHTTP(r, w)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	go func() {
	// 		defer conn.Close()
	// 		for {
	// 				msg:= <-chat
	// 				err = wsutil.WriteServerMessage(conn, ws.OpBinary, []byte(msg))
	// 				if err != nil {
	// 					log.Fatal(err)
	// 				}
	// 		}
	// 	}()
	// 	go func(){
	// 		defer conn.Close()
	// 		for {
	// 			// msg, op, err := wsutil.ReadClientData(conn)
	// 			payload, err := wsutil.ReadClientText(conn)
	// 			if err != nil {
	// 				log.Fatal(err)
	// 		   }
	// 		   log.Printf("Meesage {%s}", payload)
	// 		}
	// 	}()
	// })
	config := core.Config{
		Name:           "ws-1",
		Version:        "1",
		MaxConnections: 100,
	}
	hub := core.NewHub()
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	engine := core.NewEngine(config, infoLog, hub)
	http.HandleFunc("/", engine.HandleClient)
	go startServer(engine.MsgChannel)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func startServer(in chan string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/s", func(w http.ResponseWriter, r *http.Request) {
		msg := r.RemoteAddr
		log.Println(msg)
		in <- msg
	})
	log.Fatal(http.ListenAndServe(":8000", mux))
}
