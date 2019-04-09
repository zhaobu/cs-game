package cache

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

func AllocDeskID() (deskID uint64, err error) {
	c := redisPool.Get()
	defer c.Close()

	left, err := redis.Int(c.Do("SCARD", "emptydesk"))
	if err != nil {
		return 0, err
	}
	if left == 0 {
		rand.Seed(time.Now().Unix())
		var num int64 = 0
		for {
			enter_code := rand.Int63n(999999-100000) + 100000
			reply, err := redis.Int(c.Do("SADD", "emptydesk", enter_code))
			if err != nil {
				return 0, err
			}
			if reply == 1 && num == 1000 {
				break
			}
			if num >= 30000 {
				break
			}
			num++
		}
	}
	reply, err := redis.Strings(c.Do("SPOP", "emptydesk", "1"))
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
