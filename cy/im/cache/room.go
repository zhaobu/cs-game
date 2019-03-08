package cache

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

// hash
// key room:* (roomid)
// field uid
// value enter_time
func UserEnterRoom(uid, roomID uint64) {
	name := fmt.Sprintf("room:%d", roomID)
	redisCli.Do("hset", name, strconv.FormatUint(uid, 10), time.Now().UTC().Unix())
}

func UserExitRoom(uid, roomID uint64) {
	name := fmt.Sprintf("room:%d", roomID)
	redisCli.Do("hdel", name, strconv.FormatUint(uid, 10))
}

func RoomUsers(roomID uint64) (uids []uint64) {
	name := fmt.Sprintf("room:%d", roomID)
	reply, err := redis.Strings(redisCli.Do("hkeys", name))
	if err != nil {
		return
	}

	for _, v := range reply {
		if uid, err := strconv.ParseUint(v, 10, 64); err == nil {
			uids = append(uids, uid)
		}
	}
	return
}

//TODO 定期扫描，提出长时间在房间的人
