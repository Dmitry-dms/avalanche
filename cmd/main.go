package main

import (
	"context"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dmitry-dms/websockets/internal/core"
)

var (
	s     = new(http.Server)
	serve = make(chan error, 1)
	sig   = make(chan os.Signal, 1)
	addr  = ":8080"
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
	http.HandleFunc("/ws", engine.HandleClient)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %q error: %v", addr, err)
	}
	log.Printf("listening %s (%q)", ln.Addr(), addr)

	go startServer(engine.MsgChannel)
	signal.Notify(sig, syscall.SIGTERM|syscall.SIGINT)
	go func() { serve <- s.Serve(ln) }()

	select {
	case err := <-serve:
		log.Fatal(err)
	case sig := <-sig:
		const timeout = 5 * time.Second

		log.Printf("signal %q received; shutting down with %s timeout", sig, timeout)

		ctx, _ := context.WithTimeout(context.Background(), timeout)
		if err := s.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}
	//log.Fatal(http.ListenAndServe(":8080", nil))
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
