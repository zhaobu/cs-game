package tpl

import (
	"cy/game/cache"
	"cy/game/pb/common"
	"fmt"
	"time"
)

func (t *RoundTpl) delInvalidDesk() {
	var delDesks []uint64
	for _, key := range cache.SCAN("deskinfo:*", 50) {
		var deskID uint64
		fmt.Sscanf(key, "deskinfo:%d", &deskID)
		deskInfo, err := cache.QueryDeskInfo(deskID)
		if err != nil {
			continue
		}

		if deskInfo.GameName != t.gameName || deskInfo.GameID != t.gameID {
			continue
		}

		delDesks = append(delDesks, deskID)
	}

	for _, key := range cache.SCAN("sessioninfo:*", 50) {
		var userID uint64
		fmt.Sscanf(key, "sessioninfo:%d", &userID)
		sessInfo, err := cache.QuerySessionInfo(userID)
		if err != nil {
			continue
		}

		if sessInfo.GameName != t.gameName ||
			sessInfo.GameID != t.gameID ||
			sessInfo.Status != pbcommon.UserStatus_InGameing {
			continue
		}

		for _, deskID := range delDesks {
			if sessInfo.AtDeskID == deskID {
				cache.ExitGame(userID, sessInfo.GameName, sessInfo.GameID, deskID)
				break
			}
		}
	}

	for _, deskID := range delDesks {
		cache.DeleteClubDeskRelation(deskID)
		cache.DelDeskInfo(deskID)
		cache.FreeDeskID(deskID)
	}
}

func (t *RoundTpl) checkDeskLongTime() {
	go func() {
		time.Sleep(time.Minute * 5)
		now := time.Now().UTC()

		for _, key := range cache.SCAN("deskinfo:*", 50) {
			var deskID uint64
			fmt.Sscanf(key, "deskinfo:%d", &deskID)
			deskInfo, err := cache.QueryDeskInfo(deskID)
			if err != nil {
				continue
			}

			if deskInfo.GameName != t.gameName || deskInfo.GameID != t.gameID {
				continue
			}

			doneOk := false
			du := now.Sub(time.Unix(deskInfo.CreateTime, 0)).Minutes()

			if du > (time.Minute * 60).Minutes() {
				doneOk = t.plugin.RunLongTime(deskID, 2)
			} else if du > (time.Minute*30).Minutes() && deskInfo.Status == "" {
				doneOk = t.plugin.RunLongTime(deskID, 1)
			}

			if doneOk {
				cache.DeleteClubDeskRelation(deskID)
				cache.DelDeskInfo(deskID)
				cache.FreeDeskID(deskID)
			}
		}
	}()
}
