package main

import (
	"cy/game/db/mgo"
	pbgame "cy/game/pb/game"
	pbgame_ddz "cy/game/pb/game/ddz"
)

func getDeskByID(deskID uint64) *desk {
	return nil
}

func checkArg(req *pbgame.MakeDeskReq) (*pbgame_ddz.RoomArg, error) {
	return nil, nil
}

func calcFee(arg *pbgame_ddz.RoomArg) int64 {
	change := int64(0)
	return change
}

func (cs *mjcs) HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq) (err error) {
	return
}

func (cs *mjcs) HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq) (err error) {

	return
}

func (cs *mjcs) HandleGameAction(uid uint64, req *pbgame.GameAction) (err error) {
	return
}

func (cs *mjcs) HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq) (err error) {
	
	req.DeskID 
	return
}

func (cs *mjcs) HandleMakeDeskReq(uid uint64, req *pbgame.MakeDeskReq, deskID uint64) (err error) {
	log.Infof("HandleMakeDeskReq %+v\n", req)

	arg, err := checkArg(req)
	if err != nil {
		return
	}

	fee := calcFee(arg)

	_, err = mgo.UpdateWealthPre(uid, arg.FeeType, fee)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			mgo.UpdateWealthPreSure(uid, arg.FeeType, fee)
		}
	}()

	arg.Type = 2
	arg.DeskID = deskID
	arg.RoomId = 0

	newD := newDesk(arg, uid)
	newD.joinDesk(uid)

	return nil
}

func (cs *mjcs) HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq) (err error) {
	return
}
