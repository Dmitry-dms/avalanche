package internal

import "time"

/*
	Models were moved to a separate file for convenience of code generation.
*/


// brokerMessage is a struct which contains message to Client.
type brokerMessage struct {
	CompanyName string `json:"company_name"`
	ClientId    string `json:"client_id"`
	Message     string `json:"message"`
} //"{\"company_name\":\"testing\",\"client_id\":\"4\",\"message\":\"10\"}"

// The AddCompanyMessage is a structure containing information about the new Company.
type AddCompanyMessage struct {
	CompanyName string `json:"company_name"`
	MaxUsers    uint   `json:"max_users"`
	Duration    int    `json:"duration_hour"`
} //"{\"company_name\":\"testing\",\"max_users\":1000,\"duration_hour\":10}"

// CompanyToken is the token of the newly created Company.
type CompanyToken struct {
	Token      string `json:"token"`
	ServerName string `json:"server_name"`
	Duration   int    `json:"duration_hour"`
}
// AddCompanyResponse is the response that is sent to the message broker after the Company has been created.
type AddCompanyResponse struct {
	Token       CompanyToken `json:"company_token"`
	CompanyName string       `json:"company_name"`
}
// CompanyStats is a structure containing information about existing Сompanies.
type CompanyStats struct {
	Name        string        `json:"company_name"`
	OnlineUsers uint          `json:"online_users"`
	MaxUsers    uint          `json:"max_users"`
	Users       []ClientStat  `json:"active_users"`
	TTL         time.Duration `json:"ttl"`
	Time        time.Time     `json:"time"`
	Stopped     time.Time     `json:"stoped_time"`
	Expired     bool          `json:"expired"`
}
// CompanyStatsWrapper is a wrapper for []CompanyStats.
type CompanyStatsWrapper struct {
	Stats []CompanyStats `json:"stats"`
}
type ClientStat struct {
	UserId string `json:"user_id"`
}