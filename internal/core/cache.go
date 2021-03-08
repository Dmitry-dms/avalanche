package core

import (
	"errors"
	"sync"
	"time"
)

type deleteClientFn func() error

type Cache interface {
	AddCompany(companyName, token string, maxUsers uint, ttl time.Duration) error
	GetCompany(companyName string) (*ClientHub, error)
	DeleteCompany(companyName string)
	AddClient(companyName string, client *Client) (error, deleteClientFn)
	GetClient(companyName, clientId string) (*Client, bool)
	GetActiveUsers(companyName string) (uint, error)
	GetStatisctics() []CompanyStats
}

var (
	companyE  = "company already exists!"
	companyDe = "company doesn't exists"
)

type RamCache struct {
	mu    sync.RWMutex
	users map[string]*ClientHub
}

func NewRamCache() *RamCache {
	return &RamCache{
		users: make(map[string]*ClientHub,20),
	}
}

type CompanyStats struct {
	OnlineUsers uint
	MaxUsers    uint
}

func (r *RamCache) GetStatisctics() []CompanyStats {
	r.mu.RLock()
	var stats []CompanyStats
	for _, company := range r.users {
		stats = append(stats, CompanyStats{
			MaxUsers:    company.maxUsers,
			OnlineUsers: uint(company.GetNumActiveUsers()),
		})
	}
	r.mu.RUnlock()
	if len(stats) == 0 {
		return nil
	}
	return stats
}
func (r *RamCache) AddCompany(companyName, token string, maxUsers uint, ttl time.Duration) error {
	_, ok := r.getCompany(companyName)
	if ok {
		return errors.New(companyE)
	}
	r.mu.Lock()
	r.users[companyName] = newClientHub(maxUsers, token, ttl)
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
		return nil, errors.New(companyDe)
	}
	return c, nil
}
func (r *RamCache) DeleteCompany(companyName string) {
	r.mu.Lock()
	delete(r.users, companyName)
	r.mu.Unlock()
}
func (r *RamCache) AddClient(companyName string, client *Client) (error, deleteClientFn) {
	r.mu.Lock()
	err := r.users[companyName].AddClient(client)
	r.mu.Unlock()
	return err, func() error { return r.deleteClient(companyName, client.UserId) }
}
func (r *RamCache) deleteClient(companyName, clientId string) error {
	r.mu.Lock()
	err := r.users[companyName].DeleteClient(clientId)
	r.mu.Unlock()
	return err
}
func (r *RamCache) GetClient(companyName, clientId string) (*Client, bool) {
	r.mu.RLock()
	client, ok := r.users[companyName].get(clientId)
	r.mu.RUnlock()
	return client, ok
}
func (r *RamCache) GetActiveUsers(companyName string) (uint, error) {
	c, err := r.GetCompany(companyName)
	if err != nil {
		return 0, err
	}
	return uint(c.GetNumActiveUsers()), nil
}

// func (r *RamCache) DeleteOfflineClients() {
// 	r.mu.RLock()
// 	for companyName, clientHub := range r.users {
// 		allUSers := clientHub.GetUsers()
// 		for _, cl := range allUSers {
// 			if cl.Connection.IsClosed() {
// 				r.mu.Lock()
// 				if err := r.users[companyName].DeleteClient(cl.UserId); err != nil {
// 					r.mu.Unlock()
// 					continue
// 				}
// 				r.mu.Unlock()
// 			}
// 		}
// 	}
// }
