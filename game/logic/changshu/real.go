package main

import (
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	pbgame "cy/game/pb/game"
	cs "cy/game/pb/game/mj/changshu"
	"fmt"
)

const (
	feeTypeGold    = 1
	feeTypeMasonry = 2

	deskTypeMatch  = 1
	deskTypeFriend = 2
	deskTypeLadder = 3
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
		_, err = mgo.UpdateWealthPre(uid, feeTypeMasonry, fee)
		if err != nil {
			rsp.Code = pbgame.MakeDeskRspCode_MakeDeskNotEnoughMoney
			return
		}

		defer func() {
			if err != nil {
				mgo.UpdateWealthPreSure(uid, feeTypeMasonry, fee)
			}
		}()
	}

	newD := makeDesk(arg, uid, deskID)
	newD.doJoin(uid)

	updateID2desk(newD)
	updateUser2desk(newD, uid)

	rsp.Code = pbgame.MakeDeskRspCode_MakeDeskSucc
	return
}

func (cs *mjcs) HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq, rsp *pbgame.QueryGameConfigRsp) {
	if req.Type == 2 {
		rsp.Name, rsp.Value, _ = protobuf.Marshal(&argTpl)
	}

	return
}
