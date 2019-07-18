package main

import (
	"context"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	. "cy/other/im/common/logger"
	"cy/other/im/friend/db"
	friendpb "cy/other/im/pb/friend"
	"cy/other/im/pb/misc"
	"fmt"
	"runtime/debug"
	"strconv"

	"go.uber.org/zap"
)

func (p *friend) QueryInbox(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			Log.Errorf("recover info:%s", string(debug.Stack()))
		}
	}()

	pb, err := protobuf.Unmarshal(args.PayloadName, args.Payload)
	if err != nil {
		return err
	}
	req, ok := pb.(*friendpb.QueryInbox)
	if !ok {
		return fmt.Errorf("not friendpb.QueryInbox")
	}

	Tlog.Info("handle info", zap.String("name", args.PayloadName), zap.Uint64("fromid", args.FromUID), zap.Any("detail", req))
	gm := &misc.GroupMsg{}

	// 1>
	startMsgID, _ := strconv.ParseInt(req.StartMsgID, 10, 64)
	findResult1, err := db.RangeGetAddFriendNotif(
		tsdbCli,
		fmt.Sprintf("uid:%d", args.FromUID),
		startMsgID,
		req.Limit,
	)
	if err == nil {
		for _, notifV := range findResult1 {
			addFriendNotif := &friendpb.AddFriendNotif{}
			addFriendNotif.Source = notifV.Source
			addFriendNotif.Target = notifV.Target
			addFriendNotif.Msg = notifV.Msg
			addFriendNotif.InviteTime = notifV.InviteTime
			addFriendNotif.MsgID = strconv.FormatInt(notifV.MsgID, 10)

			gm = protobuf.GroupAppend(gm, addFriendNotif).(*misc.GroupMsg)
		}
	} else {
		Tlog.Error("err info ", zap.Error(err))
	}

	// 2>
	findResult2, err := db.RangeGetAddFriendResult(
		tsdbCli,
		fmt.Sprintf("uid:%d", args.FromUID),
		startMsgID,
		req.Limit,
	)
	if err == nil {
		for _, resultV := range findResult2 {
			addFriendResult := &friendpb.AddFriendResult{}
			addFriendResult.Source = resultV.Source
			addFriendResult.Target = resultV.Target
			addFriendResult.Msg = resultV.Msg
			addFriendResult.Code = friendpb.FriendAction(resultV.Code)
			addFriendResult.MsgID = strconv.FormatInt(resultV.MsgID, 10)

			gm = protobuf.GroupAppend(gm, addFriendResult).(*misc.GroupMsg)
		}
	} else {
		Tlog.Error("err info ", zap.Error(err))
	}

	Tlog.Info("gm info", zap.Int("len gm.Msgs", len(gm.Msgs)), zap.Any("gm.Msgs", gm.Msgs))

	if len(gm.Msgs) > 0 {
		reply.Seq = args.Seq
		reply.ToUID = args.FromUID
		reply.PayloadName, reply.Payload, err = protobuf.Marshal(gm)
		if err != nil {
			Tlog.Error("err info ", zap.Error(err))
		}
	}

	return nil
}
