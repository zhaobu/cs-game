package cache

import (
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

// hash
// key lastread:* (uid)
// field otherid
// value cnt
func LastReadID(uid uint64, otherid uint64) (lastid int64, err error) {
	key := fmt.Sprintf("lastread:%d", uid)
	field := strconv.FormatUint(otherid, 10)
	resp, err := redis.String(redisCli.Do("hget", key, field))
	if err != nil {
		return 0, err
	}

	lastid, err = strconv.ParseInt(resp, 10, 64)
	return
}

func SetLastReadID(uid uint64, otherid uint64, lastid int64) (err error) {
	key := fmt.Sprintf("lastread:%d", uid)
	field := strconv.FormatUint(otherid, 10)
	value := strconv.FormatInt(lastid, 10)
	_, err = redisCli.Do("hset", key, field, value)
	return err
}
