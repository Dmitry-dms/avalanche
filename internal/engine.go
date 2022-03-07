package internal

import (
	"context"

	"strings"

	"fmt"
	"time"

	"net"
	"net/http"

	"github.com/Dmitry-dms/avalanche/pkg/auth"
	"github.com/Dmitry-dms/avalanche/pkg/pool"
	"github.com/Dmitry-dms/avalanche/pkg/serializer"
	"github.com/Dmitry-dms/avalanche/pkg/websocket"
	"github.com/gobwas/ws"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/mailru/easygo/netpoll"
)

// Engine struct is a core of Avalanche websocket server.
// It contains all the logic for working with websockets, cache, message broker, logging.
type Engine struct {
	Context context.Context
	// conf contains the main configuration.
	conf Config
	// log is an implementaion of zerolog library.
	log *zerolog.Logger
	// Subs is an implementation of Cache interface.
	// Used for creating Hubs where users are held.
	Subs Cache
	// Server is used for exposing pprof handlers.
	Server net.Listener
	// upgrader is an implementation of the Upgrader interface.
	upgrader Upgrader
	// PoolConnection is a goroutine pool which helps with accepting many connections at once.
	PoolConnection *pool.Pool
	// PoolCommands reduces the creation of a large number of goroutines when working with
	// serialization/deserialization, receiving messages from message broker.
	PoolCommands *pool.Pool
	// Poller is an implementation of Linux epoll.
	Poller netpoll.Poller
	// msgController is an implementation of MessageController interface.
	// It can publish/subscribe to the channels of the message broker.
	// It caches CompanyHub's state and messages to offline users on shutdown and restores on startup.
	msgController MessageController

	// sendInfoFunc is a defined func to send messages to the message broker.
	sendInfoFunc func(payload []byte) error

	// serializer is an implementation of serializer interface.
	// Used for serializing/deserializng messages between Avalanche and message broker.
	serializer serializer.Serializer
	// authManager used for manipulating with JWT.
	authManager *auth.Manager
}

// NewEngine creates a core of Avalanche websocket server and initiates connection to the message broker.
func NewEngine(ctx context.Context, config Config, logger *zerolog.Logger, cache Cache,
	conn net.Listener, poolConn, poolComm *pool.Pool,
	poller netpoll.Poller, s serializer.Serializer, msgController MessageController, upg Upgrader) (*Engine, error) {

	sendInfoFunc := func(payload []byte) error {
		return msgController.Publish(ctx, config.RedisInfoPrefix, payload) //+config.Name
	}

	authManager, err := auth.NewManager(config.AuthJWTkey)
	if err != nil {
		return nil, err
	}
	engine := &Engine{
		Context:        ctx,
		conf:           config,
		log:         logger,
		Subs:           cache,
		Server:         conn,
		PoolConnection: poolConn,
		PoolCommands:   poolComm,
		Poller:         poller,
		sendInfoFunc:   sendInfoFunc,
		serializer:     s,
		authManager:    authManager,
		msgController:  msgController,
		upgrader:       upg,
	}
	if err := msgController.Ping(); err != nil {
		engine.log.Fatal().Err(err).Msg("can't connect to message broker")
	}
	go msgController.Subscribe(ctx, config.RedisCommandsPrefix, engine.listeningCommands())
	go msgController.Subscribe(ctx, config.RedisMsgPrefix, engine.startRedisListen())
	//go engine.sendStatisticAboutUsers()

	engine.startupMessage([]byte(fmt.Sprintf("WS server: {%s} {%s} succesfully connected to message broker. Port = %s", config.Name, config.Version, config.Port)))
	return engine, nil
}

func (e *Engine) startupMessage(msg []byte) error {
	return e.sendInfoFunc(msg)
}

// listeningCommands creates a function that runs when a message arrives in the command channel.
func (e *Engine) listeningCommands() func(msg string) {
	f := func(msg string) {
		e.PoolCommands.Schedule(func() {
			var c AddCompanyMessage
			err := e.serializer.Unmarshal([]byte(msg), &c)
			if err != nil {
				e.log.Warn().Err(err)
				return
			}
			token, err := e.authManager.NewJWT(c.CompanyName, time.Duration(c.Duration)*time.Hour)
			if err != nil {
				e.log.Warn().Err(err)
				return
			}
			err = e.Subs.AddCompany(c.CompanyName, c.MaxUsers, time.Duration(c.Duration))
			if err != nil {
				e.log.Warn().Err(err)
				return
			}
			resp := &AddCompanyResponse{
				CompanyName: c.CompanyName,
				Token: CompanyToken{
					ServerName: e.conf.Name,
					Token:      token,
					Duration:   c.Duration,
				},
			}
			err = e.serializeAndSend(resp)
			if err != nil {
				e.log.Warn().Err(err)
				return
			}
		})
	}
	return f
}

// serializeAndSend is a helper function which serializes message and sends it to the Redis.
func (e *Engine) serializeAndSend(v interface{}) error {
	payload, err := e.serializer.Marshal(v)
	if err != nil {
		e.log.Debug().Err(err)
		return err
	}
	err = e.sendInfoFunc(payload)
	if err != nil {
		e.log.Debug().Err(err)
		return err
	}
	return nil
}

// sendStatisticAboutUsers sends information about all users to the Redis.
func (e *Engine) sendStatisticAboutUsers() {
	ticker := time.NewTicker(time.Second * time.Duration(e.conf.SendStatisticInterval))
	defer ticker.Stop()
	for range ticker.C {
		e.PoolCommands.Schedule(func() {
			companies := e.Subs.GetStatisctics()
			if companies.Stats == nil {
				return
			}
			err := e.serializeAndSend(companies)
			if err != nil {
				e.log.Debug().Err(err)
				return
			}
		})
	}
}

// startRedisListen creates a function that runs when a message arrives for the Сlient.
func (e *Engine) startRedisListen() func(msg string) {
	e.log.Info().Msg("redis listener was started")

	f := func(msg string) {
		e.PoolCommands.Schedule(func() {
			var m brokerMessage
			err := e.serializer.Unmarshal([]byte(msg), &m)
			if err != nil {
				e.log.Warn().Err(err)
				return
			}
			client, isOnline := e.Subs.GetClient(m.CompanyName, m.ClientId)
			// if user is offline check if there is free space to store messages
			if !isOnline {
				_, length := e.msgController.GetArray(context.TODO(), m.ClientId)
				if length > int(e.conf.MaxUserMessages) {
					e.log.Warn().Msgf("user with id=%s has reached the limit of cached messages", m.ClientId)
				} else {
					e.msgController.SetArrayValue(context.TODO(), m.ClientId, msg)
				}
				return
			}
			e.log.Debug().Msgf("Message {%s} to client {%s} with company id {%s}", m.Message, m.ClientId, m.CompanyName)
			e.sendMsg(Message{client, []byte(msg)}, m.CompanyName)
		})
	}
	return f
}

// func (c *Client) HandleWrite(msg []byte, msgType ws.OpCode) error {
// 	err := c.Connection.Write(msg)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// HandleRead is a wrapper for the connection Read function.
func (e *Engine) HandleRead(c *Client) ([]byte, bool, error) {
	payload, isControl, err := c.Connection.Read()
	if err != nil {
		return nil, isControl, errors.Wrap(err, "HandleRead error")
	}
	return payload, isControl, nil
}

func (e *Engine) ParseToken(accessToken string) (string, error) {
	return e.authManager.Parse(accessToken)
}

// Handle is a function that upgrades the websocket connection,
// creates a client, fires a poller for listening messages from client.
func (e *Engine) Handle(conn net.Conn, upgr Upgrader) {
	var userId, companyName string

	//err := upgr.Upgrade(conn, 256, 1024, &userId, &companyName, e.ParseToken)
	u := ws.Upgrader{
		ReadBufferSize:  256,
		WriteBufferSize: 1024,
		OnHeader: func(key, value []byte) error {
			if string(key) == "User" {
				userId = string(value)
			} else if string(key) == "Token" {
				var err error
				companyN, err := e.ParseToken(string(value))
				if err != nil {
					return ws.RejectConnectionError(
						ws.RejectionReason(fmt.Sprintf("bad token: %s", err)),
						ws.RejectionStatus(400))
				}
				companyName = companyN
			}
			return nil
		},
	}
	_, err := u.Upgrade(conn)
	if err != nil {
		e.log.Warn().Err(err).Msg("upgrade error")
		_ = conn.Close()
		return
	}

	transport := websocket.NewWebsocketTransport(conn)

	//readDescriptor, _ := netpoll.Handle(conn, netpoll.EventRead)

	cachedMessages, length := e.msgController.GetArray(context.TODO(), userId)

	client := NewClient(transport, userId)
	err, deleteClient := e.Subs.AddClient(companyName, client)
	if err != nil {
		e.log.Info().Err(err)
		conn.Write([]byte("user already exists"))
		conn.Close()
		return
	}

	e.log.Debug().Msgf("user connected with id={%s} and company={%s}", client.UserId, companyName)

	if length > 0 {
		e.PoolCommands.Schedule(func() {
			for _, msg := range cachedMessages {
				e.sendMsg(Message{client, []byte(msg)}, companyName)
			}
			e.msgController.DeleteKey(context.TODO(), client.UserId)
		})
	}
	go func(client *Client){
		for {
			payload, isControl, err := e.HandleRead(client)
			if err != nil {
				//_ = e.Poller.Stop(readDescriptor)
				deleteClient()
				e.log.Debug().Err(err)
				return
			} else {
				if !isControl {
					e.sendMsg(Message{Client: client, Msg: []byte(strings.ToUpper(string(payload)))}, companyName)
					e.log.Debug().Msgf("Message from user={%s}: {%s}\n", client.UserId, payload)
				}
				//_ = e.Poller.Resume(readDescriptor)
			}
		}
	}(client)

	// Start a poller to notify a goroutine when the message comes in.
		// _ = e.Poller.Start(readDescriptor, func(ev netpoll.Event) {

		// 	if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
		// 		_ = e.Poller.Stop(readDescriptor)
		// 		deleteClient()
		// 		//e.log.Debug().Msgf("user with id={%s} was disconnected", client.UserId)
		// 		//e.log.Debug().Msg(ev.String())
		// 		return
		// 	}
		// 	payload, isControl, err := e.HandleRead(client)
		// 	if err != nil {
		// 		_ = e.Poller.Stop(readDescriptor)
		// 		e.log.Debug().Err(err)
		// 		return
		// 	} else {
		// 		if !isControl {
		// 			e.sendMsg(Message{Client: client, Msg: []byte(strings.ToUpper(string(payload)))}, companyName)
		// 			e.log.Debug().Msgf("Message from user={%s}: {%s}\n", client.UserId, payload)
		// 		}
		// 		_ = e.Poller.Resume(readDescriptor)
		// 	}
		// })
}

// SaveState saves the state of active CompanyHubs.
func (e *Engine) SaveState() error {
	stats := e.Subs.GetStatisctics()
	data, err := e.serializer.Marshal(stats)
	if err != nil {
		e.log.Debug().Err(err)
		return err
	}
	ctx := context.Background()
	return e.msgController.SetValue(ctx, "save", data, 0)
}
func (e *Engine) DeleteAllClients() {
	companies := e.Subs.GetAllCompanies()
	for _, c := range companies {
		c.deleteAllHubClients()
	}
}

// RestoreState restores the state of active CompanyHubs after restart.
func (e *Engine) RestoreState() error {
	ctx := context.TODO()
	val, err := e.msgController.GetValue(ctx, "save")
	if err != nil {
		return errors.New("key doesn't exists")
	}
	var stats CompanyStatsWrapper
	err = e.serializer.Unmarshal([]byte(val), &stats)
	if err != nil {
		return err
	}
	for _, c := range stats.Stats {
		
		if c.Expired == true {
			continue
		}
		ttl := c.Stopped.Sub(c.Time)
		e.log.Debug().Msgf("restored - %s ttl - %s", c.Name, c.TTL-ttl)

		e.Subs.AddCompany(c.Name, c.MaxUsers, c.TTL-ttl)
	}
	return nil
}

func (e *Engine) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	users, _ := e.Subs.GetActiveUsers("test")
	_, _ = w.Write([]byte(fmt.Sprintf("%d", users)))
	e.log.Printf("Active users = %d", users)
}

// sendMsg is a helper function that allows you to send messages to the Client.
func (e *Engine) sendMsg(msg Message, compName string) bool {
	e.Subs.SendMessage(msg, compName)
	//client.MessageChan <- msg
	//client.HandleWrite([]byte(msg), ws.OpText)
	return true
}
