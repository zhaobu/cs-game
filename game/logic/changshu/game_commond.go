package main

import (
	pbgame "cy/game/pb/game"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type gameCommond struct {
}

func (self *gameCommond) HandleCommond(uid uint64, req *pbgame.GameCommandReq) {
	switch req.CmdType {
	case 1: //强制解散房间
		self.CmdDestroy(uid, req)
	case 2: //要牌
		self.CmdWantCard(uid, req)
	}
}

//强制解散房间
func (self *gameCommond) CmdDestroy(uid uint64, req *pbgame.GameCommandReq) {
	type cmd struct {
		deskId uint64 `json:"deskId"`
	}
	tmp := &cmd{}
	if err := json.Unmarshal([]byte(req.CmdInfo), tmp); err != nil {
		fmt.Println("json.Unmarshal err = ", err)
		return
	}
	//检查桌子是否存在
	d := getDeskByID(tmp.deskId)
	if d == nil {
		tlog.Error("CmdDestroy err,不存在该桌子号", zap.Uint64("deskId", tmp.deskId))
		return
	}
	realReq := &pbgame.DestroyDeskReq{DeskID: tmp.deskId, Type: pbgame.DestroyDeskType_DestroyTypeDebug}
	realRsp := &pbgame.DestroyDeskRsp{}
	d.doDestroyDesk(uid, realReq, realRsp)
	return
}

//要牌
func (self *gameCommond) CmdWantCard(uid uint64, req *pbgame.GameCommandReq) {
	type cmd struct {
		cards []int32 `json:"cards"`
	}
	tmp := &cmd{}
	if err := json.Unmarshal([]byte(req.CmdInfo), tmp); err != nil {
		fmt.Println("json.Unmarshal err = ", err)
		return
	}
	//检查桌子是否存在
	d := getDeskByUID(uid)
	if d == nil {
		tlog.Error("CmdWantCard err,玩家不在桌子中", zap.Uint64("uid", uid))
		return
	}
	if d.gameStatus != pbgame_logic.GameStatus_GSPlaying {
		tlog.Error("CmdWantCard err,不在游戏中要牌失败", zap.Uint64("uid", uid))
		return
	}
	//检查牌是否合法
	chairId := d.GetChairidByUid(uid)
	if chairId == -1 {
		tlog.Error("CmdWantCard err,GetChairidByUid 失败", zap.Uint64("uid", uid))
		return
	}
	d.gameSink.doWantCards(chairId, tmp.cards)
	return
}
