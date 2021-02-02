package core

import (
	"errors"
	"sync"
	"sync/atomic"
)


type ClientHub struct {
	mu    sync.RWMutex
	Users map[string]*Client
	activeUsers *uint64
	maxUsers uint
}

func NewClientHub(maxUsers uint) *ClientHub {
	return &ClientHub{
		Users: make(map[string]*Client),
		activeUsers: new(uint64),
		maxUsers: maxUsers,
	}
}
func (c *ClientHub) AddClient(client *Client) error {
	if ok := c.verifyClient(client.UserId); ok {
		return errors.New("client already exists")
	}
	if c.GetNumActiveUsers() >= uint64(c.maxUsers){
		return errors.New("reached limit of max active users")
	}
	c.mu.Lock()
	c.Users[client.UserId] = client
	c.mu.Unlock()
	atomic.AddUint64(c.activeUsers, 1)
	return nil
}
func (c *ClientHub) verifyClient(userId string) bool {
	c.mu.RLock()
	_, ok := c.Get(userId)
	c.mu.RUnlock()
	return ok
}
func (c *ClientHub) DeleteClient(userId string) error {
	if ok := c.verifyClient(userId); !ok {
		return errors.New("client doesn't exists")
	}
	c.mu.Lock()
	delete(c.Users,userId)
	c.mu.Unlock()
	atomic.AddUint64(c.activeUsers, 0)
	return nil
}
func (c *ClientHub) GetNumActiveUsers() uint64 {
	return atomic.LoadUint64(c.activeUsers)
}
func (c *ClientHub) GetUsers() []*Client {
	var cl []*Client
	c.mu.RLock()
	for _,client := range c.Users{
		cl = append(cl,client)
	}
	c.mu.RUnlock()
	return cl
}
func (c *ClientHub) Get(userId string) (*Client, bool) {
	c.mu.RLock()
	client, ok := c.Users[userId]
	c.mu.RUnlock()
	return client, ok
}
