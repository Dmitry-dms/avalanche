package main

import (
	"context"
	"fmt"

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
	"github.com/Dmitry-dms/avalanche/pkg/epoll"
	"github.com/Dmitry-dms/avalanche/pkg/msg_controllers/redis"
	"github.com/Dmitry-dms/avalanche/pkg/pool"
	"github.com/Dmitry-dms/avalanche/pkg/serializer/easyjson"
	"github.com/Dmitry-dms/avalanche/pkg/websocket"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"

	//	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"


)

var (
	s     = new(http.Server)
	serve = make(chan error, 1)
	sig   = make(chan os.Signal, 1)
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
func NewZerolog(isDebug bool, name string) *zerolog.Logger {
	logLevel := zerolog.InfoLevel
	if isDebug {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("avalanche_server_name", name).Logger()
	return &logger
}

func ParseConfig() (*internal.Config, error) {
	// loads config from .env file.
	err := godotenv.Load()
	if err != nil {
		return nil, errors.Wrap(err, "ParseConfig")
	}
	var errAtoi error
	addr := os.Getenv("WS_PORT")
	maxConn, errAtoi := strconv.Atoi(os.Getenv("MAX_CONNECTIONS"))
	if errAtoi != nil {
		errors.Wrap(errAtoi, "parse max_connection")
	}
	maxUserMsg, errAtoi := strconv.Atoi(os.Getenv("MAX_USERS_MESSAGES"))
	if errAtoi != nil {
		errors.Wrap(errAtoi, "parse max_users_messages")
	}
	statsIntervalSeconds, errAtoi := strconv.Atoi(os.Getenv("SEND_STATS_USERS_SECOND"))
	if errAtoi != nil {
		errors.Wrap(errAtoi, "parse max_users_messages")
	}
	if errAtoi != nil {
		return nil, errAtoi
	}

	config := internal.Config{
		Port:                  addr,
		Name:                  os.Getenv("NAME"),
		Version:               os.Getenv("VERSION"),
		MaxConnections:        maxConn,
		AuthJWTkey:            os.Getenv("JWT_SECRET"),
		RedisAddress:          os.Getenv("REDIS_ADDRESS"),
		RedisCommandsPrefix:   os.Getenv("REDIS_COMMAND_PREFIX"),
		RedisMsgPrefix:        os.Getenv("REDIS_MSG_PREFIX"),
		RedisInfoPrefix:       os.Getenv("REDIS_INFO_PREFIX"),
		SendStatisticInterval: statsIntervalSeconds,
		MaxUserMessages:       maxUserMsg,
		MonitoringPort:        os.Getenv("MONITORING_PORT"),
	}
	return &config, nil
}

func main() {
	// Increase resources limitations
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	// go func() {
	// 	for range tick.C {
	// 		printMemUsage()
	// 	}
	// }()

	config, err := ParseConfig() // if error, return default config
	zerolg := NewZerolog(true, config.Name)
	if err != nil {
		zerolg.Fatal().Err(errors.Unwrap(err))
	}

	poller, err := epoll.MkEpoll()
	//poller, err := netpoll.New(nil)
	if err != nil {
		zerolg.Fatal().Err(err).Msg("error create netpoll")
	}
	cache := internal.NewRamCache()
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		zerolg.Fatal().Err(err).Msgf("Can't start listening on port %s", config.Port)
	}

	poolConnection := pool.NewPool(2000, 350, 20)
	poolCommands := pool.NewPool(50, 1, 3)

	jsonSerializer := &easyjson.CustomEasyJson{}
	ctx := context.Background()

	upgrader := &websocket.GobwasUpgrader{}

	red := redis.InitRedis(config.RedisAddress)
	engine, errEngine := internal.NewEngine(ctx, *config, zerolg, cache, ln, poolConnection, poolCommands, poller, jsonSerializer, red, upgrader)
	if errEngine != nil {
		zerolg.Fatal().Err(errEngine).Msg("unable to launch the engine")
	}

	errRestoreSt := engine.RestoreState()
	if errRestoreSt != nil {
		zerolg.Debug().Err(err).Msg("unable to find any company")
		engine.Subs.AddCompany("test", 100000, time.Hour*12)
	}

	zerolg.Info().Msgf("accepting websocket connections on port %s", config.Port)

	// go func() {
	// 	mux := http.NewServeMux()
	// 	mux.HandleFunc("/", engine.HandleHTTP)
	// 	log.Printf("run http websocket server on %s", config.Port)
	// 	if err := http.ListenAndServe(":"+config.Port, mux); err != nil {
	// 		zerolg.Fatal().Err(err).Msgf("Can't start listening on port %s", config.MonitoringPort)
	// 	}
	// }()
	// acceptDesc := netpoll.Must(netpoll.HandleListener(ln, netpoll.EventRead|netpoll.EventOneShot))
	// _ = engine.Poller.Start(acceptDesc, func(e netpoll.Event) {
	// 	engine.PoolConnection.Schedule(func() {

	// 		conn, err := ln.Accept()
	// 		if err != nil {
	// 			zerolg.Debug().Err(err)
	// 		} else {
	// 			upg := &websocket.GobwasUpgrader{}
	// 			go engine.Handle(conn, upg)
	// 		}
	// 	})
	// 	_ = engine.Poller.Resume(acceptDesc)
	// })
	go func() {
		for {
			conn, err := ln.Accept()
			engine.PoolConnection.Schedule(func() {
				if err != nil {
					zerolg.Debug().Err(err)
				} else {
					upg := &websocket.GobwasUpgrader{}
					go engine.Handle(conn, upg)
				}
			})
		}
	}()

	go func() {

		mux := http.NewServeMux()
		mux.HandleFunc("/p", func(w http.ResponseWriter, r *http.Request) {
			pprof.Index(w, r)
		})
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/profile", pprof.Profile)
		mux.Handle("/heap", pprof.Handler("heap"))
		mux.Handle("/allocs", pprof.Handler("allocs"))
		mux.Handle("/goroutine", pprof.Handler("goroutine"))
		mux.HandleFunc("/a", mid(engine.GetActiveUsers))
		log.Printf("run http server on %s", config.MonitoringPort)
		if err := http.ListenAndServe(":"+config.MonitoringPort, mux); err != nil {
			zerolg.Fatal().Err(err).Msgf("Can't start monitoring on port %s", config.MonitoringPort)
		}
	}()

	// defer func() {
	// 	recover()
	// 	sig <- syscall.SIGTERM
	// }()
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sig:

		zerolg.Info().Msgf("signal %q received; shutting down", sig)
		// save Hub's TTL to Storer
		if err := engine.SaveState(); err != nil {
			zerolg.Warn().Err(err)
		} else {
			zerolg.Info().Msgf("cache was successfully saved to Storer")
		}
		// delete and disconnect all active users
		engine.DeleteAllClients()
		ctx := context.Background()
		//ln.Close()
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
