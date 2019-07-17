package main

import (
	"context"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	"cy/other/im/friend/db/notif"
	"cy/other/im/friend/db/result"
	"cy/other/im/friend/pb"
	"cy/other/im/pb/misc"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/sirupsen/logrus"
)

func (p *friend) QueryInbox(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			logrus.WithFields(logrus.Fields{
				"stack": string(debug.Stack()),
			}).Error()
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

	logrus.WithFields(logrus.Fields{
		"name":   args.PayloadName,
		"fromid": args.FromUID,
		"detail": req,
	}).Info("handle")

	gm := &misc.GroupMsg{}

	// 1>
	startMsgID, _ := strconv.ParseInt(req.StartMsgID, 10, 64)
	findResult1, err := notif.RangeGetAddFriendNotif(
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
		logrus.WithFields(logrus.Fields{}).Warn(err)
	}

	// 2>
	findResult2, err := result.RangeGetAddFriendResult(
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
		logrus.WithFields(logrus.Fields{}).Warn(err)
	}

	logrus.WithFields(logrus.Fields{
		"len gm.Msgs": len(gm.Msgs),
		"gm.Msgs":     gm.Msgs,
	}).Info()

	if len(gm.Msgs) > 0 {
		reply.Seq = args.Seq
		reply.ToUID = args.FromUID
		reply.PayloadName, reply.Payload, err = protobuf.Marshal(gm)
		if err != nil {
			logrus.WithFields(logrus.Fields{}).Warn(err)
		}
	}

	return nil
}
