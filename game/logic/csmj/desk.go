package main

import (
	"cy/game/pb/game"
	"cy/game/pb/game/ddz"
)

type desk struct {
}

func newDesk(arg *pbgame_ddz.RoomArg, createUserID uint64) *desk {
	d := &desk{}
	return d
}

func (d *desk) joinDesk(uid uint64) pbgame.JoinDeskRspCode {
	return pbgame.JoinDeskRspCode_JoinDeskSucc
}
