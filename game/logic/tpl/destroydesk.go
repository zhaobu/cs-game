package tpl

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/pb/game"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (t *RoundTpl) DestroyDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		t.Log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbgame.DestroyDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.DestroyDeskReq")
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

	for _, v := range t.plugins {
		if plugin, ok := v.(DestroyDeskReqPlugin); ok {
			err = plugin.HandleDestroyDeskReq(args.UserID, req)
			if err != nil {
				break
			}
		}
	}

	if err == nil {
		cache.DeleteClubDeskRelation(req.DeskID)
		cache.DelDeskInfo(req.DeskID)
		cache.FreeDeskID(req.DeskID)
	}

	return
}
