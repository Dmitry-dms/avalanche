package core

import (
	"log"
	"net/http"

	"github.com/Dmitry-dms/avalanche/pkg/websocket"
	"github.com/gobwas/ws"
)

type Engine struct {
	Conf       Config
	Logger     *log.Logger
	Subs       *companyHub
	MsgChannel chan string
}

func NewEngine(config Config, logger *log.Logger) *Engine {
	return &Engine{
		Conf:       config,
		Logger:     logger,
		Subs:       newCompanyHub(),
		MsgChannel: make(chan string),
	}
}

func (e *Engine) HandleWrite(c *Client) {
	defer c.Connection.Close()
	defer e.Logger.Println("connection was closed write")
	for {
		if c.Connection.IsClosed(){
			break
		}
		err := c.Connection.Write(e.MsgChannel)
		if err != nil {
			e.Logger.Println(err)
			break
		}
	}
}
func (e *Engine) HandleRead(c *Client) {
	defer c.Connection.Close()
	defer e.Logger.Println("connection was closed read")
	for {
		payload, _, err := c.Connection.Read()
		if err != nil {
			e.Logger.Println(err)
			break
		}
		e.Logger.Printf("Meesage {%s}", payload)
	}
}

func (e *Engine) HandleClient(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	defer func() {
		e.Logger.Println("connection was closed")
		_ = conn.Close()
	}()
	if err != nil {
		e.Logger.Println(err)
	}
	transport := websocket.NewWebsocketTransport(conn)
	client := NewClient(transport, "test-user")
	err = e.Subs.GetCompanyClientHub("test").AddClient(client)
	if err != nil {
		e.Logger.Println(err)
	}
	go e.HandleWrite(client)
	e.HandleRead(client)
}
