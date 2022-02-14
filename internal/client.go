package internal

import (
	"github.com/Dmitry-dms/avalanche/pkg/websocket"
	"github.com/mailru/easygo/netpoll"
)

type Client struct {
	UserId      string
	Connection  websocket.Websocket
	Desc        *netpoll.Desc
}
type CloseFunc func() error

func NewClient(transport websocket.Websocket, userId string, desc *netpoll.Desc) *Client {
	return &Client{
		Connection:  transport,
		UserId:      userId,
		Desc:        desc,
	}
}

func (c *Client) Disconnect() error{
	return c.Connection.Close()
}

