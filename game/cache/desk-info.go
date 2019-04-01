package cache

import (
	"cy/game/pb/common"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

var addDeskInfoScript = redis.NewScript(1, `
	if(redis.call('EXISTS', KEYS[1]) == 1)
	then
		return 2
	else
		redis.call('HMSET', KEYS[1], 'Uuid', ARGV[1], 'ID', ARGV[2], 'CreateUserID', ARGV[3], 'CreateUserName', ARGV[4], 'CreateUserProfile', ARGV[5], 'CreateTime', ARGV[6], 'CreateFee', ARGV[7], 'ArgName',  ARGV[8], 'ArgValue',  ARGV[9], 'Status', ARGV[10], 'GameName', ARGV[11], 'GameID', ARGV[12], 'ClubID', ARGV[13], 'Kind', ARGV[14], 'SdInfos', ARGV[15], 'TotalLoop', ARGV[16], 'CurrLoop', ARGV[17])
		return 1
	end
	`)

func AddDeskInfo(info *pbcommon.DeskInfo) (err error) {
	sdInfo, _ := json.Marshal(info.SdInfos)

	c := redisPool.Get()
	defer c.Close()

	r, err := redis.Int64(addDeskInfoScript.Do(c,
		fmt.Sprintf("deskinfo:%d", info.ID),
		strconv.FormatUint(info.Uuid, 10),
		strconv.FormatUint(info.ID, 10),
		strconv.FormatUint(info.CreateUserID, 10),
		info.CreateUserName,
		info.CreateUserProfile,
		strconv.FormatInt(info.CreateTime, 10),
		strconv.FormatInt(info.CreateFee, 10),
		info.ArgName,
		info.ArgValue,
		info.Status,
		info.GameName,
		info.GameID,
		strconv.FormatInt(info.ClubID, 10),
		strconv.FormatInt(info.Kind, 10),
		sdInfo,
		strconv.FormatInt(info.TotalLoop, 10),
		strconv.FormatInt(info.CurrLoop, 10),
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
		redis.call('HMSET', KEYS[1], 'ID', ARGV[2], 'CreateUserID', ARGV[3], 'CreateUserName', ARGV[4], 'CreateUserProfile', ARGV[5], 'CreateTime', ARGV[6], 'CreateFee', ARGV[7], 'ArgName',  ARGV[8], 'ArgValue',  ARGV[9], 'Status', ARGV[10], 'GameName', ARGV[11], 'GameID', ARGV[12], 'ClubID', ARGV[13], 'Kind', ARGV[14], 'SdInfos', ARGV[15], 'TotalLoop', ARGV[16], 'CurrLoop', ARGV[17])
		return 1
	else
		return 2
	end
	`)

func UpdateDeskInfo(info *pbcommon.DeskInfo) error {
	sdInfo, _ := json.Marshal(info.SdInfos)

	c := redisPool.Get()
	defer c.Close()

	_, err := updateDeskInfoScript.Do(c,
		fmt.Sprintf("deskinfo:%d", info.ID),
		strconv.FormatUint(info.Uuid, 10),
		strconv.FormatUint(info.ID, 10),
		strconv.FormatUint(info.CreateUserID, 10),
		info.CreateUserName,
		info.CreateUserProfile,

		strconv.FormatInt(info.CreateTime, 10),
		strconv.FormatInt(info.CreateFee, 10),
		info.ArgName,
		info.ArgValue,
		info.Status,

		info.GameName,
		info.GameID,
		strconv.FormatInt(info.ClubID, 10),
		strconv.FormatInt(info.Kind, 10),
		sdInfo,

		strconv.FormatInt(info.TotalLoop, 10),
		strconv.FormatInt(info.CurrLoop, 10),
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
	info.CreateUserProfile = reply["CreateUserProfile"]
	info.CreateTime, _ = strconv.ParseInt(reply["CreateTime"], 10, 64)
	info.CreateFee, _ = strconv.ParseInt(reply["CreateFee"], 10, 64)
	info.ArgName = reply["ArgName"]
	info.ArgValue = []byte(reply["ArgValue"])
	info.Status = reply["Status"]
	info.GameName = reply["GameName"]
	info.GameID = reply["GameID"]
	info.ClubID, _ = strconv.ParseInt(reply["ClubID"], 10, 64)
	info.Kind, _ = strconv.ParseInt(reply["Kind"], 10, 64)
	info.SdInfos = make([]*pbcommon.SiteDownPlayerInfo, 0)
	json.Unmarshal([]byte(reply["SdInfos"]), &info.SdInfos)
	info.TotalLoop, _ = strconv.ParseInt(reply["TotalLoop"], 10, 64)
	info.CurrLoop, _ = strconv.ParseInt(reply["CurrLoop"], 10, 64)

	return info, nil
}
