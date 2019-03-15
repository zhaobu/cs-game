package cache

import (
	"cy/game/pb/common"
	"fmt"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

var addDeskInfoScript = redis.NewScript(1, `
	if(redis.call('EXISTS', KEYS[1]) == 1)
	then
		return 2
	else
		redis.call('HMSET', KEYS[1], 'Uuid', ARGV[1], 'ID', ARGV[2], 'CreateUserID', ARGV[3], 'CreateUserName', ARGV[4], 'CreateTime', ARGV[5], 'ArgName',  ARGV[6], 'ArgValue',  ARGV[7], 'Status', ARGV[8], 'GameName', ARGV[9], 'GameID', ARGV[10], 'UserIDs', ARGV[11], 'ClubID', ARGV[12], 'Kind', ARGV[13] )
		return 1
	end
	`)

func AddDeskInfo(info *pbcommon.DeskInfo) (err error) {
	var uids []string
	for _, v := range info.UserIDs {
		uids = append(uids, strconv.FormatUint(v, 10))
	}

	c := redisPool.Get()
	defer c.Close()

	r, err := redis.Int64(addDeskInfoScript.Do(c,
		fmt.Sprintf("deskinfo:%d", info.ID),
		strconv.FormatUint(info.Uuid, 10),
		strconv.FormatUint(info.ID, 10),
		strconv.FormatUint(info.CreateUserID, 10),
		info.CreateUserName,
		info.CreateTime,
		info.ArgName,
		info.ArgValue,
		info.Status,
		info.GameName,
		info.GameID,
		strings.Join(uids, ","),
		strconv.FormatInt(info.ClubID, 10),
		strconv.FormatInt(info.Kind, 10),
	))
	if err != nil {
		return err
	}
	if r != 1 {
		return fmt.Errorf("exists desk %d", info.ID)
	}
	return nil
}

func DelDeskInfo(deskID uint64) {
	c := redisPool.Get()
	defer c.Close()

	c.Do("DEL", fmt.Sprintf("deskinfo:%d", deskID))
}

var updateDeskInfoScript = redis.NewScript(1, `
	if (redis.call('HGET', KEYS[1], 'Uuid')==ARGV[1])
	then
		redis.call('HINCRBY', KEYS[1], 'Uuid', 1)
		redis.call('HMSET', KEYS[1], 'ID', ARGV[2], 'CreateUserID', ARGV[3], 'CreateUserName', ARGV[4], 'CreateTime', ARGV[5], 'ArgName',  ARGV[6], 'ArgValue',  ARGV[7], 'Status', ARGV[8], 'GameName', ARGV[9], 'GameID', ARGV[10], 'UserIDs', ARGV[11], 'ClubID', ARGV[12] )
		return 1
	else
		return 2
	end	
	`)

func UpdateDeskInfo(info *pbcommon.DeskInfo) error {
	var uids []string
	for _, v := range info.UserIDs {
		uids = append(uids, strconv.FormatUint(v, 10))
	}

	c := redisPool.Get()
	defer c.Close()

	_, err := updateDeskInfoScript.Do(c,
		fmt.Sprintf("deskinfo:%d", info.ID),
		strconv.FormatUint(info.Uuid, 10),
		strconv.FormatUint(info.ID, 10),
		strconv.FormatUint(info.CreateUserID, 10),
		info.CreateUserName,
		info.CreateTime,
		info.ArgName,
		info.ArgValue,
		info.Status,
		info.GameName,
		info.GameID,
		strings.Join(uids, ","),
		strconv.FormatInt(info.ClubID, 10),
	)
	return err
}

func QueryDeskInfo(deskID uint64) (*pbcommon.DeskInfo, error) {
	c := redisPool.Get()
	defer c.Close()

	key := fmt.Sprintf("deskinfo:%d", deskID)
	reply, err := redis.StringMap(c.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}

	if reply["ID"] == "" {
		return nil, fmt.Errorf("can not find desk %d", deskID)
	}

	info := &pbcommon.DeskInfo{}
	info.Uuid, _ = strconv.ParseUint(reply["Uuid"], 10, 64)
	info.ID, _ = strconv.ParseUint(reply["ID"], 10, 64)
	info.CreateUserID, _ = strconv.ParseUint(reply["CreateUserID"], 10, 64)
	info.CreateUserName = reply["CreateUserName"]
	info.CreateTime, _ = strconv.ParseInt(reply["CreateTime"], 10, 64)
	info.ArgName = reply["ArgName"]
	info.ArgValue = []byte(reply["ArgValue"])
	info.Status = reply["Status"]
	info.GameName = reply["GameName"]
	info.GameID = reply["GameID"]
	info.ClubID, _ = strconv.ParseInt(reply["ClubID"], 10, 64)
	info.Kind, _ = strconv.ParseInt(reply["Kind"], 10, 64)

	for _, v := range strings.Split(reply["UserIDs"], ",") {
		if uid, err := strconv.ParseUint(v, 10, 64); err == nil {
			info.UserIDs = append(info.UserIDs, uid)
		}
	}

	return info, nil
}
