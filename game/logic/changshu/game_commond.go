package main

import (
	pbgame "cy/game/pb/game"
	"encoding/json"
	"fmt"
)

type gameCommond struct {
}

func (self *gameCommond) HandleCommond(uid uint64, req *pbgame.GameCommandReq, rsp *pbgame.GameCommandRsp) {
	switch req.CmdType {
	case pbgame.CmdType_CmdEmoji, pbgame.CmdType_CmdProps, pbgame.CmdType_CmdPhrase, pbgame.CmdType_CmdVoicemail: //发送表情,道具,短语,语音留言
		self.SendFixedInfo(uid, req, rsp)
	case 2: //要牌
		self.CmdWantCard(uid, req, rsp)
	}
}

//发送表情,道具,短语
func (self *gameCommond) SendFixedInfo(uid uint64, req *pbgame.GameCommandReq, rsp *pbgame.GameCommandRsp) {
	//检查玩家是否在桌子
	d := getDeskByUID(uid)
	defer func() {
		if rsp.ErrMsg != "" {
			tlog.Error(rsp.ErrMsg)
		}
	}()
	if d == nil || d.GetChairidByUid(uid) == -1 {
		rsp.ErrMsg = fmt.Sprintf("SendEmojiProps err,玩家%d不在桌子中,无法发送表情道具,短语信息", uid)
		return
	}
	//广播消息
	d.SendData(0, &pbgame.GameCommandNotif{UserID: uid, CmdType: req.CmdType, CmdInfo: req.CmdInfo})
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
