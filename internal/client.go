package internal


// Client is a representation of user.
type Client struct {
	UserId            string
	Connection        Websocket
}


// NewClient creates a pointer to Client object.
func NewClient(ws Websocket, userId string) *Client {
	return &Client{
		Connection:        ws,
		UserId:            userId,
	}
}

// Disconnect closes the connection.
func (c *Client) Disconnect() error {
	return c.Connection.Close()
}

