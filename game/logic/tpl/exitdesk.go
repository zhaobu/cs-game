package tpl

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/pb/common"
	"cy/game/pb/game"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (t *RoundTpl) ExitDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
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

	t.Log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("recv %s %+v ", args.Name, *req)

	defer func() {
		r := recover()
		if r != nil {
			t.Log.WithFields(logrus.Fields{"uid": args.UserID}).Warnf("r:%v stack:%s", r, string(debug.Stack()))
		}
	}()

	sessInfo, err := cache.QuerySessionInfo(args.UserID)
	if err != nil {
		return
	}

	if sessInfo.Status != pbcommon.UserStatus_InGameing ||
		sessInfo.GameName != t.gameName ||
		sessInfo.GameID != t.gameID {
		return
	}

	err = t.plugin.HandleExitDeskReq(args.UserID, req)

	if err == nil {
		cache.ExitGame(args.UserID, t.gameName, t.gameID, sessInfo.AtDeskID)
	}

	return
}
