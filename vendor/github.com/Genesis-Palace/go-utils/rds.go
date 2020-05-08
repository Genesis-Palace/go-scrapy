package go_utils

import "github.com/go-redis/redis"

func NewRedis(config *redis.Options) *redis.Client {
	return redis.NewClient(config)
}

func NewRedisConf(host, password string, db int) *redis.Options {
	return &redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
		MaxRetries: 3,
	}
}
