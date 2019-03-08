package tpl

import (
	"context"
	"cy/game/codec"
	"cy/game/pb/game"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (t *RoundTpl) MakeDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			t.Log.WithFields(logrus.Fields{"uid": args.UserID}).Warnf("r:%v stack:%s", r, string(debug.Stack()))
		}
	}()

	pb, err := codec.Msg2Pb(args)
	if err != nil {
		t.Log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbgame.MakeDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.MakeDeskReq")
		t.Log.Error(err.Error())
		return
	}

	t.Log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("recv %s %+v ", args.Name, *req)

	for _, v := range t.plugins {
		if plugin, ok := v.(BeforeMakeDeskReqPlugin); ok {
			plugin.BeforeMakeDeskReq(req)
		}
	}

	for _, v := range t.plugins {
		if plugin, ok := v.(AfterMakeDeskReqPlugin); ok {
			plugin.AfterMakeDeskReq(req)
		}
	}

	return nil
}
