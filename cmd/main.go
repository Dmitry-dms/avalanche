package main

import (
	"context"
	"fmt"
	"runtime"

	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/Dmitry-dms/avalanche/internal/core"
	"github.com/Dmitry-dms/avalanche/pkg/websocket"
	//"github.com/pyroscope-io/pyroscope/pkg/agent/profiler"
)

var (
	s     = new(http.Server)
	serve = make(chan error, 1)
	sig   = make(chan os.Signal, 1)
	addr  = ":8080"
)

func printMemUsage() {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tStackInuse = %v MiB", bToMb(m.StackInuse))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func main() {
	runtime.GOMAXPROCS(2)
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
	config := core.Config{
		Name:           "ws-1",
		Version:        "1",
		MaxConnections: 100000,
	}
	
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	cache := core.NewRamCache()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %q error: %v", addr, err)
	}
	
	pool := websocket.NewPool(10, 1, 1)
	engine := core.NewEngine(config, infoLog, cache, ln, pool)
	
	//ticker := time.NewTicker(time.Second * 60)
	// go func() {
	// 	for range ticker.C {
	// 		engine.Subs.DeleteOfflineClients()
	// 	}
	// }()
	engine.Subs.AddCompany("test", 100000)

	http.HandleFunc("/ws", engine.SubscribeClient)       // Header: "user-id":....
	http.HandleFunc("/ws-send", engine.SendToClientById) // Header: "user-id","company-name","payload"
	http.HandleFunc("/a", engine.GetActiveUsers)
	//ln, err := net.Listen("tcp", addr)

	log.Printf("listening %s (%q)", ln.Addr(), addr)

	go func() {
		serve <- s.Serve(ln)
	}()

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
