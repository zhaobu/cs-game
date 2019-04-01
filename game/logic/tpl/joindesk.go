package tpl

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	"fmt"
	"runtime/debug"
)

func (t *RoundTpl) JoinDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			t.Log.Warnf("recover:uid=%d,stack=%s", args.UserID, string(debug.Stack()))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		t.Log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbgame.JoinDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.JoinDeskReq")
		t.Log.Error(err.Error())
		return
	}

	rsp := &pbgame.JoinDeskRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		//为保证消息顺序,加入成功消息,游戏内部发送
		if rsp.Code != pbgame.JoinDeskRspCode_JoinDeskSucc {
			t.ToGateNormal(rsp, args.UserID)
		}
	}()

	t.Log.Infof("tpl recv:uid=%d,args.Name=%s,reg=%+v", args.UserID, args.Name, *req)

	rsp.Info, err = cache.QueryDeskInfo(req.DeskID)
	if err != nil {
		rsp.Code = pbgame.JoinDeskRspCode_JoinDeskNotExist
		rsp.ErrMsg = err.Error()
		return nil
	}

	succ, err := cache.EnterGame(args.UserID, t.gameName, t.gameID, req.DeskID, false)
	if err != nil {
		t.Log.Error(err.Error())
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
			cache.ExitGame(args.UserID, t.gameName, t.gameID, req.DeskID)
		}
	}()

	t.plugin.HandleJoinDeskReq(args.UserID, req, rsp)

	return
}
