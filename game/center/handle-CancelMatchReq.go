package main

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	pbcenter "cy/game/pb/center"
	pbcommon "cy/game/pb/common"
	"fmt"
	"runtime/debug"
)

func (p *center) CancelMatchReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbcenter.CancelMatchReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.CancelMatchReq")
		tlog.Error(err.Error())
		return
	}

	rsp := &pbcenter.CancelMatchRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		r := recover()
		stack := ""
		if r != nil {
			stack = string(debug.Stack())
		}

		codec.Pb2Msg(rsp, reply)

		log.Infof("args.Name:%s:req:%v,rsp:%v,err:%v,r:%v,stack:%v", args.Name, *req, *rsp, err, r, stack)
	}()

	// ###############################
	rsp.Succ = cancelMatch(args.UserID)
	return nil
}

func cancelMatch(uid uint64) (succCancel bool) {
	sess, err := cache.QuerySessionInfo(uid)
	if err != nil {
		return
	}
	if sess.Status != pbcommon.UserStatus_InMatching {
		return
	}

	gn := sess.GameName
	rid := sess.RoomID
	roomName := fmt.Sprintf("%s_%d", gn, rid)

	mu.Lock()
	defer mu.Unlock()

	if waiter[roomName] != nil {
		return waiter[roomName].cancel(uid)
	}
	return
}
