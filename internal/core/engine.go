package core

import (
	"fmt"
	//"sync"

	"log"
	"net"
	"net/http"

	"github.com/Dmitry-dms/avalanche/pkg/websocket"

	"github.com/gobwas/ws"
)

type Engine struct {
	Conf       Config
	Logger     *log.Logger
	Subs       AvalacnheCache
	//MsgChannel chan string
	Server     net.Listener
	Pool       *websocket.Pool
}

func NewEngine(config Config, logger *log.Logger, cache AvalacnheCache, conn net.Listener, pool *websocket.Pool) *Engine {
	return &Engine{
		Conf:       config,
		Logger:     logger,
		Subs:       cache,
		//MsgChannel: make(chan string),
		Server:     conn,
		Pool:       pool,
	}
}

func (e *Engine) HandleWrite(c *Client) {
	for {
		if c.Connection.IsClosed() {
			return
		}
		err := c.Connection.Write(c.MessageChan)
		if err != nil {
			e.Logger.Println(err)
			return
		}
		e.Logger.Printf("Meesage was send to user={%s}\n", c.UserId)
	}
}
func (e *Engine) HandleRead(c *Client) {
	for {
		payload, _, err := c.Connection.Read()
		if err != nil {
			e.Logger.Println(err)
			return
		}
		e.Logger.Printf("Meesage from user={%s}: {%s}\n", c.UserId, payload)
	}

}
func (e *Engine) SubscribeClient(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("user_id")
	companyName := r.Header.Get("company_name")
	if userId == "" || companyName == "" {
		printError(w, "Please set user-id, company-name as header", http.StatusBadRequest)
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		e.Logger.Println(err)
	}
	var client *Client
	transport := websocket.NewWebsocketTransport(conn)
	client = NewClient(transport, userId)
	err, deleteFn := e.Subs.AddClient("test", client)
	if err != nil {
		e.Logger.Println(err)
		ws.RejectConnectionError(ws.RejectionReason("user already exists"), ws.RejectionStatus(400))
		return
	}
	e.Logger.Printf("User with id={%s} connected\n", client.UserId)

	go func() {
		defer func() {
			err = deleteFn()
			if err != nil {
				e.Logger.Printf("client with id={%s} doesn't exists", client.UserId)
			}
			_ = conn.Close()
			e.Logger.Printf("User with id={%s} was disconnected\n", client.UserId)
		}()
		go e.HandleWrite(client)
		e.HandleRead(client)
	}()
}
func (e *Engine) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	e.Logger.Println(w.Write([]byte(fmt.Sprintf("%d", e.Subs.GetActiveUsers()))))
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
