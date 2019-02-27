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
