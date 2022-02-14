package internal

import (
	"errors"
	"fmt"

	"sync"
	"sync/atomic"
	"time"

	"github.com/mailru/easygo/netpoll"
)

var (
	ErrUserExists  = "client already exists"
	ErrLimitUsers  = "reached the limit of max active users"
	ErrUserDExists = "client doesn't exists"
)

type ClientHub struct {
	mu          sync.RWMutex
	Users       map[string]*Client
	activeUsers uint64
	maxUsers    uint

	ttl         time.Duration
	timer       *time.Timer
	poller      netpoll.Poller
	msg         chan Message
	expired     bool

	Time time.Time
}

type Message struct {
	*Client
	Msg []byte
}

func newClientHub(maxUsers uint, ttl time.Duration, wr func(msg Message), p netpoll.Poller) *ClientHub {
	hub := ClientHub{
		Users:       make(map[string]*Client, maxUsers >> 1),
		activeUsers: 0,
		maxUsers:    maxUsers,
		ttl:         ttl,
		msg:         make(chan Message),
		timer:       time.NewTimer(ttl),
		poller:      p,
		Time: time.Now(),
	}
	go hub.listen(wr)
	return &hub
}
func (c *ClientHub) IsExpired() bool {
	return c.expired
}
func (c *ClientHub) listen(wr func(msg Message)) {
	for {
		select {
		case <-c.timer.C:
			println("timeout")
			c.deleteClients()
			c.expired = true
			return
		case usr := <-c.msg:
			wr(usr)
		}
	}
}

func (c *ClientHub) deleteClients() {
	for _, cl := range c.Users {
		cl.Desc.Close()
		//err := c.poller.Stop(cl.Desc)
		err2 := cl.Disconnect()
		err3 := c.deleteClient(cl.UserId)
		fmt.Printf("client with id = %s was deleted. {%s} {%s} \n", cl.UserId, err2, err3)
	}
}

func (c *ClientHub) addClient(client *Client) error {
	if ok := c.verifyClient(client.UserId); ok {
		return errors.New(ErrUserExists)
	}

	if c.GetNumActiveUsers() >= uint64(c.maxUsers) {
		return errors.New(ErrLimitUsers)
	}
	c.mu.Lock()
	c.Users[client.UserId] = client
	c.mu.Unlock()
	atomic.AddUint64(&c.activeUsers, 1)
	return nil
}
func (c *ClientHub) verifyClient(userId string) bool {
	//c.mu.RLock()
	_, ok := c.get(userId)
	//c.mu.RUnlock()
	return ok
}
func (c *ClientHub) deleteClient(userId string) error {
	if ok := c.verifyClient(userId); !ok {
		return errors.New(ErrUserDExists)
	}
	c.mu.Lock()
	delete(c.Users, userId)
	c.mu.Unlock()
	atomic.AddUint64(&c.activeUsers, ^uint64(0))
	return nil
}
func (c *ClientHub) GetNumActiveUsers() uint64 {
	return atomic.LoadUint64(&c.activeUsers)
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
func (c *ClientHub) GetActiveUsersId() []ClientStat {
	var cl []ClientStat
	c.mu.RLock()
	for _, client := range c.Users {
		cl = append(cl, ClientStat{UserId: client.UserId})
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
