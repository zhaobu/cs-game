package main

import (
	"game/codec/protobuf"
	"game/db/mgo"
	pbgame "game/pb/game"
	cs "game/pb/game/mj/changshu"
	"fmt"
)

func checkArg(req *pbgame.MakeDeskReq) (*cs.CreateArg, error) {
	pb, err := protobuf.Unmarshal(req.GameArgMsgName, req.GameArgMsgValue)
	if err != nil {
		return nil, err
	}

	arg, ok := pb.(*cs.CreateArg)
	if !ok {
		return nil, fmt.Errorf("not *cs.Arg")
	}

	return arg, nil
}

func calcFee(arg *cs.CreateArg) int64 {
	change := int64(0)
	return change
}

func (cs *mjcs) HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp) {
	d := getDeskByID(req.DeskID)
	if d == nil {
		return
	}
	rsp.Code = 1
	if d.clubID != 0 {
		cs.SendDeskChangeNotif(d.clubID, d.id, 3)
	}
	return
}

func (cs *mjcs) HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq, rsp *pbgame.ExitDeskRsp) {
	d := getDeskByUID(uid)
	if d == nil {
		rsp.Code = 2
		return
	}
	rsp.Code = d.doExit(uid)

	return
}

func (cs *mjcs) HandleGameAction(uid uint64, req *pbgame.GameAction) {
	return
}

func (cs *mjcs) HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq, rsp *pbgame.JoinDeskRsp) {
	d := getDeskByID(req.DeskID)
	if d == nil {
		rsp.Code = pbgame.JoinDeskRspCode_JoinDeskNotExist
		rsp.ErrMsg = fmt.Sprintf("mjcs desk %d can not find", req.DeskID)
		return
	}
	rsp.Code = d.doJoin(uid)

	return
}

func (cs *mjcs) HandleMakeDeskReq(uid uint64, deskID uint64, req *pbgame.MakeDeskReq, rsp *pbgame.MakeDeskRsp) {
	arg, err := checkArg(req)
	if err != nil {
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskArgsErr
		return
	}

	cs.Log.Infof("arg %+v\n", arg)

	fee := calcFee(arg)

	if fee != 0 {
		_, err = mgo.UpdateWealthPre(uid, pbgame.FeeType_FTMasonry, fee)
		if err != nil {
			rsp.Code = pbgame.MakeDeskRspCode_MakeDeskNotEnoughMoney
			return
		}

		defer func() {
			if err != nil {
				mgo.UpdateWealthPreSure(uid, pbgame.FeeType_FTMasonry, fee)
			}
		}()
	}

	newD := makeDesk(arg, uid, deskID, req.ClubID)
	updateID2desk(newD)

	rsp.Code = pbgame.MakeDeskRspCode_MakeDeskSucc
	return
}

func (cs *mjcs) HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq, rsp *pbgame.QueryGameConfigRsp) {
	if req.Type == 2 {
		rsp.CfgName, rsp.CfgValue, _ = protobuf.Marshal(&argTpl)
	}

	return
}

func (cs *mjcs) HandleQueryDeskInfoReq(uid uint64, req *pbgame.QueryDeskInfoReq, rsp *pbgame.QueryDeskInfoRsp) {
	rsp.Code = 1
	return
}

func (cs *mjcs) RunLongTime(deskID uint64, typ int) bool {
	d := getDeskByID(deskID)
	if d == nil {
		return false
	}

	deleteID2desk(deskID)
	// deleteUser2desk()

	return true
}