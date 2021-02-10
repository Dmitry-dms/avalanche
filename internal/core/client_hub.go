package core

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	userExists  = "client already exists"
	limitUsers  = "reached the limit of max active users"
	userDExists = "client doesn't exists"
)

type ClientHub struct {
	mu          sync.RWMutex
	Users       map[string]*Client
	activeUsers *uint64
	maxUsers    uint
}

func newClientHub(maxUsers uint) *ClientHub {
	return &ClientHub{
		Users:       make(map[string]*Client),
		activeUsers: new(uint64),
		maxUsers:    maxUsers,
	}
}
func (c *ClientHub) AddClient(client *Client) error {
	if ok := c.verifyClient(client.UserId); ok {
		return errors.New(userExists)
	}
	if c.GetNumActiveUsers() >= uint64(c.maxUsers) {
		return errors.New(limitUsers)
	}
	c.mu.Lock()
	c.Users[client.UserId] = client
	c.mu.Unlock()
	atomic.AddUint64(c.activeUsers, 1)
	return nil
}
func (c *ClientHub) verifyClient(userId string) bool {
	c.mu.RLock()
	_, ok := c.get(userId)
	c.mu.RUnlock()
	return ok
}
func (c *ClientHub) DeleteClient(userId string) error {
	if ok := c.verifyClient(userId); !ok {
		return errors.New(userDExists)
	}
	c.mu.Lock()
	delete(c.Users, userId)
	c.mu.Unlock()
	atomic.AddUint64(c.activeUsers, ^uint64(0))
	return nil
}
func (c *ClientHub) GetNumActiveUsers() uint64 {
	return atomic.LoadUint64(c.activeUsers)
}
func (c *ClientHub) GetUsers() []*Client {
	var cl []*Client
	c.mu.RLock()
	for _, client := range c.Users {
		cl = append(cl, client)
	}
	c.mu.RUnlock()
	return cl
}
func (c *ClientHub) get(userId string) (*Client, bool) {
	c.mu.RLock()
	client, ok := c.Users[userId]
	c.mu.RUnlock()
	return client, ok
}
