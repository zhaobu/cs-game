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

func (p *friend) DeleteRelationReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
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
	req, ok := pb.(*friendpb.DeleteRelationReq)
	if !ok {
		return fmt.Errorf("not friendpb.DeleteRelationReq")
	}

	Log.Infof("args info:name=%s,fromid=%d,detail=%v", args.PayloadName, args.FromUID, req)

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
