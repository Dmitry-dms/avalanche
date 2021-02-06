package core

import (
	"sync"
)

type RamCache struct {
	mu    sync.RWMutex
	users map[string]*ClientHub
}

func NewRamCache() *RamCache {
	return &RamCache{
		users: make(map[string]*ClientHub),
	}
}
func (r *RamCache) AddCompany(companyName string, maxUsers uint) {
	r.mu.Lock()
	r.users[companyName] = newClientHub(maxUsers)
	r.mu.Unlock()
}
func (r *RamCache) GetCompany(companyName string) *ClientHub {
	r.mu.Lock()
	hub := r.users[companyName]
	r.mu.Unlock()
	return hub
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
	return err, func() error {return r.deleteClient(companyName, client.UserId)}
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
func (r *RamCache) GetActiveUsers() uint {
	return uint(r.users["test"].GetNumActiveUsers())
}
func (r *RamCache) DeleteOfflineClients() {
	r.mu.RLock()
	for companyName, clientHub := range r.users {
		allUSers := clientHub.GetUsers()
		for _, cl := range allUSers {
			if cl.Connection.IsClosed() {
				r.mu.Lock()
				if err := r.users[companyName].DeleteClient(cl.UserId); err != nil {
					r.mu.Unlock()
					continue
				}
				r.mu.Unlock()
			}
		}
	}
}


