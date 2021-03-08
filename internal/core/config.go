package core

type Config struct {
	Name           string
	Version        string
	MaxConnections int
	RedisAddress   string

	// Redis pub/sub prefixes to communicate with outer world.
	RedisMsgPrefix      string
	RedisInfoPrefix     string
	RedisCommandsPrefix string

	AuthJWTkey string
}
