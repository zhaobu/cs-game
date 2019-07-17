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

func (p *friend) DeleteRelationReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
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
	req, ok := pb.(*friendpb.DeleteRelationReq)
	if !ok {
		return fmt.Errorf("not friendpb.DeleteRelationReq")
	}

	logrus.WithFields(logrus.Fields{
		"name":   args.PayloadName,
		"fromid": args.FromUID,
		"detail": req,
	}).Info()

	if req.MyID != args.FromUID {
		//return
	}

	cache.DelFriend(req.MyID, req.DelID)
	cache.DelFriend(req.DelID, req.MyID)

	rsp := &friendpb.DeleteRelationRsp{}
	rsp.Seq = req.Seq
	rsp.Code = 1
	rsp.DelID = req.DelID
	reply.Seq = args.Seq
	reply.PayloadName, reply.Payload, err = protobuf.Marshal(rsp)
	if err != nil {
		return err
	}

	return nil
}
