package tpl

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	pbgame "cy/game/pb/game"
	"fmt"
	"runtime/debug"
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

	rsp := &pbgame.DestroyDeskRsp{}

	t.Log.Infof("uid%d recv %s %+v", args.UserID, args.Name, *req)

	defer func() {
		r := recover()
		if r != nil {
			t.Log.Warnf("uid%d recover:%v stack:%s", args.UserID, r, string(debug.Stack()))
		}
	}()

	t.plugin.HandleDestroyDeskReq(args.UserID, req, rsp)

	if rsp.Code == 1 {
		cache.DeleteClubDeskRelation(req.DeskID)
		cache.DelDeskInfo(req.DeskID)
		cache.FreeDeskID(req.DeskID)
	}

	return
}
