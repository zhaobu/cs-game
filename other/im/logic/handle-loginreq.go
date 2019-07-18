package main

import (
	"context"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	impb "cy/other/im/pb"
	"fmt"
	"runtime/debug"
)

func (p *logic) LoginReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {

	defer func() {
		r := recover()
		if r != nil {
			log.Errorf("recover info,fromid=%d,toid=%d,flag=%v,plname=%s,err=%s,r=%s,stack=%s", args.FromUID, args.ToUID, args.Flag, args.PayloadName, err, r, string(debug.Stack()))
		}
	}()

	pb, err := protobuf.Unmarshal(args.PayloadName, args.Payload)
	if err != nil {
		return err
	}
	_, ok := pb.(*impb.LoginReq)
	if !ok {
		return fmt.Errorf("not impb.LoginReq")
	}

	reply.Seq = args.Seq
	reply.Flag = args.Flag
	reply.PayloadName, reply.Payload, err = protobuf.Marshal(&impb.LoginRsp{})
	if err != nil {
		return err
	}

	return nil
}
