package cache

import (
	"github.com/gomodule/redigo/redis"
)

func ScanDeskInfo(cursor string, pattern string) (keys []string, next string, err error) {
	c := redisPool.Get()
	defer c.Close()

	result, err := redis.MultiBulk(c.Do("SCAN", cursor, "MATCH", pattern, "COUNT", 20))
	if err != nil {
		return nil, "0", err
	}

	next = string(result[0].([]byte))
	for _, v := range result[1].([]interface{}) {
		keys = append(keys, string(v.([]byte)))
	}
	return
}
