package main

import (
	"cy/game/db/mgo"

	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	cs "cy/game/pb/game/mj/changshu"
	"sync"

	"github.com/gogo/protobuf/proto"
)

type (
	desk_state uint8
)

const (
	LOOKON   desk_state = iota //观察者状态
	SIT_DOWN                   //坐下准备状态
	PLAYING                    //游戏状态
)

type deskUserInfo struct {
	info       *pbcommon.UserInfo
	desk_state desk_state
}

type Desk struct {
	*mjcs
	sync.Mutex
	masterId    uint64                   //房主
	id          uint64                   //桌子id
	gameIndex   uint32                   //第几局
	gameSink    *GameSink                //游戏逻辑
	deskPlayers map[uint64]*deskUserInfo //本桌玩家信息
	// lookonPlayers map[uint64]*deskUserInfo //观察玩家信息
	playChair  map[uint64]uint16 //玩家uid到chairid
	deskConfig *cs.CreateArg     //桌子参数
}

func makeDesk(arg *cs.CreateArg, masterId, deskID uint64) *Desk {
	d := &Desk{id: deskID, masterId: masterId, deskConfig: arg}
	d.gameSink = new(GameSink)
	// d.gameSink.Ctor(arg)
	d.playChair = make(map[uint64]uint16)
	d.deskPlayers = make(map[uint64]*deskUserInfo)
	return d
}

func (d *Desk) doEnter(uid uint64) pbgame.JoinDeskRspCode {
	// 判断条件 是否能加入

	// 重复加入处理
	return pbgame.JoinDeskRspCode_JoinDeskSucc
}

func (d *Desk) getFreeChair() (uint16, bool) {
	var i uint32
	for i = 0; i < d.deskConfig.PlayerCount; i++ {
		var f bool = false
		for _, chair := range d.playChair {
			if chair == uint16(i) {
				f = true
				break
			}
		}
		if f == false {
			return uint16(i), true
		}

	}
	return 0, false
}

func (d *Desk) checkStart() bool {
	if len(d.playChair) < 3 {
		return false
	}
	//检查是否所有玩家都准备好
	for uid, _ := range d.playChair {
		if d.deskPlayers[uid].desk_state != SIT_DOWN {
			return false
		}
	}
	return true
}

func (d *Desk) doJoin(uid uint64) pbgame.JoinDeskRspCode {
	//判断人数满没满
	chair, ok := d.getFreeChair()
	if false == ok {
		return pbgame.JoinDeskRspCode_JoinDeskDeskFull
	}
	var err error
	dUserInfo := &deskUserInfo{desk_state: SIT_DOWN}
	userInfo, err := mgo.QueryUserInfo(uid)
	if err != nil {
		return pbgame.JoinDeskRspCode_JoinDeskUserStatusErr
	}
	dUserInfo.info = userInfo
	d.deskPlayers[uid] = dUserInfo
	d.playChair[uid] = chair
	d.gameSink.AddPlayer(chair, uid, userInfo.GetName())
	//先发送加入成功消息
	d.SendData(&pbgame.JoinDeskRsp{Code: pbgame.JoinDeskRspCode_JoinDeskSucc}, uid)
	//再判断游戏开始
	if d.checkStart() {
		d.gameSink.StartGame()
	}
	return pbgame.JoinDeskRspCode_JoinDeskSucc
}

func (d *Desk) doExit(uid uint64) uint32 {
	// 判断条件 是否能离开
	return 1 // 默认离开
}

func (d *Desk) doAction() {

}

func (d *Desk) SendData(pb proto.Message, uids ...uint64) {
	d.mjcs.RoundTpl.ToGateNormal(pb, uids...)
}

func (d *Desk) SendGameMessage(pb proto.Message, uids ...uint64) {
	d.mjcs.RoundTpl.ToGate(pb, uids...)
}
