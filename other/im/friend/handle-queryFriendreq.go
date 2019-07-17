package main

import (
	"context"
	"cy/other/im/cache"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	friendpb "cy/other/im/pb/friend"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (p *friend) QueryFriendReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
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
	req, ok := pb.(*friendpb.QueryFriendReq)
	if !ok {
		return fmt.Errorf("not friendpb.QueryFriendReq")
	}

	logrus.WithFields(logrus.Fields{
		"name":   args.PayloadName,
		"fromid": args.FromUID,
		"detail": req,
	}).Info()

	rsp := &friendpb.QueryFriendRsp{}
	rsp.Seq = req.Seq
	rsp.Friends = make(map[uint64]bool)

	fs, err := cache.UserFriend(args.FromUID)
	if err != nil {
		logrus.Warn(err)
		return
	}

	places, _ := cache.QueryUsers(fs...)

	for _, fid := range fs {
		rsp.Friends[fid] = places[fid] != ""
	}

	logrus.WithFields(logrus.Fields{
		"user": args.FromUID,
		"rsp":  rsp,
	}).Info("batchQueryPlace")

	reply.Seq = args.Seq
	reply.PayloadName, reply.Payload, err = protobuf.Marshal(rsp)
	if err != nil {
		return err
	}

	return nil
}
