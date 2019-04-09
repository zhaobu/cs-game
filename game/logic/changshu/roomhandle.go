package main

import (
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	pbgame "cy/game/pb/game"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"fmt"

	"go.uber.org/zap"
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

func (self *roomHandle) HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp) {
	return
}

//HandleExitDeskReq退出桌子
func (self *roomHandle) HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq, rsp *pbgame.ExitDeskRsp) {
	d := getDeskByUID(uid)
	if d == nil {
		rsp.Code = 2
		return
	}
	rsp.Code = d.doExit(uid)
	return
}

func (self *roomHandle) HandleGameAction(uid uint64, req *pbgame.GameAction) {
	d := getDeskByUID(uid)
	if d != nil {
		d.doAction(uid, req.ActionName, req.ActionValue)
	} else {
		log.Infof("can not find desk by uid %d", uid)
	}
	return
}

//加入桌子
func (self *roomHandle) HandleJoinDeskReq(uid uint64, req *pbgame.JoinDeskReq, rsp *pbgame.JoinDeskRsp) {
	//检查桌子是否存在
	d := getDeskByID(req.DeskID)
	if d == nil {
		rsp.Code = pbgame.JoinDeskRspCode_JoinDeskNotExist
		rsp.ErrMsg = fmt.Sprintf("roomHandle Desk %d can not find", req.DeskID)
		return
	}
	//检查玩家是否已经在桌子中
	if old := getDeskByUID(uid); old != nil {
		rsp.Code = pbgame.JoinDeskRspCode_JoinDeskAlreadyInDesk
		rsp.ErrMsg = fmt.Sprintf("user already in desk:%d", old.id)
		return
	}
	updateUser2desk(d, uid)
	rsp.Code = d.doJoin(uid)
	return
}

//坐下准备
func (self *roomHandle) HandleSitDownReq(uid uint64, req *pbgame.SitDownReq, rsp *pbgame.SitDownRsp) {
	//检查玩家是否存在桌子
	d := getDeskByUID(uid)
	if d == nil {
		rsp.Code = pbgame.SitDownRspCode_SitDownNotInDesk
		rsp.ErrMsg = fmt.Sprintf("user%d not in desk", uid)
		return
	}
	//检查玩家是否存在桌子信息
	if _, ok := d.deskPlayers[uid]; !ok {
		rsp.Code = pbgame.SitDownRspCode_SitDownNotInDesk
		rsp.ErrMsg = fmt.Sprintf("user%d in desk,but has no deskinfo", uid)
		return
	}
	//TODO距离限制
	d.doSitDown(uid, rsp)
	return
}

//起立取消准备
func (self *roomHandle) HandleStandUpReq(uid uint64, req *pbgame.SitDownReq, rsp *pbgame.SitDownRsp) {
	//检查玩家是否存在桌子
	d := getDeskByUID(uid)
	if d == nil {
		rsp.Code = pbgame.SitDownRspCode_SitDownNotInDesk
		rsp.ErrMsg = fmt.Sprintf("user%d not in desk", uid)
		return
	}
	//检查玩家是否存在桌子信息
	if _, ok := d.deskPlayers[uid]; !ok {
		rsp.Code = pbgame.SitDownRspCode_SitDownNotInDesk
		rsp.ErrMsg = fmt.Sprintf("user%d in desk,but has no deskinfo", uid)
		return
	}
	//TODO距离限制
	d.doSitDown(uid, rsp)
	return
}

func (self *roomHandle) HandleMakeDeskReq(uid uint64, deskID uint64, req *pbgame.MakeDeskReq, rsp *pbgame.MakeDeskRsp) {
	arg, err := checkArg(req)
	if err != nil {
		tlog.Error("err checkArg()", zap.Error(err))
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskArgsErr
		return
	}

	tlog.Info("HandleMakeDeskReq", zap.Any("CreateArg", arg))

	fee := calcFee(arg)

	if fee != 0 {
		_, err = mgo.UpdateWealthPre(uid, feeTypeMasonry, fee)
		if err != nil {
			tlog.Error("err mgo.UpdateWealthPre()", zap.Error(err))
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
	newD.gameNode = self.RoomServie
	//把桌子加入管理
	updateID2desk(newD)

	rsp.Code = pbgame.MakeDeskRspCode_MakeDeskSucc
	return
}

func (self *roomHandle) HandleQueryGameConfigReq(uid uint64, req *pbgame.QueryGameConfigReq, rsp *pbgame.QueryGameConfigRsp) {
	if req.Type == 2 {
		rsp.CfgName, rsp.CfgValue, _ = protobuf.Marshal(&argTpl)
	}

	return
}

func (self *roomHandle) HandleQueryDeskInfoReq(uid uint64, req *pbgame.QueryDeskInfoReq, rsp *pbgame.QueryDeskInfoRsp) {
	rsp.Code = 1
	return
}

func (self *roomHandle) RunLongTime(deskID uint64, typ int) bool {
	d := getDeskByID(deskID)
	if d == nil {
		return false
	}

	deleteID2desk(deskID)
	// deleteUser2desk()

	return true
}
