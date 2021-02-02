package core

import "sync"

type companyHub struct {
	mu        sync.RWMutex
	companies map[string]*ClientHub
}

func newCompanyHub() *companyHub {
	return &companyHub{
		companies: make(map[string]*ClientHub),
	}
}
func (c *companyHub) AddCompany(companyName string, maxUsers uint) {
	c.mu.Lock()
	c.companies[companyName] = NewClientHub(maxUsers)
	c.mu.Unlock()
}
func (c *companyHub) GetCompanyClientHub(companyName string) *ClientHub{
	c.mu.RLock()
	hub := c.companies[companyName]
	c.mu.RUnlock()
	return hub
}