package internal

import (
	"bufio"
	"context"
	"os"
	"strings"

	"fmt"
	"time"

	//"sync"

	"net"
	"net/http"

	"github.com/Dmitry-dms/avalanche/pkg/auth"
	"github.com/Dmitry-dms/avalanche/pkg/pool"
	"github.com/Dmitry-dms/avalanche/pkg/serializer"
	"github.com/Dmitry-dms/avalanche/pkg/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/go-redis/redis/v8"
	"github.com/mailru/easygo/netpoll"

	"github.com/gobwas/ws"
)

type Engine struct {
	Context          context.Context
	Conf             Config
	Logger           *zerolog.Logger
	Subs             Cache
	Server           net.Listener
	PoolConnection   *pool.Pool
	PoolCommands     *pool.Pool
	Poller           netpoll.Poller
	Redis            *redis.Client
	RedisMsgSub      *redis.PubSub
	RedisSendInfo    func(payload []byte) error
	RedisCommandsSub *redis.PubSub
	Serializer       serializer.AvalancheSerializer
	AuthManager      *auth.Manager
}

func NewEngine(ctx context.Context, config Config, logger *zerolog.Logger, cache Cache,
	conn net.Listener, poolConn *pool.Pool, poolComm *pool.Pool,
	poller netpoll.Poller, s serializer.AvalancheSerializer) (*Engine, error) {

	red := initRedis(config.RedisAddress)

	redisInfo := func(payload []byte) error {
		return red.Publish(ctx, config.RedisInfoPrefix, payload).Err() //+config.Name
	}

	redisMsg := red.Subscribe(ctx, config.RedisMsgPrefix)       //+config.Name)
	redisMain := red.Subscribe(ctx, config.RedisCommandsPrefix) //+config.Name)
	authManager, err := auth.NewManager(config.AuthJWTkey)
	if err != nil {
		return nil, err
	}
	engine := &Engine{
		Context:          ctx,
		Conf:             config,
		Logger:           logger,
		Subs:             cache,
		Server:           conn,
		PoolConnection:   poolConn,
		PoolCommands:     poolComm,
		Poller:           poller,
		RedisMsgSub:      redisMsg,
		RedisSendInfo:    redisInfo,
		RedisCommandsSub: redisMain,
		Serializer:       s,
		AuthManager:      authManager,
		Redis:            red,
	}
	if err := red.Ping(red.Context()).Err(); err != nil {
		engine.Logger.Fatal().Err(err).Msg("can't' connect to redis")
	}

	//go engine.startRedisListen()
	//go engine.sendStatisticAboutUsers()
	//go engine.listeningCommands()
	engine.startupMessage([]byte(fmt.Sprintf("WS server: {%s} {%s} succesfully connected to hub", config.Name, config.Version)))
	return engine, nil
}

type redisMessage struct {
	CompanyName string `json:"company_name"`
	ClientId    string `json:"client_id"`
	Message     string `json:"message"`
} //"{\"company_name\":\"testing\",\"client_id\":\"4\",\"message\":\"10\"}"
type AddCompanyMessage struct {
	CompanyName string `json:"company_name"`
	MaxUsers    uint   `json:"max_users"`
	Duration    int    `json:"duration_hour"`
} //"{\"company_name\":\"testing\",\"max_users\":1000,\"duration_hour\":10}"
type CompanyToken struct {
	Token      string `json:"token"`
	ServerName string `json:"server_name"`
	Duration   int    `json:"duration_hour"`
}
type AddCompanyResponse struct {
	Token       CompanyToken `json:"company_token"`
	CompanyName string       `json:"company_name"`
}

func (e *Engine) startupMessage(msg []byte) error {
	return e.RedisSendInfo(msg)
}
func (e *Engine) listeningCommands() {
	for s := range e.RedisCommandsSub.Channel() {
		e.PoolCommands.Schedule(func() {
			var c AddCompanyMessage
			err := e.Serializer.Deserialize([]byte(s.Payload), &c)
			if err != nil {
				e.Logger.Warn().Err(err)
				return
			}
			token, err := e.AuthManager.NewJWT(c.CompanyName, time.Duration(c.Duration)*time.Hour)
			if err != nil {
				e.Logger.Warn().Err(err)
				return
			}
			err = e.Subs.AddCompany(c.CompanyName, c.MaxUsers, time.Duration(c.Duration))
			if err != nil {
				e.Logger.Warn().Err(err)
				return
			}
			resp := &AddCompanyResponse{
				CompanyName: c.CompanyName,
				Token: CompanyToken{
					ServerName: e.Conf.Name,
					Token:      token,
					Duration:   c.Duration,
				},
			}
			err = e.serializeAndSend(resp)
			if err != nil {
				e.Logger.Warn().Err(err)
				return
			}
		})
	}
}

func (e *Engine) serializeAndSend(v interface{}) error {
	payload, err := e.Serializer.Serialize(v)
	if err != nil {
		e.Logger.Warn().Err(err)
		return err
	}
	err = e.RedisSendInfo(payload)
	if err != nil {
		e.Logger.Warn().Err(err)
		return err
	}
	return nil
}
func (e *Engine) sendStatisticAboutUsers() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for range ticker.C {
		e.PoolCommands.Schedule(func() {
			companies := e.Subs.GetStatisctics()
			if companies == nil {
				return
			}
			err := e.serializeAndSend(companies)
			if err != nil {
				e.Logger.Warn().Err(err)
				return
			}
		})
	}
}
func (e *Engine) startRedisListen() {
	e.Logger.Info().Msg("redis listener was started")

	for msg := range e.RedisMsgSub.Channel() {
		e.PoolCommands.Schedule(func() {
			var m redisMessage
			err := e.Serializer.Deserialize([]byte(msg.Payload), &m)
			if err != nil {
				e.Logger.Warn().Err(err)
				return // TODO: Handle error
			}
			client, isOnline := e.Subs.GetClient(m.CompanyName, m.ClientId)
			if !isOnline {
				e.Logger.Warn().Err(err)
				return // TODO: Handle error
			}
			e.Logger.Printf("Message {%s} to client {%s} with company id {%s}", m.Message, m.ClientId, m.CompanyName)
			e.sendMsg(Message{client, []byte(msg.Payload)}, m.CompanyName)
		})
	}
}
func initRedis(address string) *redis.Client {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	return r
}

func (c *Client) HandleWrite(msg []byte, msgType ws.OpCode) error {
	// if c.Connection.IsClosed() {
	// 	return errors.New("connection was closed")
	// }
	err := c.Connection.Write(msg)
	if err != nil {
		return err
	}
	return nil
}
func (e *Engine) HandleRead(c *Client) ([]byte, bool, error) {
	payload, isControl, err := c.Connection.Read()
	if err != nil {
		return nil, isControl, errors.Wrap(err, "handle read error")
	}
	return payload, isControl, nil
}
func (e *Engine) Handle(conn net.Conn) {
	var userId, companyName string
	u := ws.Upgrader{
		ReadBufferSize:  256,
		WriteBufferSize: 1024,
		OnHeader: func(key, value []byte) error {
			if string(key) == "User" {
				userId = string(value)
			} else if string(key) == "Token" {
				var err error
				companyName, err = e.AuthManager.Parse(string(value))
				if err != nil {
					return ws.RejectConnectionError(
						ws.RejectionReason(fmt.Sprintf("bad token: %s", err)),
						ws.RejectionStatus(400))
				}
			}
			return nil
		},
	}
	_, err := u.Upgrade(conn)
	if err != nil {
		e.Logger.Warn().Err(err).Msg("upgrade error")
		_ = conn.Close()
		return
	}

	//var client *Client
	transport := websocket.NewWebsocketTransport(conn)

	readDescriptor, _ := netpoll.Handle(conn, netpoll.EventRead) //netpoll.Must(netpoll.HandleRead(conn))

	client := NewClient(transport, userId, readDescriptor)
	err, deleteFn := e.Subs.AddClient(companyName, client)
	if err != nil {
		e.Logger.Info().Err(err)
		conn.Write([]byte("User already exists"))
		conn.Close()
		return
	}
	closeAndDel := func(cl *Client) {
		e.Logger.Info().Msgf("user with id={%s} was disconnected\n", client.UserId)
		deleteFn()
	}
	e.Logger.Info().Msgf("user connected with id={%s} and {%s}\n", client.UserId, companyName)

	// go func() {
	// write:
	// 	for {
	// 		select {
	// 		case <-client.Connection.CloseCh():
	// 			//deleteFn()
	// 			break write
	// 		case <-client.Connection.Timer.C:
	// 			fmt.Println("timer has expired")
	// 			client.Connection.Close()

	// 			break write

	// 		case msg := <-client.MessageChan:
	// 			e.PoolCommands.Schedule(func() {
	// 				if err := client.HandleWrite([]byte(msg), ws.OpText); err != nil {
	// 					//e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
	// 					//deleteFn()
	// 					//client.Connection.Close()

	// 				} else {
	// 					e.Logger.Printf("Message was send to user={%s}\n", client.UserId)
	// 				}
	// 			})
	// 		//default:
	// 		}
	// 	}

	//  }()

	//})
	//})
	go func() {
		_ = e.Poller.Start(readDescriptor, func(ev netpoll.Event) {

			if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
				_ = e.Poller.Stop(readDescriptor)
				closeAndDel(client)
				e.Logger.Info().Msgf("user with id={%s} was disconnected", client.UserId)
				e.Logger.Debug().Msg(ev.String())
				return
			}
			payload, isControl, err := e.HandleRead(client)
			if err != nil {
				_ = e.Poller.Stop(readDescriptor)
				e.Logger.Debug().Err(err)
				return
			} else {
				if !isControl {
					if string(payload) == "test" {
						client.HandleWrite([]byte{}, ws.OpPing)
					}
					e.sendMsg(Message{Client: client, Msg: []byte(strings.ToUpper(string(payload)))}, companyName)
					e.Logger.Printf("Message from user={%s}: {%s}\n", client.UserId, payload)
				}
				_ = e.Poller.Resume(readDescriptor)
			}
		})
	}()

}

func (e *Engine) SaveState() error {
	stats := e.Subs.GetStatisctics()
	data, err := e.Serializer.Serialize(&stats)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return e.Redis.Set(ctx, "save", data, 0).Err()
}

func (e *Engine) RestoreState() error {
	ctx := context.Background()
	val, err := e.Redis.Get(ctx, "save").Result()
	if err == redis.Nil {
		return errors.New("key doesn't exists")
	}
	var stats []CompanyStats
	err = e.Serializer.Deserialize([]byte(val), &stats)
	if err != nil {
		return err
	}
	for _, c := range stats {
		if c.Expired == true {
			continue
		}
		ttl := c.Stopped.Sub(c.Time)
		e.Logger.Debug().Msgf("restored - %s ttl - %s", c.Name, c.TTL-ttl)

		e.Subs.AddCompany(c.Name, c.MaxUsers, c.TTL-ttl)
	}
	return nil
}

func (e *Engine) listenCommands() {
	for {
		command, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		switch {
		case command == "help":
			fmt.Println("There will be help info")
		// case command[:4] == "add":
		// 	splited := strings.Split(command, " ")
		// 	if len(splited) > 3 {
		// 		fmt.Println("Wrong add command")
		// 	} else {

		// 	}
		 }
	}
}

func (e *Engine) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	users, _ := e.Subs.GetActiveUsers("test")
	_, _ = w.Write([]byte(fmt.Sprintf("%d", users)))
	e.Logger.Printf("Active users = %d", users)
}
func (e *Engine) sendMsg(msg Message, compName string) bool {

	// if client.MessageChan == nil {
	// 	return false
	// }
	e.Subs.SendMessage(msg, compName)
	//client.MessageChan <- msg
	//client.HandleWrite([]byte(msg), ws.OpText)
	return true
}
func (e *Engine) SendToClientById(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("user-id")
	companyName := r.Header.Get("company-name")
	payload := r.Header.Get("payload")
	if userId == "" || companyName == "" || payload == "" {
		printError(w, "Please set user-id, company-name, payload as header", http.StatusBadRequest)
		return
	}
	client, ok := e.Subs.GetClient(companyName, userId)
	if !ok {
		printError(w, "User offline", http.StatusBadRequest)
		return
	}
	ms := Message{client, []byte(payload)}
	ok = e.sendMsg(ms, companyName)
}
func printError(w http.ResponseWriter, msg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	str := fmt.Sprintf(`{"Error":"%s"}`, msg)
	fmt.Fprint(w, str)
}
