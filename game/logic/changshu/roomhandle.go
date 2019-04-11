package main

import (
	"cy/game/codec/protobuf"
	"cy/game/db/mgo"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"fmt"

	"go.uber.org/zap"
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
	// 支付方式 1 个人支付 2 平局支付
	if arg.PaymentType == 1 {
		change = int64(arg.RInfo.Fee)
	} else if arg.PaymentType == 2 {
		change = int64(arg.RInfo.Fee / uint32(arg.PlayerCount))
	}
	change = 0
	return change
}

//HandleDestroyDeskReq 解散请求
func (self *roomHandle) HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp) {
	//检查桌子是否存在
	d := getDeskByID(req.DeskID)
	if d == nil {
		rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskNotExist
		rsp.ErrMsg = fmt.Sprintf("没有该房间号:%d", req.DeskID)
		return
	}
	d.doDestroyDesk(uid, rsp)
	return
}

//HandleExitDeskReq退出桌子
func (self *roomHandle) HandleExitDeskReq(uid uint64, req *pbgame.ExitDeskReq, rsp *pbgame.ExitDeskRsp) {
	d := getDeskByUID(uid)
	if d == nil {
		rsp.Code = pbgame.ExitDeskRspCode_ExitDeskNotInDesk
		return
	}
	d.doExit(uid, rsp)
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
		rsp.ErrMsg = fmt.Sprintf("user already in desk:%d", old.deskId)
		return
	}
	updateUser2desk(d, uid)
	d.doJoin(uid, rsp)
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
	d.doSitDown(uid, req.ChairId, rsp)
	return
}

func (self *roomHandle) HandleMakeDeskReq(uid uint64, deskID uint64, req *pbgame.MakeDeskReq, rsp *pbgame.MakeDeskRsp) bool {
	arg, err := checkArg(req)
	if err != nil {
		tlog.Error("err checkArg()", zap.Error(err))
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskArgsErr
		return false
	}

	tlog.Info("HandleMakeDeskReq", zap.Any("CreateArg", arg))

	fee := calcFee(arg)

	if fee != 0 {
		_, err = mgo.UpdateWealthPre(uid, pbgame.FeeType_FTMasonry, fee)
		if err != nil {
			tlog.Error("err mgo.UpdateWealthPre()", zap.Error(err))
			rsp.Code = pbgame.MakeDeskRspCode_MakeDeskNotEnoughMoney
			return false
		}

		//建房失败,返还扣除的房费
		defer func() {
			if err != nil {
				mgo.UpdateWealthPreSure(uid, pbgame.FeeType_FTMasonry, fee)
			}
		}()
	}

	//构建桌子参数
	deskArg := &pbgame_logic.DeskArg{Args: arg, Enable: true, Type: pbcommon.DeskType_DTFriend, FeeType: pbgame.FeeType_FTMasonry, DeskID: deskID}
	newD := makeDesk(deskArg, uid, deskID, req.ClubID)
	newD.gameNode = self.RoomServie
	//把桌子加入管理
	updateID2desk(newD)

	//返回桌子参数
	rsp.Info.ArgName, rsp.Info.ArgValue, _ = protobuf.Marshal(newD.deskConfig)
	return true
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
