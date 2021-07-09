package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	//"sync"

	"log"
	"net"
	"net/http"

	"github.com/Dmitry-dms/avalanche/pkg/auth"
	"github.com/Dmitry-dms/avalanche/pkg/pool"
	"github.com/Dmitry-dms/avalanche/pkg/serializer"
	"github.com/Dmitry-dms/avalanche/pkg/websocket"

	"github.com/go-redis/redis/v8"
	"github.com/mailru/easygo/netpoll"

	"github.com/gobwas/ws"
)

type Engine struct {
	Context          context.Context
	Conf             Config
	Logger           *log.Logger
	Subs             Cache
	Server           net.Listener
	PoolConnection   *pool.Pool
	PoolCommands     *pool.Pool
	Poller           netpoll.Poller
	RedisMsgSub      *redis.PubSub
	RedisSendInfo    func(payload []byte) error
	RedisAddCompany  func(company addCompanyResponse) error
	RedisCommandsSub *redis.PubSub
	Serializer       serializer.AvalancheSerializer
	AuthManager      *auth.Manager
}

func NewEngine(ctx context.Context, config Config, logger *log.Logger, cache Cache,
	conn net.Listener, poolConn *pool.Pool, poolComm *pool.Pool,
	poller netpoll.Poller, s serializer.AvalancheSerializer) (*Engine, error) {

	red := initRedis(config.RedisAddress)

	redisInfo := func(payload []byte) error {
		return red.Publish(ctx, config.RedisInfoPrefix, payload).Err() //+config.Name
	}
	redisAddCompany := func(company addCompanyResponse) error {
		return red.Append(ctx, company.CompanyName, company.toString()).Err()
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
		RedisAddCompany:  redisAddCompany,
		Serializer:       s,
		AuthManager:      authManager,
	}
	if err := red.Ping(red.Context()).Err(); err != nil {
		engine.Logger.Println(err)
	}
	go engine.startRedisListen()
	go engine.sendStatisticAboutUsers()
	go engine.listeningCommands()
	engine.startupMessage([]byte("WS server succesfully connected to hub"))
	return engine, nil
}

type redisMessage struct {
	CompanyName string `json:"company_name"`
	ClientId    string `json:"client_id"`
	Message     string `json:"message"`
}//"{\"company_name\":\"testing\",\"client_id\":\"4\",\"message\":\"10\"}"
type AddCompanyMessage struct {
	CompanyName string `json:"company_name"`
	MaxUsers    uint   `json:"max_users"`
	Duration    int    `json:"duration_hour"`
} //"{\"company_name\":\"testing\",\"max_users\":1000,\"duration_hour\":10}"
type companyToken struct {
	Token      string `json:"token"`
	ServerName string `json:"server_name"`
	Duration   int    `json:"duration_hour"`
}
type addCompanyResponse struct {
	Token       companyToken `json:"company_token"`
	CompanyName string       `json:"company_name"`
}

func (r *addCompanyResponse) toString() string {
	return fmt.Sprintf("%s:%s", r.Token.ServerName, r.Token.Token)
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
				e.Logger.Println(err)
				return
			}
			e.Logger.Println(c)

			token, err := e.AuthManager.NewJWT(c.CompanyName, time.Duration(c.Duration)*time.Hour)
			if err != nil {
				e.Logger.Println(err)
				return
			}
			e.Logger.Println(token)
			err = e.Subs.AddCompany(c.CompanyName, token, c.MaxUsers, time.Duration(c.Duration))
			if err != nil {
				e.Logger.Println(err)
				return
			}
			err = e.RedisAddCompany(addCompanyResponse{
				CompanyName: c.CompanyName,
				Token: companyToken{
					ServerName: e.Conf.Name,
					Token:      token,
				},
			})
			if err != nil {
				e.Logger.Println(err)
				return
			}
		})
	}
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
			serialized, err := e.Serializer.Serialize(companies)
			if err != nil {
				e.Logger.Println(err)
				return
			}
			err = e.RedisSendInfo(serialized)
			if err != nil {
				e.Logger.Println(err)
				return
			}
		})
	}
}
func (e *Engine) startRedisListen() {
	e.Logger.Println("listener was started")
	for msg := range e.RedisMsgSub.Channel() {
		e.PoolCommands.Schedule(func() {
			var m redisMessage
			err := e.Serializer.Deserialize([]byte(msg.Payload), &m)
			if err != nil {
				e.Logger.Println(err)
				return // TODO: Handle error
			}
			client, isOnline := e.Subs.GetClient(m.CompanyName, m.ClientId)
			if !isOnline {
				e.Logger.Println(err)
				return // TODO: Handle error
			}
			e.Logger.Printf("Message {%s} to client {%s} with company id {%s}", m.Message, m.ClientId, m.CompanyName)
			client.MessageChan <- m.Message
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

func (c *Client) HandleWrite(msg []byte) error {
	if c.Connection.IsClosed() {
		return errors.New("connection was closed")
	}
	err := c.Connection.Write(msg)
	if err != nil {
		return err
	}
	return nil
}
func (e *Engine) HandleRead(c *Client) ([]byte, bool, error) {
	payload, isControl, err := c.Connection.Read()
	if err != nil {
		e.Logger.Println(err)
		return nil, isControl, err
	}
	return payload, isControl, err
}
func (e *Engine) Handle(conn net.Conn) {

	var userId, companyName string
	u := ws.Upgrader{
		ReadBufferSize:  256,
		WriteBufferSize: 1024,
		OnHeader: func(key, value []byte) error {
			if string(key) == "User" {
				userId = string(value)
			}
			if string(key) == "Token" {
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
	// Zero-copy upgrade to WebSocket connection.
	_, err := u.Upgrade(conn)
	if err != nil {
		log.Printf("%s: upgrade error: %v", conn, err)
		_ = conn.Close()
		return
	}

	var client *Client
	transport := websocket.NewWebsocketTransport(conn)
	client = NewClient(transport, userId)
	err, deleteFn := e.Subs.AddClient(companyName, client)
	if err != nil {
		e.Logger.Println(err)
		conn.Close()
		_ = ws.RejectConnectionError(ws.RejectionReason("user already exists"), ws.RejectionStatus(400))
		return
	}

	readDescriptor := netpoll.Must(netpoll.HandleReadOnce(conn))
	_ = e.Poller.Start(readDescriptor, func(ev netpoll.Event) {

		if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			_ = e.Poller.Stop(readDescriptor)
			deleteFn()
			client.Connection.Close()
			e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
			//e.Logger.Println("read poller error")
			return
		}

		e.PoolCommands.Schedule(func() {

			if payload, isControl, err := e.HandleRead(client); err != nil {
				_ = e.Poller.Stop(readDescriptor)
				e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
				deleteFn()
				client.Connection.Close()

			} else {

				if !isControl {
					e.Logger.Printf("Meesage from user={%s}: {%s}\n", client.UserId, payload)
				}
				_ = e.Poller.Resume(readDescriptor)

			}
		})
	})
	go func() {
		for {
			select {
			case <-client.Connection.CloseCh():
				deleteFn()
				e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
				return
			case msg := <-client.MessageChan:
				e.PoolCommands.Schedule(func() {
					if err := client.HandleWrite([]byte(msg)); err != nil {
						e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
						deleteFn()
						client.Connection.Close()
						return
					} else {
						e.Logger.Printf("Meesage was send to user={%s}\n", client.UserId)
					}
				})
			}
			time.Sleep(5 * time.Second)
		}
	}()
	
	
	
}

func (e *Engine) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	users, _ := e.Subs.GetActiveUsers("test")
	_, _ = w.Write([]byte(fmt.Sprintf("%d", users)))
	e.Logger.Printf("Active users = %d", users)
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
		printError(w, "client doesn't exists", http.StatusInternalServerError)
		return
	}
	if client.MessageChan == nil {
		printError(w, "client's channel was deleted", http.StatusInternalServerError)
		return
	}
	client.MessageChan <- payload
}
func printError(w http.ResponseWriter, msg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	str := fmt.Sprintf(`{"Error":"%s"}`, msg)
	fmt.Fprint(w, str)
}
