package internal

import "time"

type redisMessage struct {
	CompanyName string `json:"company_name"`
	ClientId    string `json:"client_id"`
	Message     string `json:"message"`
} //"{\"company_name\":\"testing\",\"client_id\":\"4\",\"message\":\"10\"}"
type AddCompanyMessage struct {
	CompanyName string `json:"company_name"`
	MaxUsers    uint   `json:"max_users"`
	Duration    int    `json:"duration_hour"`
} //"{\"company_name\":\"testing\",\"max_users\":1000,\"duration_hour\":10}"
type CompanyToken struct {
	Token      string `json:"token"`
	ServerName string `json:"server_name"`
	Duration   int    `json:"duration_hour"`
}
type AddCompanyResponse struct {
	Token       CompanyToken `json:"company_token"`
	CompanyName string       `json:"company_name"`
}
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
type ClientStat struct {
	UserId string `json:"user_id"`
}