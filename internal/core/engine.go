package core

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
	"github.com/Dmitry-dms/avalanche/pkg/serializer"
	"github.com/Dmitry-dms/avalanche/pkg/websocket"
	"github.com/go-redis/redis/v8"
	"github.com/mailru/easygo/netpoll"
	"github.com/panjf2000/ants/v2"

	"github.com/gobwas/httphead"
	"github.com/gobwas/ws"
)

type Engine struct {
	Conf             Config
	Logger           *log.Logger
	Subs             Cache
	Server           net.Listener
	Pool             *ants.Pool
	Poller           netpoll.Poller
	RedisMsgSub      *redis.PubSub
	RedisSendInfo    func(payload []byte) error
	RedisAddCompany  func(company addCompanyResponse) error
	RedisCommandsSub *redis.PubSub
	Serializer       serializer.AvalancheSerializer
	AuthManager      *auth.Manager
}

func NewEngine(config Config, logger *log.Logger, cache Cache,
	conn net.Listener, pool *ants.Pool,
	poller netpoll.Poller, s serializer.AvalancheSerializer) (*Engine, error) {
	red := initRedis(config.RedisAddress)
	redisMsg := red.Subscribe(context.Background(), config.RedisMsgPrefix+config.Name)
	redisInfo := func(payload []byte) error {
		return red.Publish(context.Background(), config.RedisInfoPrefix+config.Name, payload).Err()
	}
	redisAddCompany := func(company addCompanyResponse) error {
		return red.Append(context.Background(), company.companyName, company.toString()).Err()
	}
	redisMain := red.Subscribe(context.Background(), config.RedisCommandsPrefix+config.Name)
	authManager, err := auth.NewManager(config.AuthJWTkey)
	if err == nil {
		return nil, err
	}
	engine := &Engine{
		Conf:             config,
		Logger:           logger,
		Subs:             cache,
		Server:           conn,
		Pool:             pool,
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
	return engine, nil
}

type redisMessage struct {
	companyName string
	ClientId    string
	Message     string
}
type addCompanyMessage struct {
	CompanyName string
	MaxUsers    uint
	Duration    int
}
type companyToken struct {
	Token      string
	ServerName string
	Duration   int
}
type addCompanyResponse struct {
	token       companyToken
	companyName string
}

func (r *addCompanyResponse) toString() string {
	return fmt.Sprintf("%s:%s", r.token.ServerName, r.token.Token)
}
func (e *Engine) listeningCommands() {
	for s := range e.RedisCommandsSub.Channel() {
		_ = e.Pool.Submit(func() {
			var c addCompanyMessage
			err := e.Serializer.Deserialize([]byte(s.Payload), c)
			if err != nil {
				e.Logger.Println(err)
				return
			}
			token, err := e.AuthManager.NewJWT(c.CompanyName, time.Duration(c.Duration))
			if err != nil {
				e.Logger.Println(err)
				return
			}
			err = e.Subs.AddCompany(c.CompanyName, token, c.MaxUsers, time.Duration(c.Duration))
			if err != nil {
				e.Logger.Println(err)
				return
			}
			err = e.RedisAddCompany(addCompanyResponse{
				companyName: c.CompanyName,
				token: companyToken{
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
		_ = e.Pool.Submit(func() {
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
		_ = e.Pool.Submit(func() {
			var m redisMessage
			err := e.Serializer.Deserialize([]byte(msg.Payload), &m)
			if err != nil {
				e.Logger.Println(err)
				return // TODO: Handle error
			}
			client, isOnline := e.Subs.GetClient(m.companyName, m.ClientId)
			if !isOnline {
				e.Logger.Println(err)
				return // TODO: Handle error
			}
			e.Logger.Printf("Message {%s} to client {%s} with company id {%s}", m.Message, m.ClientId, m.companyName)
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
	start := time.Now()
	var userId, companyName string
	//companyName = "test"
	u := ws.Upgrader{
		ReadBufferSize:  256,
		WriteBufferSize: 1024,
		OnHeader: func(key, value []byte) error {
			if string(key) != "Cookie" {
				return nil
			}
			ok := httphead.ScanCookie(value, func(key, value []byte) bool {
				userId = string(key)
				return true
			})
			if ok {
				return nil
			}
			return ws.RejectConnectionError(
				ws.RejectionReason("bad cookie"),
				ws.RejectionStatus(400),
			)

		}, Extension: func(op httphead.Option) bool {
			token, ok := op.Parameters.Get("token")
			err := errors.New("")
			companyName, err = e.AuthManager.Parse(string(token))
			if err != nil {
				return !ok
			}
			return ok
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
		_ = ws.RejectConnectionError(ws.RejectionReason("user already exists"), ws.RejectionStatus(400))
		return
	}
	//e.Logger.Printf("User with id={%s} connected\n", client.UserId)
	e.Logger.Printf("TAKEN TIME = {%s}", time.Since(start))
	readDescriptor := netpoll.Must(netpoll.HandleReadOnce(conn))

	_ = e.Poller.Start(readDescriptor, func(ev netpoll.Event) {

		if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			_ = e.Poller.Stop(readDescriptor)
			_ = deleteFn()
			client.Connection.Close()
			e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
			e.Logger.Println("read poller error")
			return
		}
		_ = e.Pool.Submit(func() {
			if payload, isControl, err := e.HandleRead(client); err != nil {
				_ = e.Poller.Stop(readDescriptor)
				e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
				_ = deleteFn()
				client.Connection.Close()
				return
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
				_ = deleteFn()
				return
			case msg := <-client.MessageChan:
				err := e.Pool.Submit(func() {
					if err := client.HandleWrite([]byte(msg)); err != nil {
						e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
						_ = deleteFn()
						client.Connection.Close()
						return
					} else {
						e.Logger.Printf("Meesage was send to user={%s}\n", client.UserId)
					}
				})
				if err != nil {
					e.Logger.Printf("Error from read shedule^ %s", err.Error())
				}
			}

		}
	}()
}

func (e *Engine) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	users, _ := e.Subs.GetActiveUsers("test")
	_, _ = w.Write([]byte(fmt.Sprintf("%d", users)))
	e.Logger.Printf("Active users = %d", users)
}
func (e *Engine) SendToClientById(w http.ResponseWriter, r *http.Request) {
	userId := "0"         //r.Header.Get("user-id")
	companyName := "test" //r.Header.Get("company-name")
	payload := "hello!!!" //r.Header.Get("payload")
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
