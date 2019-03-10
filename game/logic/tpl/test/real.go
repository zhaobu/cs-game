package main

import (
	"cy/game/pb/game"
	"cy/game/pb/game/ddz"
)

type deskMgr struct {
}

func (d *deskMgr) BeforeMakeDeskReq(*pbgame.MakeDeskReq) error {
	log.Infof("BeforeMakeDeskReq deskMgr")

	// 检查参数

	return nil
}

func (d *deskMgr) AfterMakeDeskReq(*pbgame.MakeDeskReq) error {
	log.Infof("AfterMakeDeskReq deskMgr")
	return nil
}

func checkArg(req *pbgame.MakeDeskReq) error {
	return nil
}

func calcFee(arg *pbgame_ddz.RoomArg) int64 {
	change := int64(0)
	if arg.PaymentType == 1 {
		change = int64(arg.Fee)
	} else if arg.PaymentType == 2 {
		change = int64(arg.Fee / arg.SeatCnt)
	}
	return change
}
