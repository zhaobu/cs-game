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

func (t *RoundTpl) ExitDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
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

	req, ok := pb.(*pbgame.ExitDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.ExitDeskReq")
		t.Log.Error(err.Error())
		return
	}

	rsp := &pbgame.ExitDeskRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {

		t.ToGateNormal(rsp, args.UserID)
	}()

	t.Log.Infof("tpl recv:uid=%d,args.Name=%s,reg=%+v", args.UserID, args.Name, *req)

	sessInfo, err := cache.QuerySessionInfo(args.UserID)
	if err != nil {
		rsp.Code = 2
		return
	}

	if sessInfo.Status != pbcommon.UserStatus_InGameing ||
		sessInfo.GameName != t.gameName ||
		sessInfo.GameID != t.gameID {
		return
	}

	t.plugin.HandleExitDeskReq(args.UserID, req, rsp)

	if rsp.Code == 1 {
		cache.ExitGame(args.UserID, t.gameName, t.gameID, sessInfo.AtDeskID)
	}

	return
}
