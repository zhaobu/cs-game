package cache

import (
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

func AllocDeskID() (deskID uint64, err error) {
	c := redisPool.Get()
	defer c.Close()

	reply, err := redis.Strings(c.Do("SPOP", "emptydesk", "1"))
	if err != nil {
		return 0, err
	}
	if len(reply) == 1 {
		return strconv.ParseUint(reply[0], 10, 64)
	}
	return 0, fmt.Errorf("not enough")
}

func FreeDeskID(deskID uint64) (err error) {
	c := redisPool.Get()
	defer c.Close()

	_, err = c.Do("SADD", "emptydesk", strconv.FormatUint(deskID, 10))
	return
}

func SCAN(pattern string, count int) (find []string) {
	c := redisPool.Get()
	defer c.Close()

	if count < 1 || count > 50 {
		count = 50
	}

	const start = string("0")
	var cursor string = start

	for {
		reply, err := redis.MultiBulk(c.Do("SCAN", cursor, "MATCH", pattern, "COUNT", count))
		if err != nil {
			return
		}

		if len(reply) != 2 {
			break
		}

		findN, _ := redis.ByteSlices(reply[1], nil)
		for _, v := range findN {
			find = append(find, string(v))
		}

		cursor = string(reply[0].([]byte))
		if cursor == start {
			break
		}
	}
	return
}
