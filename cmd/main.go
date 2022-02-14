package main

import (
	"context"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/Dmitry-dms/avalanche/internal"
	"github.com/Dmitry-dms/avalanche/pkg/pool"
	"github.com/Dmitry-dms/avalanche/pkg/serializer/json"
	"github.com/arl/statsviz"
	"github.com/joho/godotenv"

	//"github.com/prometheus/client_golang/prometheus"
	//"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"github.com/mailru/easygo/netpoll"
)

// TODO:

// 2. Log to file

// 4. Accept simultaneous connections without crash

var (
	s     = new(http.Server)
	serve = make(chan error, 1)
	sig   = make(chan os.Signal, 1)
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
func NewConsole(isDebug bool) *zerolog.Logger {
	logLevel := zerolog.InfoLevel
	if isDebug {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &logger
}

// var getWsCounter = prometheus.NewCounterVec(
// 	prometheus.CounterOpts{
// 		Name: "ws_request_connect", // metric name
// 		Help: "Number of ws_connects request.",
// 	},
// 	[]string{"status"}, // labels
// )
// func init() {
//     // must register counter on init
// 	prometheus.MustRegister(getWsCounter)
// }
func main() {
	//runtime.GOMAXPROCS(4)

	// go func() {
	// 	for {
	// 		time.Sleep(10 * time.Second)
	// 		printMemUsage()
	// 	}
	// }()
	// profiler.Start(profiler.Config{
	// 	ApplicationName: "avalanche",
	// 	ServerAddress:   "http://host.docker.internal:4040",
	// })
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	addr := os.Getenv("WS_PORT")
	maxConn, _ := strconv.Atoi(os.Getenv("MAX_CONNECTIONS"))
	maxUserMsg, _ := strconv.Atoi(os.Getenv("MAX_USERS_MESSAGES"))
	config := internal.Config{
		Name:                os.Getenv("NAME"),
		Version:             os.Getenv("VERSION"),
		MaxConnections:      maxConn,
		AuthJWTkey:          os.Getenv("JWT_SECRET"),
		RedisAddress:        os.Getenv("REDIS_ADDRESS"),
		RedisCommandsPrefix: os.Getenv("REDIS_COMMAND_PREFIX"),
		RedisMsgPrefix:      os.Getenv("REDIS_MSG_PREFIX"),
		RedisInfoPrefix:     os.Getenv("REDIS_INFO_PREFIX"),
		MaxUserMessages:     maxUserMsg,
	}
	zerolg := NewConsole(true)
	poller, err := netpoll.New(nil)
	if err != nil {
		zerolg.Fatal().Err(err).Msg("error create netpoll")
	}

	writeFunc := func(msg internal.Message) {
		msg.Client.Connection.Write(msg.Msg)
	}

	//infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	cache := internal.NewRamCache(writeFunc, poller)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		zerolg.Fatal().Err(err).Msgf("Can't start listening on port %s", addr)
		//log.Fatalf("listen %q error: %v", addr, err)
	}

	poolConnection := pool.NewPool(128, 20, 3)
	poolCommands := pool.NewPool(50, 1, 3)

	jsonSerializer := &json.CustomJsonSerializer{}
	ctx := context.Background()
	engine, _ := internal.NewEngine(ctx, config, zerolg, cache, ln, poolConnection, poolCommands, poller, jsonSerializer)

	err = engine.RestoreState()
	if err != nil {
		zerolg.Debug().Err(err).Msg("unable to find any company")
		engine.Subs.AddCompany("test", 100000, time.Hour*12)
	}

	//log.Printf("listening %s (%q)", ln.Addr(), addr)

	acceptDesc := netpoll.Must(netpoll.HandleListener(ln, netpoll.EventRead|netpoll.EventOneShot))
	//accept := make(chan error, 1)

	_ = engine.Poller.Start(acceptDesc, func(e netpoll.Event) {

		engine.PoolConnection.Schedule(func() {
			//go func() {
			//for {
			conn, err := ln.Accept()
			//go func(conn net.Conn, err error) {
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

			go engine.Handle(conn)

			//getWsCounter.WithLabelValues(status).Inc()
		})
		_ = engine.Poller.Resume(acceptDesc)
	})

	go func() {
		port := os.Getenv("MONITORING_PORT")
		mux := http.NewServeMux()
		mux.HandleFunc("/p", func(w http.ResponseWriter, r *http.Request) {
			pprof.Index(w, r)
		})
		statsviz.Register(mux)
		mux.HandleFunc("/ws-send", engine.SendToClientById)
		//mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/profile", pprof.Profile)
		mux.Handle("/heap", pprof.Handler("heap"))
		mux.Handle("/allocs", pprof.Handler("allocs"))
		mux.Handle("/goroutine", pprof.Handler("goroutine"))
		mux.HandleFunc("/a", mid(engine.GetActiveUsers))
		log.Printf("run http server on %s", port)
		if err := http.ListenAndServe(port, mux); err != nil {
			zerolg.Fatal().Err(err).Msgf("Can't start monitoring on port %s", port)
		}
	}()
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sig:
		//const timeout = 5 * time.Second
		zerolg.Info().Msgf("signal %q received; shutting down", sig)
		engine.SaveState()
		zerolg.Info().Msgf("successfully saved cache to redis")
		ctx := context.Background()
		if err := s.Shutdown(ctx); err != nil {
			zerolg.Fatal().Err(err)
		}
	}

}
func mid(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		runtime.GC()
		h.ServeHTTP(rw, r)
	}
}
