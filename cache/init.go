package cache

import (
	"github.com/go-redis/redis"
)

var (
	redisCli *redis.Client
)

func Init(address string, db int) error {
	redisCli = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "", // no password set
		DB:       db, // use default DB
	})
	return nil
}

func FlushDb(db int) error {
	redisCli.Do("FLUSHDB")
	return nil
}
