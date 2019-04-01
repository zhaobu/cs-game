package main

import (
	"cy/game/pb/common"
	"cy/game/pb/game"
	cs "cy/game/pb/game/mj/changshu"
	"sync"
)

type playerGameInfo struct {
	info *pbcommon.UserInfo
}

type desk struct {
	sync.Mutex
	createUserID uint64
	id           uint64
	currLoop     int
	clubID       int64

	sitDown map[int]*playerGameInfo
}

func makeDesk(arg *cs.CreateArg, createUserID, deskID uint64, clubID int64) *desk {
	d := &desk{}
	d.id = deskID
	d.clubID = clubID
	return d
}

func (d *desk) doJoin(uid uint64) pbgame.JoinDeskRspCode {
	// 判断条件 是否能加入
	// 重复加入处理
	return pbgame.JoinDeskRspCode_JoinDeskSucc
}

func (d *desk) doExit(uid uint64) uint32 {
	// 判断条件 是否能离开
	return 1 // 默认离开
}

func (d *desk) doAction() {

}
