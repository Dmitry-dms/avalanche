package easyjson

import "time"

type CompanyStats struct {
	Name        string        `json:"company_name"`
	OnlineUsers uint          `json:"online_users"`
	MaxUsers    uint          `json:"max_users"`
	//Users       []ClientStat  `json:"active_users"`
	TTL         time.Duration `json:"ttl"`
	Time        time.Time     `json:"time"`
	Stopped     time.Time     `json:"stoped_time"`
	Expired     bool          `json:"expired"`
}