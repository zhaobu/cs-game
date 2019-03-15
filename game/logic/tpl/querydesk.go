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

func (t *RoundTpl) QueryDeskInfoReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
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

	req, ok := pb.(*pbgame.QueryDeskInfoReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.QueryDeskInfoReq")
		t.Log.Error(err.Error())
		return
	}

	rsp := &pbgame.QueryDeskInfoRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		t.ToGateNormal(rsp, args.UserID)
	}()

	t.Log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("tpl recv %s %+v ", args.Name, *req)

	rsp.Info, err = cache.QueryDeskInfo(req.DeskID)

	t.plugin.HandleQueryDeskInfoReq(args.UserID, req, rsp)

	return
}
