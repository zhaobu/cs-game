package cache

import (
	"fmt"
	pbcommon "game/pb/common"
	"strconv"
)

func AddClubDeskRelation(clubID int64, deskID uint64) error {
	key := fmt.Sprintf("clubdesk:%d", clubID)

	redisCli.SAdd(key, key, strconv.FormatUint(deskID, 10))
	return nil
}

func DeleteClubDeskRelation(deskID uint64) error {
	deskInfo, err := QueryDeskInfo(deskID)
	if err != nil {
		return err
	}
	if deskInfo.ClubID == 0 {
		return nil
	}

	key := fmt.Sprintf("clubdesk:%d", deskInfo.ClubID)
	redisCli.SRem(key, strconv.FormatUint(deskID, 10))
	return nil
}

func QueryClubDeskID(clubID int64) (deskIDs []uint64, err error) {
	key := fmt.Sprintf("clubdesk:%d", clubID)
	cmdStr, err := redisCli.SMembers(key).Result()
	if err != nil {
		return
	}
	for _, v := range cmdStr {
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
