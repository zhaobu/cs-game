package cache

import (
	"github.com/gomodule/redigo/redis"
)

var (
	redisPool *redis.Pool
)

func Init(address string, db int) error {
	redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address, redis.DialDatabase(db))
		},
	}
	return nil
}

func FlushDb(db int) error {
	c := redisPool.Get()
	defer c.Close()
	c.Do("FLUSHDB")
	return nil
}
