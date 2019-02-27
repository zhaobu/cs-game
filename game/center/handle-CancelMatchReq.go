package main

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/pb/center"
	"cy/game/pb/common"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (p *center) CancelMatchReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbcenter.CancelMatchReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.CancelMatchReq")
		logrus.Error(err.Error())
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

		logrus.WithFields(logrus.Fields{
			"req":   *req,
			"rsp":   *rsp,
			"err":   err,
			"r":     r,
			"stack": stack,
		}).Info(args.Name)
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
