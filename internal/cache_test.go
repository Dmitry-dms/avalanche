package internal

import (
	"testing"
	"time"
)

const(
	CompanyName = "test"
	ClientId = "test"
)

func TestAddCompany(t *testing.T) {
	cache := NewRamCache()
	_ = cache.AddCompany(CompanyName, 10, time.Second*5)
	_, err := cache.GetCompany(CompanyName)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteCompany(t *testing.T) {
	cache := NewRamCache()
	_ = cache.AddCompany(CompanyName, 10, time.Second*5)
	err := cache.DeleteCompany(CompanyName)
	if err != nil {
		t.Error(err)
	}
	_, err = cache.GetCompany(CompanyName)
	if err == nil {
		t.Errorf("the company should not exist")
	}
}
func TestAddClient(t *testing.T) {
	cache := NewRamCache()
	_ = cache.AddCompany(CompanyName, 10, time.Second*5)
	client := &Client{UserId: ClientId}
	err, _ := cache.AddClient(CompanyName, client)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteClient(t *testing.T) {
	cache := NewRamCache()
	_ = cache.AddCompany(CompanyName, 10, time.Second*5)
	client := &Client{UserId: ClientId}
	err, deleteClientFn:= cache.AddClient(CompanyName, client)
	if err != nil {
		t.Error(err)
	}
	deleteClientFn()
	_, ok := cache.GetClient(CompanyName, ClientId)
	if ok {
		t.Error("the client should not exist")
	}
}


