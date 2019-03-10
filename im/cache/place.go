package cache

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	redisCli redis.Conn // TODO pool
)

func Init(address string) error {
	var err error
	redisCli, err = redis.Dial("tcp", address)
	if err != nil {
		return err
	}
	return nil
}

// hash
// key place
// field * (uid)
// value [json_data]
func UserOnline(userID uint64, gateID string) error {
	field := fmt.Sprintf("%d", userID)
	var data = struct {
		GateID string `json:"gateid"`
		T      int64  `josn:"t"`
	}{gateID, time.Now().UTC().Unix()}

	value, _ := json.Marshal(data)
	_, err := redis.Int64(redisCli.Do("hset", "place", field, value))
	return err
}

func UserOffline(userID uint64) error {
	field := fmt.Sprintf("%d", userID)
	_, err := redis.Int64(redisCli.Do("hdel", "place", field))
	return err
}

func QueryUser(userID uint64) (gateID string, t int64, err error) {
	field := fmt.Sprintf("%d", userID)
	reply, err2 := redis.Bytes(redisCli.Do("hget", "place", field))
	if err2 != nil {
		err = err2
		return
	}
	var data = struct {
		GateID string `json:"gateid"`
		T      int64  `josn:"t"`
	}{}
	err = json.Unmarshal(reply, &data)
	if err != nil {
		return
	}
	gateID = data.GateID
	t = data.T
	return
}

func QueryUsers(userIDs ...uint64) (places map[uint64]string, err error) {
	var args []interface{}
	args = append(args, "place")
	for _, v := range userIDs {
		args = append(args, strconv.FormatUint(v, 10))
	}

	reply, err2 := redis.ByteSlices(redisCli.Do("hmget", args...))
	if err2 != nil {
		err = err2
		return
	}

	places = make(map[uint64]string, len(userIDs))
	for idx, uid := range userIDs {
		v := reply[idx]
		if string(v) != "" {
			var data = struct {
				GateID string `json:"gateid"`
				T      int64  `josn:"t"`
			}{}
			if err2 := json.Unmarshal(v, &data); err2 == nil {
				places[uid] = data.GateID
			}
		}
	}

	return
}
