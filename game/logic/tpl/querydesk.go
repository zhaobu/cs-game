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

func (t *RoundTpl) QueryDeskInfoReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
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

	t.Log.Infof("tpl recv:uid=%d,args.Name=%s,reg=%+v", args.UserID, args.Name, *req)

	rsp.Info, err = cache.QueryDeskInfo(req.DeskID)

	t.plugin.HandleQueryDeskInfoReq(args.UserID, req, rsp)

	return
}
