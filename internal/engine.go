package internal

import (
	"context"
	"strings"

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
	"github.com/pkg/errors"

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
	}
	if err := red.Ping(red.Context()).Err(); err != nil {
		engine.Logger.Println(err)
	}
	go engine.startRedisListen()
	go engine.sendStatisticAboutUsers()
	go engine.listeningCommands()
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

type mongoMessage struct {
	FromId  string `json:"from"`
	ToId    string `json:"to"`
	Payload string `json:"payload"`
	IsRead  bool   `json:"is_read"`
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
				e.Logger.Println(err)
				return
			}
		})
	}
}

func (e *Engine) sendMongoMsg(from, to, msg string, read bool) error {
	mongoMsg := mongoMessage{
		FromId:  from,
		ToId:    to,
		Payload: msg,
		IsRead:  read,
	}
	err := e.serializeAndSend(mongoMsg)
	return err
}

func (e *Engine) serializeAndSend(v interface{}) error {
	payload, err := e.Serializer.Serialize(v)
	if err != nil {
		e.Logger.Println(err)
		return err
	}
	err = e.RedisSendInfo(payload)
	if err != nil {
		e.Logger.Println(err)
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

func (c *Client) HandleWrite(msg []byte, msgType ws.OpCode) error {
	if c.Connection.IsClosed() {
		return errors.New("connection was closed")
	}
	err := c.Connection.Write(msg, msgType)
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
	//companyName = "test"
	//userId = "user"
	u := ws.Upgrader{
		ReadBufferSize:  256,
		WriteBufferSize: 1024,
		OnHeader: func(key, value []byte) error {
			//e.Logger.Printf("Key =%s, value = %s", key, value)
			// if string(key) == "User" {
			// 	userId = string(value)
			// }
			// if userId == "" {
			// 	return ws.RejectConnectionError(
			// 		ws.RejectionReason("UserID is empty"),
			// 		ws.RejectionStatus(400))
			// }
			// if string(key) == "Token" {
			// 	var err error
			// 	companyName, err = e.AuthManager.Parse(string(value))
			// 	if err != nil {
			// 		return ws.RejectConnectionError(
			// 			ws.RejectionReason(fmt.Sprintf("bad token: %s", err)),
			// 			ws.RejectionStatus(400))
			// 	}
			// }

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
			// } else {

			// }
			return nil
		},
	}
	_, err := u.Upgrade(conn)
	if err != nil {
		log.Printf("%s: upgrade error: %v", conn, err)
		_ = conn.Close()
		return
	}

	var client *Client
	transport := websocket.NewWebsocketTransport(conn, time.Second*20)
	client = NewClient(transport, userId)
	err, deleteFn := e.Subs.AddClient(companyName, client)
	if err != nil {
		e.Logger.Println(err)
		conn.Close()
		conn.Write([]byte("User already exists"))
		return
	}
	closeAndDel := func(cl *Client) {
		e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
		deleteFn()
	}

	e.Logger.Printf("User connected with id={%s} and {%s}\n", client.UserId, companyName)
	//	readDescriptor := netpoll.Must(netpoll.HandleReadOnce(conn))
	//_ = e.Poller.Start(readDescriptor, func(ev netpoll.Event) {

	//if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
	//	_ = e.Poller.Stop(readDescriptor)
	//	deleteFn()
	//client.Connection.Close()
	//e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
	//e.Logger.Println("read poller error")
	//return
	//}

	//e.PoolCommands.Schedule(func() {
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

	// }()
	//})
	//})
	//go func() {

	//}()
	for {
		payload, isControl, err := e.HandleRead(client)
		if err != nil {
			//fmt.Println("error read")
			//_ = e.Poller.Stop(readDescriptor)
			//e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
			//deleteFn()
			//client.Connection.Close()
			closeAndDel(client)
			return
		} else {
			if !isControl {
				if string(payload) == "test" {
					client.HandleWrite([]byte{}, ws.OpPing)
					continue
				}
				client.MessageChan <- strings.ToUpper(string(payload))
				e.Logger.Printf("Message from user={%s}: {%s}\n", client.UserId, payload)
			}
			//_ = e.Poller.Resume(readDescriptor)
		}
	}
}

func (e *Engine) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	users, _ := e.Subs.GetActiveUsers("test")
	_, _ = w.Write([]byte(fmt.Sprintf("%d", users)))
	e.Logger.Printf("Active users = %d", users)
}
func (e *Engine) sendMsg(client *Client, msg string) bool {

	if client.MessageChan == nil {
		return false
	}
	client.MessageChan <- msg
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
	ok = e.sendMsg(client, payload)
}
func printError(w http.ResponseWriter, msg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	str := fmt.Sprintf(`{"Error":"%s"}`, msg)
	fmt.Fprint(w, str)
}
