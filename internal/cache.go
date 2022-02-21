package internal

import (
	"errors"
	"fmt"
	"sync"
	"time"

)

type deleteClientFn func() error

type Cache interface {
	AddCompany(companyName string, maxUsers uint, ttl time.Duration) error
	GetCompany(companyName string) (*ClientHub, error)
	DeleteCompany(companyName string) error
	AddClient(companyName string, client *Client) (error, deleteClientFn)
	GetClient(companyName, clientId string) (*Client, bool)
	GetActiveUsers(companyName string) (uint, error)
	GetStatisctics() []CompanyStats
	SendMessage(msg Message, companyName string)
}

const (
	ErrCompanyAE  = "company already exists"
	ErrCompanyDE = "company doesn't exists"

	MaxCharLength = 36 // RFC4122 p.3

)
var (
	ErrCharMaxLength = fmt.Sprintf("name must be less than %d characters", MaxCharLength)
)

type RamCache struct {
	mu    sync.RWMutex
	users map[string]*ClientHub
}

func NewRamCache() *RamCache {
	return &RamCache{
		users: make(map[string]*ClientHub, 20),
	}
}



func (r *RamCache) SendMessage(msg Message, companyName string) {
	hub, _ := r.GetCompany(companyName)
	hub.msg <- msg
}

func (r *RamCache) GetStatisctics() []CompanyStats {
	r.mu.RLock()
	var stats []CompanyStats
	for companyName, company := range r.users {
		usersStat := company.GetActiveUsersId()

		stats = append(stats, CompanyStats{
			Name:        companyName,
			Users:       usersStat,
			MaxUsers:    company.maxUsers,
			OnlineUsers: uint(company.GetNumActiveUsers()),
			TTL:         company.ttl,
			Time:        company.Time,
			Stopped:     time.Now(),
			Expired: company.IsExpired(),
		})
	}
	r.mu.RUnlock()
	if len(stats) == 0 {
		return nil
	}
	return stats
}
func (r *RamCache) AddCompany(companyName string, maxUsers uint, ttl time.Duration) error {
	_, ok := r.getCompany(companyName)
	if ok {
		return errors.New(ErrCompanyAE)
	}
	if len(companyName) > MaxCharLength {
		return errors.New(ErrCharMaxLength)
	}
	r.mu.Lock()
	r.users[companyName] = newClientHub(maxUsers, ttl)
	r.mu.Unlock()
	return nil
}
func (r *RamCache) getCompany(name string) (*ClientHub, bool) {
	r.mu.RLock()
	c, ok := r.users[name]
	r.mu.RUnlock()
	return c, ok
}
func (r *RamCache) GetCompany(name string) (*ClientHub, error) {
	c, ok := r.getCompany(name)
	if !ok {
		return nil, errors.New(ErrCompanyDE)
	}
	return c, nil
}
func (r *RamCache) DeleteCompany(companyName string) error {
	_, ok := r.getCompany(companyName)
	if !ok {
		return errors.New(ErrCompanyDE)
	}
	r.mu.Lock()
	delete(r.users, companyName)
	r.mu.Unlock()
	return nil
}
func (r *RamCache) AddClient(companyName string, client *Client) (error, deleteClientFn) {
	company, ok := r.getCompany(companyName)
	if !ok || company.IsExpired() {
		return errors.New(ErrCompanyDE), nil
	}
	if len(client.UserId) > MaxCharLength {
		return errors.New(ErrCharMaxLength), nil
	}
	err := company.addClient(client)
	closeAndDel := func() error {
		err := r.deleteClient(companyName, client.UserId)
		client = nil
		return err
	}
	return err, closeAndDel
}
func (r *RamCache) deleteClient(companyName, clientId string) error {
	company, ok := r.getCompany(companyName)
	if !ok {
		return errors.New(ErrCompanyDE)
	}
	return company.deleteClient(clientId)
}
func (r *RamCache) GetClient(companyName, clientId string) (*Client, bool) {
	company, ok := r.getCompany(companyName)
	if !ok {
		return nil, ok
	}
	client, ok := company.get(clientId)
	return client, ok
}
func (r *RamCache) GetActiveUsers(companyName string) (uint, error) {
	c, ok := r.getCompany(companyName)
	if !ok {
		return 0, errors.New(ErrCompanyDE)
	}
	return uint(c.GetNumActiveUsers()), nil
}

