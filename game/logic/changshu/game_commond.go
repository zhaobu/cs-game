package main

import (
	pbgame "cy/game/pb/game"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type gameCommond struct {
}

func (self *gameCommond) HandleCommond(uid uint64, req *pbgame.GameCommandReq, rsp *pbgame.GameCommandRsp) {
	switch req.CmdType {
	case 1: //强制解散房间
		self.CmdDestroy(uid, req, rsp)
	case 2: //要牌
		self.CmdWantCard(uid, req, rsp)
	}
}

//强制解散房间
func (self *gameCommond) CmdDestroy(uid uint64, req *pbgame.GameCommandReq, rsp *pbgame.GameCommandRsp) {
	type cmd struct {
		DeskId uint64 `json:"deskId"`
	}
	tmp := &cmd{}
	if err := json.Unmarshal([]byte(req.CmdInfo), tmp); err != nil {
		fmt.Println("json.Unmarshal err = ", err)
		return
	}
	//检查桌子是否存在
	d := getDeskByID(tmp.DeskId)
	if d == nil {
		tlog.Error("CmdDestroy err,不存在该桌子号", zap.Uint64("deskId", tmp.DeskId))
		return
	}
	realReq := &pbgame.DestroyDeskReq{DeskID: tmp.DeskId, Type: pbgame.DestroyDeskType_DestroyTypeDebug}
	realRsp := &pbgame.DestroyDeskRsp{}
	d.doDestroyDesk(uid, realReq, realRsp)
	return
}

//要牌
func (self *gameCommond) CmdWantCard(uid uint64, req *pbgame.GameCommandReq, rsp *pbgame.GameCommandRsp) {
	type cmd struct {
		Cards []int32 `json:"cards"`
	}
	defer func() {
		if rsp.ErrMsg != "" {
			tlog.Error(rsp.ErrMsg)
		}
	}()
	tmp := &cmd{}
	if err := json.Unmarshal([]byte(req.CmdInfo), tmp); err != nil {
		rsp.ErrMsg = fmt.Sprintf("CmdWantCard err,uid=%d, json.Unmarshal err:%v", uid, err)
		return
	}
	//检查桌子是否存在
	d := getDeskByUID(uid)
	if d == nil {
		rsp.ErrMsg = fmt.Sprintf("CmdWantCard err,uid=%d, 玩家不在桌子中", uid)
		return
	}
	// if d.gameStatus <= pbgame_logic.GameStatus_GSWait {
	// 	rsp.ErrMsg = fmt.Sprintf("CmdWantCard err,uid=%d, 当前游戏还没开始,不能要牌", uid)
	// 	return
	// }
	//检查牌是否合法
	chairId := d.GetChairidByUid(uid)
	if chairId == -1 {
		rsp.ErrMsg = fmt.Sprintf("CmdWantCard err,uid=%d,找不到座位号", uid)
		return
	}
	if len(tmp.Cards) > 0 {
		rsp.ErrMsg = d.gameSink.doWantCards(chairId, tmp.Cards)
	}
	return
}
