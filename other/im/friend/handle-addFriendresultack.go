package main

import (
	"context"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	"cy/other/im/friend/db/result"
	"cy/other/im/friend/pb"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/sirupsen/logrus"
)

func (p *friend) AddFriendResultAck(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
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
	req, ok := pb.(*friendpb.AddFriendResultAck)
	if !ok {
		return fmt.Errorf("not friendpb.AddFriendResultAck")
	}

	logrus.WithFields(logrus.Fields{
		"name":   args.PayloadName,
		"fromid": args.FromUID,
		"detail": req,
	}).Info("handle")

	msgID, _ := strconv.ParseInt(req.MsgID, 10, 64)
	result.DeleteAddFriendResult(tsdbCli, fmt.Sprintf("uid:%d", args.FromUID), msgID)
	return nil
}
