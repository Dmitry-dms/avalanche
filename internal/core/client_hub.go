package core

import "sync"

type ClientHub struct {
	mu    sync.Mutex
	Users map[string]*Client
}

func NewHub() *ClientHub {
	return &ClientHub{
		Users: make(map[string]*Client),
	}
}
func (c *ClientHub) addClient(id string, client *Client) error {
	c.mu.Lock()
	c.Users[id] = client
	c.mu.Unlock()
	return nil
}
func (c *ClientHub) Get(id string) *Client {
	return c.Users[id]
}