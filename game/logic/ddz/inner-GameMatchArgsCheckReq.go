package main

import (
	"context"
	"cy/game/codec"
	"cy/game/codec/protobuf"
	"cy/game/logic/ddz/desk"
	"cy/game/pb/inner"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (p *ddz) GameMatchArgsCheckReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	gameMatchArgsCheckReq, ok := pb.(*pbinner.GameMatchArgsCheckReq)
	if !ok {
		err = fmt.Errorf("not *pbinner.GameMatchArgsCheckReq")
		logrus.Error(err.Error())
		return
	}

	gameMatchArgsCheckRsp := &pbinner.GameMatchArgsCheckRsp{}

	defer func() {
		r := recover()
		stack := ""
		if r != nil {
			stack = string(debug.Stack())
		}

		codec.Pb2Msg(gameMatchArgsCheckRsp, reply)

		logrus.WithFields(logrus.Fields{
			"req":   *gameMatchArgsCheckReq,
			"rsp":   *gameMatchArgsCheckRsp,
			"err":   err,
			"r":     r,
			"stack": stack,
			"name":  args.Name,
		}).Info("recv")
	}()

	// ###############################################
	// 取参数
	roomArg := desk.QueryMatchRoomArg(gameMatchArgsCheckReq.RoomId)
	if roomArg == nil || !roomArg.Enable {
		gameMatchArgsCheckRsp.Code = 2
		return
	}

	uinfo, err := desk.QueryUserInfo(gameMatchArgsCheckReq.UserID)
	if err != nil {
		gameMatchArgsCheckRsp.Code = 3
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
		gameMatchArgsCheckRsp.Code = 4
		return
	}

	// 一同返回桌子参数
	gameMatchArgsCheckRsp.GameArgMsgName, gameMatchArgsCheckRsp.GameArgMsgValue, _ = protobuf.Marshal(roomArg)
	gameMatchArgsCheckRsp.Code = 1

	return nil
}
