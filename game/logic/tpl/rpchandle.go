package tpl

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/db/mgo"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	"fmt"
	"runtime/debug"
	"time"

	"go.uber.org/zap"
)

type RpcHandle struct {
	service *RoomServie
}

//DestroyDeskReq解散请求
func (self *RpcHandle) DestroyDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			self.service.tlog.Error("recover info", zap.Uint64("uid", args.UserID), zap.Any("stack", string(debug.Stack())))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		self.service.tlog.Error("error info", zap.Error(err))
		return err
	}

	req, ok := pb.(*pbgame.DestroyDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.DestroyDeskReq")
		self.service.tlog.Error("error info", zap.Error(err))
		return
	}

	rsp := &pbgame.DestroyDeskRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	self.service.tlog.Info("recv from gate", zap.Uint64("uid", args.UserID), zap.String("msgName", args.Name), zap.Any("msgValue", *req))

	self.service.roomHandle.HandleDestroyDeskReq(args.UserID, req, rsp)

	if rsp.Code == pbgame.DestroyDeskRspCode_DestroyDeskSucc {
		cache.DeleteClubDeskRelation(req.DeskID)
		cache.DelDeskInfo(req.DeskID)
		cache.FreeDeskID(req.DeskID)
	}

	return
}

//ExitDeskReq退出桌子
func (self *RpcHandle) ExitDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			self.service.tlog.Error("recover info", zap.Uint64("uid", args.UserID), zap.Any("stack", string(debug.Stack())))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		self.service.tlog.Error("error info", zap.Error(err))
		return err
	}

	req, ok := pb.(*pbgame.ExitDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.ExitDeskReq")
		self.service.tlog.Error("error info", zap.Error(err))
		return
	}

	rsp := &pbgame.ExitDeskRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		self.service.ToGateNormal(rsp, args.UserID)
	}()

	self.service.tlog.Info("recv from gate", zap.Uint64("uid", args.UserID), zap.String("msgName", args.Name), zap.Any("msgValue", *req))

	sessInfo, err := cache.QuerySessionInfo(args.UserID)
	if err != nil {
		rsp.Code = pbgame.ExitDeskRspCode_ExitDeskInternalServerError
		return
	}

	if sessInfo.Status != pbcommon.UserStatus_InGameing ||
		sessInfo.GameName != self.service.GameName ||
		sessInfo.GameID != self.service.GameID {
		rsp.Code = pbgame.ExitDeskRspCode_ExitDeskInternalServerError
		return
	}

	self.service.roomHandle.HandleExitDeskReq(args.UserID, req, rsp)

	if rsp.Code == pbgame.ExitDeskRspCode_ExitDeskSucc {
		cache.ExitGame(args.UserID, self.service.GameName, self.service.GameID, sessInfo.AtDeskID)
	}

	return
}

//GameAction 游戏动作
func (self *RpcHandle) GameAction(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			self.service.tlog.Error("recover info", zap.Uint64("uid", args.UserID), zap.Any("stack", string(debug.Stack())))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		self.service.tlog.Error("error info", zap.Error(err))
		return err
	}

	req, ok := pb.(*pbgame.GameAction)
	if !ok {
		err = fmt.Errorf("not *pbgame.GameAction")
		self.service.tlog.Error("error info", zap.Error(err))
		return
	}

	self.service.roomHandle.HandleGameAction(args.UserID, req)

	return
}

//JoinDeskReq 加入桌子
func (self *RpcHandle) JoinDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			self.service.tlog.Error("recover info", zap.Uint64("uid", args.UserID), zap.Any("stack", string(debug.Stack())))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		self.service.tlog.Error("error info", zap.Error(err))
		return err
	}

	req, ok := pb.(*pbgame.JoinDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.JoinDeskReq")
		self.service.tlog.Error("error info", zap.Error(err))
		return
	}

	rsp := &pbgame.JoinDeskRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		if rsp.Code != pbgame.JoinDeskRspCode_JoinDeskSucc {
			self.service.ToGateNormal(rsp, args.UserID)
		}
	}()

	self.service.tlog.Info("recv from gate", zap.Uint64("uid", args.UserID), zap.String("msgName", args.Name), zap.Any("msgValue", *req))

	rsp.Info, err = cache.QueryDeskInfo(req.DeskID)
	if err != nil {
		rsp.Code = pbgame.JoinDeskRspCode_JoinDeskNotExist
		rsp.ErrMsg = err.Error()
		return nil
	}

	succ, err := cache.EnterGame(args.UserID, self.service.GameName, self.service.GameID, req.DeskID, false)
	if err != nil {
		self.service.tlog.Error("error info", zap.Error(err))
		rsp.Code = pbgame.JoinDeskRspCode_JoinDeskInternalServerError
		rsp.ErrMsg = err.Error()
		return nil
	}

	if !succ {
		rsp.Code = pbgame.JoinDeskRspCode_JoinDeskUserStatusErr
		return
	}

	defer func() {
		if rsp.Code != pbgame.JoinDeskRspCode_JoinDeskSucc {
			cache.ExitGame(args.UserID, self.service.GameName, self.service.GameID, req.DeskID)
		}
	}()

	self.service.roomHandle.HandleJoinDeskReq(args.UserID, req, rsp)

	return
}

//SitDownReq 坐下准备
func (self *RpcHandle) SitDownReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			self.service.tlog.Error("recover info", zap.Uint64("uid", args.UserID), zap.Any("stack", string(debug.Stack())))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		self.service.tlog.Error("error info", zap.Error(err))
		return err
	}

	req, ok := pb.(*pbgame.SitDownReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.SitDownReq")
		self.service.tlog.Error("error info", zap.Error(err))
		return
	}

	rsp := &pbgame.SitDownRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		//为保证消息顺序,准备成功消息,游戏内部发送
		if rsp.Code != pbgame.SitDownRspCode_SitDownSucc {
			self.service.ToGateNormal(rsp, args.UserID)
		}

	}()

	self.service.tlog.Info("recv from gate", zap.Uint64("uid", args.UserID), zap.String("msgName", args.Name), zap.Any("msgValue", *req))

	self.service.roomHandle.HandleSitDownReq(args.UserID, req, rsp)

	return
}

//MakeDeskReq 创建桌子
func (self *RpcHandle) MakeDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			self.service.tlog.Error("recover info", zap.Uint64("uid", args.UserID), zap.Any("stack", string(debug.Stack())))
			fmt.Println(string(debug.Stack()))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		self.service.tlog.Error("err codec.Msg2Pb(args)", zap.Error(err))
		return
	}

	req, ok := pb.(*pbgame.MakeDeskReq)
	if !ok {
		self.service.tlog.Error("err pb.(*pbgame.MakeDeskReq)")
		return
	}

	rsp := &pbgame.MakeDeskRsp{Info: &pbcommon.DeskInfo{}}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	var newDeskID uint64

	defer func() {
		self.service.ToGateNormal(rsp, args.UserID)

		if rsp.Code == pbgame.MakeDeskRspCode_MakeDeskSucc && req.ClubID != 0 {
			self.service.SendDeskChangeNotif(req.ClubID, newDeskID, 1)
		}
	}()

	self.service.tlog.Info("recv from gate", zap.Uint64("uid", args.UserID), zap.String("msgName", args.Name), zap.Any("msgValue", *req))
	// 1> 分配桌子ID
	newDeskID, err = cache.AllocDeskID()
	if err != nil {
		self.service.tlog.Error("err cache.AllocDeskID()", zap.Error(err))
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskNotEnoughDesk
		return nil
	}

	defer func() {
		if rsp.Code != pbgame.MakeDeskRspCode_MakeDeskSucc {
			cache.FreeDeskID(newDeskID)
			if req.ClubID != 0 {
				cache.DeleteClubDeskRelation(newDeskID)
			}
		}
	}()

	// 2> 俱乐部和桌子的关系
	var clubInfo *mgo.Club
	if req.ClubID != 0 {
		clubInfo, err = mgo.QueryClubByID(req.ClubID)
		if err != nil {
			self.service.tlog.Error("err mgo.QueryClubByID()", zap.Error(err))
			rsp.Code = pbgame.MakeDeskRspCode_MakeDeskCanNotFindClubID
			return nil
		}

		_ = clubInfo

		cache.AddClubDeskRelation(req.ClubID, newDeskID)

		// // 用默认建房参数
		// if !clubInfo.IsAutoCreate && !clubInfo.IsCustomGameArg {
		// 	for idx, a := range clubInfo.GameArgs {
		// 		if a.Enable && a.GameName == req.GameName {
		// 			req.GameArgMsgName = a.GameArgMsgName
		// 			req.GameArgMsgValue = a.GameArgMsgValue
		// 			self.service.Log.Infof("club:%d use default arg %s %d",
		// 				req.ClubID, req.GameArgMsgName, idx)
		// 			break
		// 		}
		// 	}
		// }
	}

	if self.service.roomHandle.HandleMakeDeskReq(args.UserID, newDeskID, req, rsp) {
		deskInfo := rsp.Info
		deskInfo.ID = newDeskID
		deskInfo.CreateUserID = args.UserID
		if ui, err := mgo.QueryUserInfo(args.UserID); err == nil {
			deskInfo.CreateUserName = ui.Name
			deskInfo.CreateUserProfile = ui.Profile
		}
		deskInfo.CreateTime = time.Now().UTC().Unix()
		// deskInfo.CreateFee =
		// deskInfo.ArgName = rsp.Info.ArgName
		// deskInfo.ArgValue = rsp.Info.ArgValue
		deskInfo.Status = "1"
		deskInfo.GameName = self.service.GameName
		deskInfo.GameID = self.service.GameID
		deskInfo.ClubID = req.ClubID
		deskInfo.Kind = pbgame.DeskType_DTFriend
		// deskInfo.SdInfos
		deskInfo.TotalLoop = 0
		deskInfo.CurrLoop = 0

		err = cache.AddDeskInfo(deskInfo) // 2> 保存桌子信息
		if err != nil {
			self.service.tlog.Error("err cache.AddDeskInfo", zap.Error(err))
			rsp.Code = pbgame.MakeDeskRspCode_MakeDeskInternalServerError
			return nil
		}
		rsp.Code = pbgame.MakeDeskRspCode_MakeDeskSucc
	}
	return nil
}

//QueryDeskInfoReq查询桌子信息
func (self *RpcHandle) QueryDeskInfoReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			self.service.tlog.Error("recover info", zap.Uint64("uid", args.UserID), zap.Any("stack", string(debug.Stack())))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		self.service.tlog.Error("error info", zap.Error(err))
		return err
	}

	req, ok := pb.(*pbgame.QueryDeskInfoReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.QueryDeskInfoReq")
		self.service.tlog.Error("error info", zap.Error(err))
		return
	}

	rsp := &pbgame.QueryDeskInfoRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		self.service.ToGateNormal(rsp, args.UserID)
	}()

	self.service.tlog.Info("recv from gate", zap.Uint64("uid", args.UserID), zap.String("msgName", args.Name), zap.Any("msgValue", *req))

	rsp.Info, err = cache.QueryDeskInfo(req.DeskID)

	self.service.roomHandle.HandleQueryDeskInfoReq(args.UserID, req, rsp)

	return
}

//QueryGameConfigReq查询游戏配置
func (self *RpcHandle) QueryGameConfigReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			self.service.tlog.Error("recover info", zap.Uint64("uid", args.UserID), zap.Any("stack", string(debug.Stack())))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		self.service.tlog.Error("error info", zap.Error(err))
		return err
	}

	req, ok := pb.(*pbgame.QueryGameConfigReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.QueryGameConfigReq")
		self.service.tlog.Error("error info", zap.Error(err))
		return
	}

	rsp := &pbgame.QueryGameConfigRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		self.service.ToGateNormal(rsp, args.UserID)
	}()

	self.service.tlog.Info("recv from gate", zap.Uint64("uid", args.UserID), zap.String("msgName", args.Name), zap.Any("msgValue", *req))

	self.service.roomHandle.HandleQueryGameConfigReq(args.UserID, req, rsp)

	return
}
