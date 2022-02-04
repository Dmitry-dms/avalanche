package main

import (
	"context"

	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/Dmitry-dms/avalanche/internal"
	"github.com/Dmitry-dms/avalanche/pkg/pool"
	"github.com/Dmitry-dms/avalanche/pkg/serializer/json"
	"github.com/mailru/easygo/netpoll"
)

var (
	s     = new(http.Server)
	serve = make(chan error, 1)
	sig   = make(chan os.Signal, 1)

	addr = ":8000"
)

func printMemUsage() {
	// runtime.GC()
	// var m runtime.MemStats
	// runtime.ReadMemStats(&m)
	// fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	// fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	// fmt.Printf("\tStackInuse = %v MiB", bToMb(m.StackInuse))
	// fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	// fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func main() {
	//runtime.GOMAXPROCS(4)

	go func() {
		for {
			time.Sleep(10 * time.Second)
			printMemUsage()
		}
	}()
	// profiler.Start(profiler.Config{
	// 	ApplicationName: "avalanche",
	// 	ServerAddress:   "http://host.docker.internal:4040",
	// })
	config := internal.Config{
		Name:                "ws-1",
		Version:             "1",
		MaxConnections:      100000,
		AuthJWTkey:          "token23",
		RedisAddress:        "host.docker.internal:6160",
		RedisCommandsPrefix: "comm",
		RedisMsgPrefix:      "msg",
		RedisInfoPrefix:     "info",
	}
	poller, err := netpoll.New(nil)
	if err != nil {
		log.Println("error create netpoll: ", err.Error())
	}
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	cache := internal.NewRamCache()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %q error: %v", addr, err)
	}

	poolConnection := pool.NewPool(128, 20, 10)
	poolCommands := pool.NewPool(50, 1, 20)

	jsonSerializer := &json.CustomJsonSerializer{}
	ctx := context.Background()
	engine, _ := internal.NewEngine(ctx, config, infoLog, cache, ln, poolConnection, poolCommands, poller, jsonSerializer)

	engine.Subs.AddCompany("test", "1", 100000, time.Hour*24)

	log.Printf("listening %s (%q)", ln.Addr(), addr)

	//acceptDesc := netpoll.Must(netpoll.HandleListener(ln, netpoll.EventRead|netpoll.EventOneShot))
	//accept := make(chan error, 1)

	//_ = engine.Poller.Start(acceptDesc, func(e netpoll.Event) {
	// We do not want to accept incoming connection when goroutine pool is
	// busy. So if there are no free goroutines during 1ms we want to
	// cooldown the server and do not receive connection for some short
	// time.
	//engine.PoolConnection.Schedule(func() {
	go func() {
		for {
			conn, err := ln.Accept()
			go func(conn net.Conn, err error) {
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Temporary() {
						goto cooldown
					}
					log.Fatalf("accept error: %v", err)
				cooldown:
					delay := 5 * time.Millisecond
					log.Printf("accept error: %v; retrying in %s", err, delay)
					time.Sleep(delay)
					//return
				}
				//accept <- nil
				go engine.Handle(conn)
			}(conn, err)
		}
	}()

	//})
	//if err == nil {
	// err = <-accept
	// //}
	// if err != nil {

	// }

	//_ = engine.Poller.Resume(acceptDesc)
	//})

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/p", func(w http.ResponseWriter, r *http.Request) {
			pprof.Index(w, r)
		})
		mux.HandleFunc("/ws-send", engine.SendToClientById)
		mux.Handle("/heap", pprof.Handler("heap"))
		mux.Handle("/allocs", pprof.Handler("allocs"))
		mux.Handle("/goroutine", pprof.Handler("goroutine"))
		mux.HandleFunc("/a", engine.GetActiveUsers)
		log.Printf("run http server on :8090")
		if err := http.ListenAndServe(":8090", mux); err != nil {
			log.Fatalf("error start listen 8090: %s", err.Error())
		}
	}()

	select {
	case err := <-serve:
		log.Fatalf("error serve: %s", err.Error())
	case sig := <-sig:
		const timeout = 5 * time.Second
		log.Printf("signal %q received; shutting down with %s timeout", sig, timeout)

		ctx, _ := context.WithTimeout(context.Background(), timeout)
		if err := s.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}

}
