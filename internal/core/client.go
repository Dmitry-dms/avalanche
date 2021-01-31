package core

import (
	"github.com/Dmitry-dms/websockets/pkg/websocket"
)

type Client struct {
	MessageChan chan string
	Connection  *websocket.CustomWebsocketTransport
}

func NewClient(transport *websocket.CustomWebsocketTransport) *Client {
	return &Client{
		MessageChan: make(chan string, 1),
		Connection:  transport,
	}
}
func (c *Client) isClosed() bool {
	return c.Connection.IsClosed()
}
func (c *Client) getCh() chan struct{} {
	return c.Connection.CloseCh()
}
