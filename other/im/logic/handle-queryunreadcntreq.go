package main

import (
	"context"
	"cy/other/im/cache"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	impb "cy/other/im/pb"
	"fmt"
	"runtime/debug"
)

func (p *logic) QueryUnreadCntReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *impb.QueryUnreadCntReq
	var ok bool
	rsp := &impb.QueryUnreadCntRsp{}

	defer func() {
		r := recover()
		if r != nil {
			log.Errorf("recover info,fromid=%d,toid=%d,flag=%v,plname=%s,req=%v,rsp=%v,err=%s,r=%s,stack=%s", args.FromUID, args.ToUID, args.Flag, args.PayloadName, req, rsp, err, r, string(debug.Stack()))
		}
	}()

	pb, err := protobuf.Unmarshal(args.PayloadName, args.Payload)
	if err != nil {
		return err
	}
	req, ok = pb.(*impb.QueryUnreadCntReq)
	if !ok {
		return fmt.Errorf("not impb.QueryUnreadCntReq")
	}

	rsp.Seq = req.Seq
	rsp.Cnt, err = cache.UnreadCnt(args.FromUID)
	if err != nil {
		return err
	}
	reply.Seq = args.Seq
	reply.Flag = args.Flag
	reply.PayloadName, reply.Payload, err = protobuf.Marshal(rsp)
	if err != nil {
		return err
	}

	return nil
}
