package core

import (
	"log"
	"net/http"

	"github.com/Dmitry-dms/websockets/pkg/websocket"
	"github.com/gobwas/ws"
)

type Engine struct {
	Conf       Config
	Logger     *log.Logger
	Subs       *ClientHub
	MsgChannel chan string
}

func NewEngine(config Config, logger *log.Logger, hub *ClientHub) *Engine {
	return &Engine{
		Conf:       config,
		Logger:     logger,
		Subs:       hub,
		MsgChannel: make(chan string),
	}
}

func (e *Engine) HandleWrite(c *Client) {
	defer c.Connection.Close()
	defer e.Logger.Println("connection was closed write")
	for {
		// if c.isClosed() {
		// 	return
		// }

		msg := <-e.MsgChannel
		err := c.Connection.Write([]byte(msg))
		if err != nil {
			e.Logger.Fatal(err)
			return
		}

	}
}
func (e *Engine) HandleRead(c *Client) {
	defer c.Connection.Close()
	defer e.Logger.Println("connection was closed read")
	for {
		// if c.isClosed() {
		// 	return
		// }

		payload, _, err := c.Connection.Read()
		if err != nil {
			e.Logger.Fatal(err)
			e.Logger.Println("handle read err")
			//c.Connection.Closed=true
			return
		}
		e.Logger.Printf("Meesage {%s}", payload)

	}
}

func (e *Engine) HandleClient(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		e.Logger.Fatal(err)
	}
	transport := websocket.NewWebsocketTransport(conn)
	client := NewClient(transport)
	err = e.Subs.addClient("test", client)
	if err != nil {
		e.Logger.Fatal(err)
	}
	go e.HandleWrite(client)
	go e.HandleRead(client)
}
