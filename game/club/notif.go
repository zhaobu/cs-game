package main

import (
	"cy/game/db/mgo"
	"cy/game/pb/club"
)

func sendClubChangeInfo(clubID int64, typ clubChangeTyp, changeUserID uint64) {
	onlineSubMembers := make([]uint64, 0)

	cc := getClub(clubID)
	if cc == nil {
		return
	}

	cc.RLock()
	clubChangeInfo := &pbclub.ClubChangeInfo{}
	clubChangeInfo.Typ = int32(typ)
	clubChangeInfo.UserID = changeUserID
	clubChangeInfo.Info = &pbclub.BriefInfo{
		ID:           clubID,
		Name:         cc.Name,
		Profile:      cc.Profile,
		MasterUserID: cc.MasterUserID,
	}

	for _, m := range cc.Members {
		cu := mustGetUserOther(m.UserID)
		cu.RLock()
		subed := cu.subFlag
		cu.RUnlock()
		if subed {
			onlineSubMembers = append(onlineSubMembers, m.UserID)
		}
	}

	cc.RUnlock()

	toGateNormal(clubChangeInfo, onlineSubMembers...)
}

//指定发送目标
func sendClubChangeInfoByuIds(clubID int64, typ clubChangeTyp, changeUserID uint64,uIds []uint64) {
	onlineSubMembers := make([]uint64, 0)

	cc := getClub(clubID)
	if cc == nil {
		return
	}

	cc.RLock()
	clubChangeInfo := &pbclub.ClubChangeInfo{}
	clubChangeInfo.Typ = int32(typ)
	clubChangeInfo.UserID = changeUserID
	clubChangeInfo.Info = &pbclub.BriefInfo{
		ID:           clubID,
		Name:         cc.Name,
		Profile:      cc.Profile,
		MasterUserID: cc.MasterUserID,
	}

	for _, m := range uIds {
		cu := mustGetUserOther(m)
		cu.RLock()
		subed := cu.subFlag
		cu.RUnlock()
		if subed {
			onlineSubMembers = append(onlineSubMembers, m)
		}
	}

	cc.RUnlock()

	toGateNormal(clubChangeInfo, onlineSubMembers...)
}

func sendClubEmail(ce *mgo.ClubEmail, to ...uint64) {
	ceci := &pbclub.ClubEmailChangeInfo{}
	ceci.Emails = append(ceci.Emails, &pbclub.ClubEmail{
		ID:       ce.ID,
		SendTime: ce.SendTime,
		Typ:      ce.Typ,
		Content:  ce.Content,
		Flag:     ce.Flag,
		ClubID:   ce.ClubID,
	})

	toGateNormal(ceci, to...)
}
