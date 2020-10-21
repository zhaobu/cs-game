package main

import (
	"context"
	"game/codec"
	"game/codec/protobuf"
	"game/logic/ddz/desk"
	pbinner "game/pb/inner"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (p *ddz) GameMatchArgsCheckReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbinner.GameMatchArgsCheckReq)
	if !ok {
		err = fmt.Errorf("not *pbinner.GameMatchArgsCheckReq")
		log.Error(err.Error())
		return
	}

	rsp := &pbinner.GameMatchArgsCheckRsp{}

	defer func() {
		r := recover()
		stack := ""
		if r != nil {
			stack = string(debug.Stack())
		}

		codec.Pb2Msg(rsp, reply)

		log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("req:%s %+v rsp:%+v err:%s r:%v stack:%s",
			args.Name, *req, *rsp, err, r, stack)
	}()

	// 取参数
	roomArg := desk.QueryMatchRoomArg(req.RoomId)
	if roomArg == nil || !roomArg.Enable {
		rsp.Code = 2
		return
	}

	uinfo, err := desk.QueryUserInfo(req.UserID)
	if err != nil {
		rsp.Code = 3
		return
	}

	// 进入限制判断
	var haveWealth uint64
	if roomArg.FeeType == 1 {
		haveWealth = uinfo.Gold
	} else if roomArg.FeeType == 2 {
		haveWealth = uinfo.Masonry
	}
	if haveWealth < roomArg.EnterMin || haveWealth > roomArg.EnterMax {
		rsp.Code = 4
		return
	}

	// 一同返回桌子参数
	rsp.GameArgMsgName, rsp.GameArgMsgValue, _ = protobuf.Marshal(roomArg)
	rsp.Code = 1

	return nil
}
