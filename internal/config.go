package internal

type Config struct {
	Address        string
	Name           string
	Version        string
	MaxConnections int
	RedisAddress   string


	// Redis pub/sub prefixes to communicate with outer world.
	RedisMsgPrefix      string
	RedisInfoPrefix     string
	RedisCommandsPrefix string
	MaxUserMessages     int
	SendStatisticInterval int

	AuthJWTkey string
}
