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
	if arg.PaymentType == 3 && req.ClubMasterUid == 0 {
		return nil, fmt.Errorf("群主支付时,群主uid为0")
	}
	return arg, nil
}

func calcFee(arg *pbgame_logic.CreateArg) (change int64) {
	// 支付方式 1 房主支付 2 AA支付
	if arg.PaymentType == 1 {
		change = int64(int32(arg.RInfo.Fee) * arg.PlayerCount)
	} else if arg.PaymentType == 2 {
		change = int64(arg.RInfo.Fee)
	}
	return
}

//HandleGameCommandReq游戏指令
func (self *roomHandle) HandleGameCommandReq(uid uint64, req *pbgame.GameCommandReq) {
	rsp := &pbgame.GameCommandRsp{Head: &pbcommon.RspHead{Seq: req.Head.Seq}, CmdType: req.CmdType}
	self.gameCommond.HandleCommond(uid, req, rsp)
	if rsp.ErrMsg != "" {
		rsp.Code = 1
	}
	self.RoomServie.ToGateNormal(rsp, uid)
}

//HandleVoteDestroyDeskReq玩家选择解散请求
func (self *roomHandle) HandleVoteDestroyDeskReq(uid uint64, req *pbgame.VoteDestroyDeskReq) {
	//检查桌子是否存在
	d := getDeskByUID(uid)
	if d == nil {
		tlog.Info("HandleVoteDestroyDeskReq find no desk", zap.Uint64("uid", uid))
		return
	}
	if d.voteInfo == nil {
		tlog.Info("HandleVoteDestroyDeskReq find no voteInfo", zap.Uint64("uid", uid))
		return
	}
	if req.Option == pbgame.VoteOption_VoteOptionNone {
		tlog.Info("HandleVoteDestroyDeskReq bad option", zap.Uint64("uid", uid))
		return
	}
	d.doVoteDestroyDesk(uid, req)
}

//HandleDestroyDeskReq 解散请求
func (self *roomHandle) HandleDestroyDeskReq(uid uint64, req *pbgame.DestroyDeskReq, rsp *pbgame.DestroyDeskRsp) {
	defer func() {
		if rsp.ErrMsg != "" {
			tlog.Info(rsp.ErrMsg)
		}
	}()
	//检查桌子是否存在
	d := getDeskByID(req.DeskID)
	if d == nil {
		rsp.Code = pbgame.DestroyDeskRspCode_DestroyDeskNotExist
		rsp.ErrMsg = fmt.Sprintf("没有该房间号:%d", req.DeskID)
		return
	}
	d.doDestroyDesk(uid, req, rsp)
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
	//如果是俱乐部房间,检查权限
	if d.clubId != 0 {
		rspCode := mgo.CheckSitDownClubRoom(uid, d.clubId, nil)
		if rspCode != 0 {
			switch rspCode {
			case 1:
				rsp.Code = pbgame.SitDownRspCode_SitDownClubNotExist
				rsp.ErrMsg = fmt.Sprintf("roomHandle Desk %d,俱乐部不存在 ", d.deskId)
			case 2:
				rsp.Code = pbgame.SitDownRspCode_SitDownNotInClub
				rsp.ErrMsg = fmt.Sprintf("roomHandle Desk %d,玩家不是俱乐部成员 ", d.deskId)
			case 3:
				rsp.Code = pbgame.SitDownRspCode_SitDownCanNotSitClubDesk
				rsp.ErrMsg = fmt.Sprintf("roomHandle Desk %d,玩家无权加入房间 ", d.deskId)
			case 4:
				rsp.Code = pbgame.SitDownRspCode_SitDownClubHaveReleation
				rsp.ErrMsg = fmt.Sprintf("roomHandle Desk %d,加入俱乐部时有亲属关系 ", d.deskId)
			}
			return
		}
	}
	//检查玩家是否存在桌子信息
	if _, ok := d.deskPlayers[uid]; !ok {
		rsp.Code = pbgame.SitDownRspCode_SitDownNotInDesk
		rsp.ErrMsg = fmt.Sprintf("user%d in desk,but has no deskinfo", uid)
		return
	}
	d.doSitDown(uid, req.ChairId, rsp)
	return
}

func (self *roomHandle) HandleMakeDeskReq(uid, clubMasterUid uint64, deskID uint64, req *pbgame.MakeDeskReq, rsp *pbgame.MakeDeskRsp) bool {
	//检查玩家是否已经在别的房间中
	d := getDeskByUID(uid)
	if d != nil {
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskAlreadyInDesk
		return false
	}
	arg, err := checkArg(req)
	if err != nil {
		tlog.Error("err checkArg()", zap.Error(err))
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskArgsErr
		return false
	}

	tlog.Info("HandleMakeDeskReq", zap.Any("CreateArg", arg))

	//构建桌子参数
	deskArg := &pbgame_logic.DeskArg{Args: arg, Enable: true, Type: pbcommon.DeskType_DTFriend, FeeType: pbgame.FeeType_FTMasonry, DeskID: deskID}
	newD := makeDesk(deskArg, uid, clubMasterUid, deskID, req.ClubID, self.RoomServie)
	payUid := uid
	if arg.PaymentType == 3 { //群主支付
		payUid = req.ClubMasterUid
	}
	if err = newD.updateWealth(payUid, 0); err != nil {
		tlog.Error("err 预扣钻失败", zap.Error(err))
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskNotEnoughMoney
		return false
	}

	//把桌子加入管理
	updateID2desk(newD)

	//返回桌子参数
	rsp.Info.ArgName, rsp.Info.ArgValue, _ = protobuf.Marshal(newD.deskConfig)
	rsp.Info.TotalLoop = int64(newD.deskConfig.Args.RInfo.LoopCnt)
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
