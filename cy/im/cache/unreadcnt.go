package cache

import (
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

// hash
// key unread_cnt:* uid
// field otherid
// value cnt >int
func UnreadCnt(uid uint64) (cnt map[uint64]int64, err error) {

	name := fmt.Sprintf("unread_cnt:%d", uid)
	resp, err := redis.StringMap(redisCli.Do("hgetall", name))
	if err != nil {
		return nil, err
	}

	cnt = make(map[uint64]int64)
	for k, v := range resp {
		otherid, _ := strconv.ParseUint(k, 10, 64)
		urcnt, _ := strconv.ParseInt(v, 10, 64)
		cnt[otherid] = urcnt
	}
	return
}

func ChangeUnreadCnt(uid uint64, cnt map[uint64]int64) (err error) {
	key := fmt.Sprintf("unread_cnt:%d", uid)
	for otherid, change := range cnt {
		redisCli.Do("hincrby", key, strconv.FormatUint(otherid, 10), change)
	}
	return nil
}
