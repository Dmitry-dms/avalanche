package internal

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrUserExists  = "client already exists"
	ErrLimitUsers  = "reached the limit of max active users"
	ErrUserDExists = "client doesn't exists"
)

// ClientHub is a struct that stores information about all
// connected clients and manipulates their state.
type ClientHub struct {
	mu    sync.RWMutex
	Users map[string]*Client
	// amount of active users.
	activeUsers uint64
	// the maximum number of connected users.
	maxUsers uint
	// says how long the ClientHub will exist.
	ttl time.Duration
	// starts a timer that will close all connections when the time expires.
	timer *time.Timer
	// msg is a channel for delivering messages to users.
	msg chan Message
	// expired is a flag that helps not create a ClientHub after a restart.
	expired bool
	// stores creation time in order to correctly calculate TTL after restart.
	Time time.Time
}

// Message is a representation of the message to the Client.
type Message struct {
	*Client
	Msg []byte
}

func newClientHub(maxUsers uint, ttl time.Duration) *ClientHub {
	hub := ClientHub{
		Users:       make(map[string]*Client, maxUsers>>1), // makes a map with capacity maxUsers/2
		activeUsers: 0,
		maxUsers:    maxUsers,
		ttl:         ttl,
		msg:         make(chan Message),
		timer:       time.NewTimer(ttl),
		Time:        time.Now(),
		expired:     false,
	}
	go hub.listen()
	return &hub
}

// IsExpired checks to see if the ClientHub TTL has expired.
func (c *ClientHub) IsExpired() bool {
	return c.expired
}

// listen runs the for loop. It listens for incoming messages and timer expiration.
func (c *ClientHub) listen() {
	for {
		select {
		case <-c.timer.C:
			c.deleteAllHubClients()
			c.expired = true
			return
		case usr := <-c.msg:
			usr.Connection.Write(usr.Msg)
		}
	}
}

// deleteAllHubClients deletes all active Clients.
func (c *ClientHub) deleteAllHubClients() {
	for _, cl := range c.Users {
		cl.Disconnect()
		c.deleteClient(cl.UserId)
		//fmt.Printf("client with id = %s was deleted.  \n", cl.UserId)
	}
}

// addClient adds a client to the ClientHub.
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

// verifyClient checks if the Client exists.
func (c *ClientHub) verifyClient(userId string) bool {
	_, ok := c.get(userId)
	return ok
}

// deleteClient deletes a client from the ClientHub.
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

// GetNumActiveUsers returns the number of acvtive Clients.
func (c *ClientHub) GetNumActiveUsers() uint64 {
	return atomic.LoadUint64(&c.activeUsers)
}

// GetUsers gets all Clients.
func (c *ClientHub) GetUsers() []*Client {
	c.mu.RLock()
	length := len(c.Users)
	cl := make([]*Client, length)
	for _, client := range c.Users {
		cl = append(cl, client)
	}
	c.mu.RUnlock()
	return cl
}

// GetActiveUsersId gets active Client's id.
func (c *ClientHub) GetActiveUsersId() []ClientStat {
	c.mu.RLock()
	length := len(c.Users)
	cl := make([]ClientStat, length)
	counter := 0
	for _, client := range c.Users {
		cl[counter] = ClientStat{UserId: client.UserId}
		counter++
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
