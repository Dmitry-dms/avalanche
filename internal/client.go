package internal

import (
	"github.com/Dmitry-dms/avalanche/pkg/websocket"
)

// Client is a representation of user.
type Client struct {
	UserId            string
	Connection        websocket.Websocket
}

//type CloseFunc func() error

// NewClient creates a Client object.
func NewClient(transport websocket.Websocket, userId string) *Client {
	return &Client{
		Connection:        transport,
		UserId:            userId,
	}
}

// Disconnect is responsible for closing the connection.
func (c *Client) Disconnect() error {
	return c.Connection.Close()
}

