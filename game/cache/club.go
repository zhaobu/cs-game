package cache

import (
	"cy/game/pb/common"
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

func AddClubDeskRelation(clubID int64, deskID uint64) error {
	c := redisPool.Get()
	defer c.Close()

	key := fmt.Sprintf("clubdesk:%d", clubID)

	_, err := c.Do("SADD", key,
		strconv.FormatUint(deskID, 10),
	)
	return err
}

func DeleteClubDeskRelation(deskID uint64) error {
	deskInfo, err := QueryDeskInfo(deskID)
	if err != nil {
		return err
	}
	if deskInfo.ClubID == 0 {
		return nil
	}

	c := redisPool.Get()
	defer c.Close()

	key := fmt.Sprintf("clubdesk:%d", deskInfo.ClubID)

	_, err = c.Do("SREM", key, strconv.FormatUint(deskID, 10))
	return err
}

func QueryClubDeskID(clubID int64) (deskIDs []uint64, err error) {
	c := redisPool.Get()
	defer c.Close()

	key := fmt.Sprintf("clubdesk:%d", clubID)

	reply, err := redis.ByteSlices(c.Do("SMEMBERS", key))
	if err != nil {
		return
	}
	for _, v := range reply {
		if x, err := strconv.ParseUint(string(v), 10, 64); err == nil {
			deskIDs = append(deskIDs, x)
		}
	}
	return
}

func QueryClubDeskInfo(clubID int64) (infos []*pbcommon.DeskInfo, err error) {
	deskIDs, err := QueryClubDeskID(clubID)
	if err != nil {
		return nil, err
	}

	for _, deskID := range deskIDs {
		if info, err := QueryDeskInfo(deskID); err == nil {
			infos = append(infos, info)
		}
	}
	return
}
