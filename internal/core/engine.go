package core

import (
	"errors"
	"fmt"
	"time"

	//"sync"

	"log"
	"net"
	"net/http"

	"github.com/Dmitry-dms/avalanche/pkg/websocket"
	"github.com/mailru/easygo/netpoll"
	"github.com/panjf2000/ants/v2"

	"github.com/gobwas/httphead"
	"github.com/gobwas/ws"
)

type Engine struct {
	Conf   Config
	Logger *log.Logger
	Subs   AvalacnheCache
	Server net.Listener
	Pool   *ants.Pool
	Poller netpoll.Poller
}

func NewEngine(config Config, logger *log.Logger, cache AvalacnheCache, conn net.Listener, pool *ants.Pool, poller netpoll.Poller) *Engine {
	return &Engine{
		Conf:   config,
		Logger: logger,
		Subs:   cache,
		Server: conn,
		Pool:   pool,
		Poller: poller,
	}
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
	start:=time.Now()
	var userId, companyName string
	companyName = "test"
	u := ws.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 512,
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
		ws.RejectConnectionError(ws.RejectionReason("user already exists"), ws.RejectionStatus(400))
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
		e.Pool.Submit(func() {
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
			select{
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
				if err!= nil {
					e.Logger.Printf("Error from read shedule^ %s", err.Error())
				}
			}
			
		}
	}()
}

func (e *Engine) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	e.Logger.Println(w.Write([]byte(fmt.Sprintf("%d", e.Subs.GetActiveUsers()))))
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
