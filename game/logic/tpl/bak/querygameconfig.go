package tpl

import (
	"context"
	"cy/game/codec"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	"fmt"
	"runtime/debug"
)

func (t *RoundTpl) QueryGameConfigReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
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

	req, ok := pb.(*pbgame.QueryGameConfigReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.QueryGameConfigReq")
		t.Log.Error(err.Error())
		return
	}

	rsp := &pbgame.QueryGameConfigRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		t.ToGateNormal(rsp, args.UserID)
	}()

	t.Log.Infof("tpl recv:uid=%d,args.Name=%s,reg=%+v", args.UserID, args.Name, *req)

	t.plugin.HandleQueryGameConfigReq(args.UserID, req, rsp)

	return
}
