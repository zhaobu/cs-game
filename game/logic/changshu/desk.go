package main

import (
	"cy/game/logic/changshu/game_logic"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	cs "cy/game/pb/game/mj/changshu"
	"sync"
)

type deskUserInfo struct {
	info *pbcommon.UserInfo
}

type desk struct {
	sync.Mutex
	masterId    uint64                //房主
	id          uint64                //桌子id
	gameIndex   int                   //第几局
	gameSink    game_logic.GameSink   //游戏逻辑
	deskPlayers map[int]*deskUserInfo //本桌玩家信息
}

func makeDesk(arg *cs.CreateArg, masterId, deskID uint64) *desk {
	d := &desk{id: deskID}
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
