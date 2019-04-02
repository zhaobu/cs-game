package main

import (
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	pbgame "cy/game/pb/game"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"fmt"
)

const (
	feeTypeGold    = 1
	feeTypeMasonry = 2

	deskTypeMatch  = 1
	deskTypeFriend = 2
	deskTypeLadder = 3
)

func checkArg(req *pbgame.MakeDeskReq) (*pbgame_logic.CreateArg, error) {
	pb, err := protobuf.Unmarshal(req.GameArgMsgName, req.GameArgMsgValue)
	if err != nil {
		return nil, err
	}

	arg, ok := pb.(*pbgame_logic.CreateArg)
	if !ok {
		return nil, fmt.Errorf("not *pbgame_logic.Arg")
	}

	return arg, nil
}

func calcFee(arg *pbgame_logic.CreateArg) int64 {
	change := int64(0)
	return change
}

func (self *mjcs) HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp) {
	return
}

func (self *mjcs) HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq, rsp *pbgame.ExitDeskRsp) {
	d := getDeskByUID(uid)
	if d == nil {
		rsp.Code = 2
		return
	}
	rsp.Code = d.doExit(uid)
	return
}

func (self *mjcs) HandleGameAction(uid uint64, req *pbgame.GameAction) {
	d := getDeskByUID(uid)
	if d != nil {
		d.doAction(uid, req.ActionName, req.ActionValue)
	} else {
		log.Infof("can not find desk by uid %d", uid)
	}
	return
}

func (self *mjcs) HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq, rsp *pbgame.JoinDeskRsp) {
	d := getDeskByID(req.DeskID)
	if d == nil {
		rsp.Code = pbgame.JoinDeskRspCode_JoinDeskNotExist
		rsp.ErrMsg = fmt.Sprintf("mjcs Desk %d can not find", req.DeskID)
		return
	}
	updateID2desk(d)
	updateUser2desk(d, uid)
	rsp.Code = d.doJoin(uid)
	return
}

func (self *mjcs) HandleMakeDeskReq(uid uint64, deskID uint64, req *pbgame.MakeDeskReq, rsp *pbgame.MakeDeskRsp) {
	arg, err := checkArg(req)
	if err != nil {
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskArgsErr
		return
	}

	self.Log.Infof("arg %+v\n", arg)

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
	newD.gameNode = self
	newD.doJoin(uid)

	updateID2desk(newD)
	updateUser2desk(newD, uid)

	rsp.Code = pbgame.MakeDeskRspCode_MakeDeskSucc
	return
}

func (self *mjcs) HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq, rsp *pbgame.QueryGameConfigRsp) {
	if req.Type == 2 {
		rsp.CfgName, rsp.CfgValue, _ = protobuf.Marshal(&argTpl)
	}

	return
}

func (self *mjcs) HandleQueryDeskInfoReq(uid uint64, req *pbgame.QueryDeskInfoReq, rsp *pbgame.QueryDeskInfoRsp) {
	rsp.Code = 1
	return
}

func (self *mjcs) RunLongTime(deskID uint64, typ int) bool {
	d := getDeskByID(deskID)
	if d == nil {
		return false
	}

	deleteID2desk(deskID)
	// deleteUser2desk()

	return true
}
