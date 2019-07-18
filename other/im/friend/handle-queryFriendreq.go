package main

import (
	"context"
	"cy/other/im/cache"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	. "cy/other/im/common/logger"
	friendpb "cy/other/im/pb/friend"
	"fmt"
	"runtime/debug"
)

func (p *friend) QueryFriendReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
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
	req, ok := pb.(*friendpb.QueryFriendReq)
	if !ok {
		return fmt.Errorf("not friendpb.QueryFriendReq")
	}

	Log.Infof("args info:name=%s,fromid=%d,detail=%v", args.PayloadName, args.FromUID, req)

	rsp := &friendpb.QueryFriendRsp{}
	rsp.Seq = req.Seq
	rsp.Friends = make(map[uint64]bool)

	fs, err := cache.UserFriend(args.FromUID)
	if err != nil {
		Log.Warn(err.Error())
		return
	}

	places, _ := cache.QueryUsers(fs...)

	for _, fid := range fs {
		rsp.Friends[fid] = places[fid] != ""
	}

	Log.Infof("batchQueryPlace info:user=%d,rsp=%v", args.FromUID, rsp)

	reply.Seq = args.Seq
	reply.PayloadName, reply.Payload, err = protobuf.Marshal(rsp)
	if err != nil {
		return err
	}

	return nil
}
