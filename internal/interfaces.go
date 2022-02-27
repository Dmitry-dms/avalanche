package internal

import (
	"context"
	"io"
	"time"
)

// Upgrader is used to upgrade http connection to websocket.
type Upgrader interface {
	// Upgrade accepts a connection and upgrades it if it has User and Token headers. It also needs to 
	// parse the JWT token.
	Upgrade(conn io.ReadWriter, readBufferSize, writeBufferSize int, userId, companyName *string, parseFunc func(accessToken string) (string, error)) error
}

// deleteClientFn is responsible for removing and disabling the Client.
type deleteClientFn func() error

type Cache interface {
	// AddCompany creates a new CompanyHub.
	AddCompany(companyName string, maxUsers uint, ttl time.Duration) error
	// GetCompany gets a pointer to the CompanyHub if it exists.
	GetCompany(companyName string) (*ClientHub, error)
	// DeleteCompany deletes CompanyHub.
	DeleteCompany(companyName string) error
	// AddClient adds a Client to CompanyHub.
	AddClient(companyName string, client *Client) (error, deleteClientFn)
	// GetClient gets a pointer to the client if it exists. 
	GetClient(companyName, clientId string) (*Client, bool)
	// GetActiveUsers gets number of active users from CompanyHub.
	GetActiveUsers(companyName string) (uint, error)
	// GetStatisctics gets infromation about all CompanyHubs.
	GetStatisctics() CompanyStatsWrapper
	// SendMessage sends message to the Client.
	SendMessage(msg Message, companyName string)
}


type Publisher interface {
	// Publish messages to a message broker.
	Publish(ctx context.Context, channel string, payload []byte) error
}
type Subscriber interface {
	// Subscribe to a message broker channel and pass in a function that will respond to messages.
	Subscribe(ctx context.Context, channel string, f func(msg string))
}
// Storer is responsible for caching key-values.
type Storer interface {
	GetValue(ctx context.Context, k string) (string, error)
	SetValue(ctx context.Context, k string, v []byte, ttl time.Duration) error
	SetArrayValue(ctx context.Context, k string, vals ...interface{}) error
	GetArray(ctx context.Context, k string) ([]string, int)
	DeleteKey(ctx context.Context, k string) error
}
// MessageController is an abstraction for a message broker and a local cache.
type MessageController interface {
	Publisher
	Subscriber
	Storer
	Ping() error
}

