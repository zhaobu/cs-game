package cache

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

// hash
// key friends:* (userid)
// field otherid
// value relaion_time
func AddFriend(userID uint64, otherID uint64) {
	key := fmt.Sprintf("friends:%d", userID)
	redisCli.Do("hset", key, strconv.FormatUint(otherID, 10), time.Now().UTC().Unix())
}

func DelFriend(userID uint64, otherID uint64) {
	key := fmt.Sprintf("friends:%d", userID)
	redisCli.Do("hdel", key, strconv.FormatUint(otherID, 10))
}

func UserFriend(userID uint64) (uids []uint64, err error) {
	key := fmt.Sprintf("friends:%d", userID)
	reply, err := redis.Strings(redisCli.Do("hkeys", key))
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

func AddFriendPending(uid uint64, fid uint64) (exist bool, err error) {
	var key string
	if uid > fid {
		key = fmt.Sprintf("friend_pending:%d_%d", uid, fid)
	} else {
		key = fmt.Sprintf("friend_pending:%d_%d", fid, uid)
	}

	resp, err := redis.Int64(redisCli.Do("setnx", key, 1))
	if err != nil {
		return false, err
	}
	exist = (resp == 0)
	return
}

func DeleteFriendPending(uid uint64, fid uint64) (err error) {
	var key string
	if uid > fid {
		key = fmt.Sprintf("friend_pending:%d_%d", uid, fid)
	} else {
		key = fmt.Sprintf("friend_pending:%d_%d", fid, uid)
	}

	redisCli.Do("del", key)
	return nil
}
