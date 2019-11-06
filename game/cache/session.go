package cache

import (
	"fmt"
	pbcommon "game/pb/common"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func QuerySessionInfo(userID uint64) (*pbcommon.SessionInfo, error) {
	key := fmt.Sprintf("sessioninfo:%d", userID)
	reply, err := redisCli.HGetAll(key).Result()
	if err != nil {
		return nil, err
	}

	info := &pbcommon.SessionInfo{}
	info.Uuid, _ = strconv.ParseUint(reply["Uuid"], 10, 64)
	info.UserID, _ = strconv.ParseUint(reply["UserID"], 10, 64)
	info.SessionID = reply["SessionID"]
	info.Status = pbcommon.UserStatus(pbcommon.UserStatus_value[reply["Status"]])
	info.AtDeskID, _ = strconv.ParseUint(reply["AtDeskID"], 10, 64)
	info.GameName = reply["GameName"]
	info.GameID = reply["GameID"]
	info.LastActiveTime, _ = strconv.ParseInt(reply["LastActiveTime"], 10, 64)
	rid, _ := strconv.ParseUint(reply["RoomID"], 10, 64)
	info.RoomID = uint32(rid)
	return info, nil
}

var enterMatchScript = redis.NewScript(`
	local key = KEYS[1]
	local game_name = ARGV[1]
	local room_id = ARGV[2]
	local last_active_time = ARGV[3]

	local xx = redis.call('HMGET', key, 'Status', 'GameName', 'RoomID')

	if(xx[1]==false)
	then
		redis.call('HMSET', KEYS[1], 'Status', 'InMatching', 'GameName', game_name, 'RoomID', room_id, 'LastActiveTime', last_active_time, 'Uuid', 1)
		return {1, 'InMatching', game_name, room_id}
	end

	if(xx[1]=='InMatching' and xx[2]==game_name and xx[3]==room_id)
	then
		return {2, 'InMatching', game_name, room_id}
	end

	return {3, xx[1], xx[2], xx[3]}
	`)

func EnterMatch(userID uint64, gameName string, roomID uint32) (code int64, inStatus, inGameName string, inRoomID uint32, err error) {
	r, err := enterMatchScript.Run(redisCli,
		[]string{fmt.Sprintf("sessioninfo:%d", userID)},
		gameName,
		strconv.FormatUint(uint64(roomID), 10),
		time.Now().UTC().Unix(),
	).Result()
	arrary := r.([]interface{})
	code = arrary[0].(int64)
	if arrary[1] != nil {
		inStatus = string(arrary[1].([]byte))
	}
	if arrary[2] != nil {
		inGameName = string(arrary[2].([]byte))
	}
	if arrary[3] != nil {
		rid, _ := strconv.ParseUint(string(arrary[3].([]byte)), 10, 64)
		inRoomID = uint32(rid)
	}

	return
}

// 没有加GameName和RoomID的判断
// redis.call('HDEL', KEYS[1], 'Status', 'GameName', 'RoomID')
var exitMatchScript = redis.NewScript(`
	if(redis.call('HGET', KEYS[1], 'Status')=='InMatching')
	then
		redis.call('DEL', KEYS[1])
		return 1
	else
		return 2
	end
	`)

func ExitMatch(userID uint64) (succ bool, err error) {
	r, err := exitMatchScript.Run(redisCli,
		[]string{fmt.Sprintf("sessioninfo:%d", userID)},
	).Result()
	if err != nil {
		return false, err
	}
	if r == 1 {
		return true, nil
	}
	return false, nil
}

var enterGameScript = redis.NewScript(`
	local key = KEYS[1]
	local from_match = ARGV[1]
	local game_name = ARGV[2]
	local game_id = ARGV[3]
	local desk_id = ARGV[4]
	local last_active_time = ARGV[5]

	if(from_match=='1')
	then
		if(redis.call('HGET', key, 'Status')=='InMatching')
		then
			redis.call('HMSET', key, 'Status', 'InGameing', 'GameName', game_name, 'GameID', game_id, 'AtDeskID', desk_id, 'LastActiveTime', last_active_time)
			return 1
		end
		return -1
	else
		local xx = redis.call('HMGET', key, 'Status', 'GameName', 'GameID', 'AtDeskID')
		if(xx[1]=='InGameing' and xx[2]==game_name and xx[3]==game_id and xx[4]==desk_id)
		then
			return 3
		else
			if(xx[1]==false)
			then
				redis.call('HMSET', key, 'Uuid', 1, 'Status', 'InGameing', 'GameName', game_name, 'GameID', game_id, 'AtDeskID', desk_id, 'LastActiveTime', last_active_time)
				return 2
			end
		end
		return -2
	end
	`)

func EnterGame(userID uint64, gameName, gameID string, deskID uint64, fromMatch bool) (succ bool, err error) {
	r, err := enterGameScript.Run(redisCli,
		[]string{fmt.Sprintf("sessioninfo:%d", userID)},
		fromMatch,
		gameName,
		gameID,
		strconv.FormatUint(deskID, 10),
		time.Now().UTC().Unix()).Result()
	if err != nil {
		return false, err
	}

	if r.(int64) > 0 {
		return true, nil
	}
	return false, nil
}

// redis.call('HDEL', KEYS[1], 'Status', 'GameName', 'GameID', 'AtDeskID', 'LastActiveTime', 'RoomID')
var exitGameScript = redis.NewScript(`
	local xx = redis.call('HMGET', KEYS[1], 'Status', 'GameName', 'GameID', 'AtDeskID')
	if(xx[1]=='InGameing' and xx[2]==ARGV[1] and xx[3]==ARGV[2] and xx[4]==ARGV[3])
	then
		redis.call('DEL', KEYS[1])
		return 1
	else
		return 2
	end
	`)

func ExitGame(userID uint64, gameName, gameID string, deskID uint64) (succ bool, err error) {
	r, err := exitGameScript.Run(redisCli,
		[]string{fmt.Sprintf("sessioninfo:%d", userID)},
		gameName,
		gameID,
		strconv.FormatUint(deskID, 10),
	).Result()
	if err != nil {
		return false, err
	}
	if r == 1 {
		return true, nil
	}
	return false, nil
}
