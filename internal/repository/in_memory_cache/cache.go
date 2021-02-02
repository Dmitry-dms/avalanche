package in_memory_cache

import (
	"sync"

	"github.com/Dmitry-dms/avalanche/internal/core"
)

type RamCache struct {
	mu    sync.RWMutex
	users map[string]*core.ClientHub
}

func NewRamCache() *RamCache {
	return &RamCache{
		users: make(map[string]*core.ClientHub),
	}
}
func (r *RamCache) AddCompany(companyName string, maxUsers uint) {
	r.mu.Lock()
	r.users[companyName] = core.NewClientHub(maxUsers)
	r.mu.Unlock()
}
func (r *RamCache) GetCompany(companyName string) *core.ClientHub {
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
func (r *RamCache) AddClient(companyName string, client *core.Client) error {
	r.mu.Lock()
	err := r.users[companyName].AddClient(client)
	r.mu.Unlock()
	return err
}
func (r *RamCache) DeleteClient(companyName, clientId string) error {
	r.mu.Lock()
	err := r.users[companyName].DeleteClient(clientId)
	r.mu.Unlock()
	return err
}
func (r *RamCache) GetClient(companyName, clientId string) (*core.Client, bool) {
	r.mu.RLock()
	client, ok := r.users[companyName].Get(clientId)
	r.mu.RUnlock()
	return client, ok
}
func (r *RamCache) DeleteOfflineClients() {
	r.mu.RLock()
	for companyName, clientHub := range r.users {
		allUSers := clientHub.GetUsers()
		for _, cl := range allUSers {
			if cl.Connection.IsClosed() {
				r.mu.Lock()
				if err := r.users[companyName].DeleteClient(cl.UserId); err != nil {
					continue
				}
				r.mu.Unlock()
			}
		}
	}
}

type AvalacnheCache interface {
	AddCompany(companyName string, maxUsers uint)
	DeleteCompany(companyName string)
	AddClient(companyName, clientId string, client *core.Client) error
	GetClient(companyName, clientId string) (*core.Client, bool)
	DeleteClient(companyName, clientId string) error
	DeleteOfflineClients()
}
