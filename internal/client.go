package internal

import (
	"github.com/Dmitry-dms/avalanche/pkg/websocket"
)

type Client struct {
	UserId            string
	Connection        websocket.Websocket
}
type CloseFunc func() error

func NewClient(transport websocket.Websocket, userId string) *Client {
	return &Client{
		Connection:        transport,
		UserId:            userId,
	}
}

func (c *Client) Disconnect() error {
	return c.Connection.Close()
}

