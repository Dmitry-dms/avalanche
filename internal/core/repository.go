package core

type deleteClientFn func() error
type AvalacnheCache interface {
	AddCompany(companyName string, maxUsers uint)
	GetCompany(companyName string) *ClientHub
	DeleteCompany(companyName string)
	AddClient(companyName string, client *Client) (error, deleteClientFn)
	GetClient(companyName, clientId string) (*Client, bool)
	DeleteOfflineClients()
	GetActiveUsers() uint
}