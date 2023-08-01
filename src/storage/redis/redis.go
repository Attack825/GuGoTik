package redis

import (
	"GuGoTik/src/constant/config"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func init() {
	Client = redis.NewClient(&redis.Options{
		Addr:     config.EnvCfg.RedisAddr,
		Password: config.EnvCfg.RedisPassword,
		DB:       config.EnvCfg.RedisDB,
	})

	if err := redisotel.InstrumentTracing(Client); err != nil {
		panic(err)
	}
}
