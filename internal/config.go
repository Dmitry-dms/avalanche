package internal

type Config struct {
	Port           string
	MonitoringPort string
	Name           string
	Version        string
	// Tells how much connection server can hold.
	MaxConnections int
	// Address to Redis (ip:port).
	RedisAddress string
	// Redis pub/sub prefixes to communicate with outer world.
	RedisMsgPrefix      string
	RedisInfoPrefix     string
	RedisCommandsPrefix string
	// How much messages to user can be stored in Redis.
	MaxUserMessages int
	// The interval for sending information about CompanyHubs to the Redis.
	SendStatisticInterval int
	// Secret JWT key
	AuthJWTkey string
}
