package main

import (
	"context"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	. "cy/other/im/common/logger"
	"cy/other/im/friend/db"
	friendpb "cy/other/im/pb/friend"
	"fmt"
	"runtime/debug"
	"strconv"
)

func (p *friend) AddFriendResultAck(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			Log.Errorf("recover info, err=%s,stack info:%s", err, string(debug.Stack()))
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

	Log.Infof("args info:name=%s,fromid=%d,detail=%v", args.PayloadName, args.FromUID, req)

	msgID, _ := strconv.ParseInt(req.MsgID, 10, 64)
	db.DeleteAddFriendResult(tsdbCli, fmt.Sprintf("uid:%d", args.FromUID), msgID)
	return nil
}
