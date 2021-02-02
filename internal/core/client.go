package core

import (
	"github.com/Dmitry-dms/avalanche/pkg/websocket"
)

type Client struct {
	UserId string
	MessageChan chan string
	Connection  *websocket.CustomWebsocketTransport
}

func NewClient(transport *websocket.CustomWebsocketTransport, userId string) *Client {
	return &Client{
		MessageChan: make(chan string, 1),
		Connection:  transport,
		UserId: userId,
	}
}
func (c *Client) isClosed() bool {
	return c.Connection.IsClosed()
}
func (c *Client) getCh() chan struct{} {
	return c.Connection.CloseCh()
}
