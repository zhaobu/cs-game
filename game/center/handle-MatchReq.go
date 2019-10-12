package main

import (
	"context"
	"game/codec"
	"game/codec/protobuf"
	pbcenter "game/pb/center"
	pbcommon "game/pb/common"
	pbinner "game/pb/inner"
	"fmt"
	"runtime/debug"
)

func (p *center) MatchReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	matchReq, ok := pb.(*pbcenter.MatchReq)
	if !ok {
		err = fmt.Errorf("not *pbcenter.MatchReq")
		tlog.Error(err.Error())
		return
	}

	// ###############################
	matchRsp := &pbcenter.MatchRsp{}
	if matchReq.Head != nil {
		matchRsp.Head = &pbcommon.RspHead{Seq: matchReq.Head.Seq}
	}

	defer func() {
		r := recover()
		stack := ""
		if r != nil {
			stack = string(debug.Stack())
		}

		codec.Pb2Msg(matchRsp, reply)

		log.Infof("args.Name:%s:req:%v,rsp:%v,err:%v,r:%v,stack:%v", args.Name, *matchReq, *matchRsp, err, r, stack)
	}()

	// 1> 向逻辑游戏检查参数
	cli, err := getGameCli(matchReq.GameName)
	if err != nil {
		matchRsp.Code = pbcenter.MatchRspCode_InvalidGame
		matchRsp.StrCode = err.Error()
		log.Warnf("bad gamename %s", matchReq.GameName)
		return nil
	}

	reqRCall := &codec.Message{}
	codec.Pb2Msg(&pbinner.GameMatchArgsCheckReq{RoomId: matchReq.RoomId, UserID: args.UserID}, reqRCall)
	rspRCall := &codec.Message{}

	err = cli.Call(context.Background(), "GameMatchArgsCheckReq", reqRCall, rspRCall)
	if err != nil {
		matchRsp.Code = pbcenter.MatchRspCode_InternalServerError
		matchRsp.StrCode = err.Error()
		tlog.Error(err.Error())
		return nil
	}

	pbRCall, err := protobuf.Unmarshal(rspRCall.Name, rspRCall.Payload)
	if err != nil {
		matchRsp.Code = pbcenter.MatchRspCode_InternalServerError
		matchRsp.StrCode = err.Error()
		tlog.Error(err.Error())
		return nil
	}

	gameMatchArgsCheckRsp := pbRCall.(*pbinner.GameMatchArgsCheckRsp)
	switch gameMatchArgsCheckRsp.Code {
	case 2:
		matchRsp.Code = pbcenter.MatchRspCode_InvalidRoomID
	case 3:
		matchRsp.Code = pbcenter.MatchRspCode_InternalServerError
	case 4:
		matchRsp.Code = pbcenter.MatchRspCode_NotEnoughMoney
	}

	matchRsp.GameArgMsgName = gameMatchArgsCheckRsp.GameArgMsgName
	matchRsp.GameArgMsgValue = gameMatchArgsCheckRsp.GameArgMsgValue

	if gameMatchArgsCheckRsp.Code != 1 {
		return
	}

	// 2> 进入匹配
	enterMatch(args.UserID, matchReq.GameName, matchReq.RoomId, matchRsp)

	return nil
}
