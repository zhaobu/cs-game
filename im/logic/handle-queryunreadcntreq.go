package main

import (
	"context"
	"cy/im/cache"
	"cy/im/codec"
	"cy/im/codec/protobuf"
	"cy/im/pb"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (p *logic) QueryUnreadCntReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *impb.QueryUnreadCntReq
	var ok bool
	rsp := &impb.QueryUnreadCntRsp{}

	defer func() {
		stack := ""
		r := recover()
		if r != nil {
			stack = string(debug.Stack())
		}
		logrus.WithFields(logrus.Fields{
			"fromid": args.FromUID,
			"toid":   args.ToUID,
			"flag":   args.Flag,
			"plname": args.PayloadName,
			"err":    err,
			"r":      r,
			"stack":  stack,
			"req":    req,
			"rsp":    rsp,
		}).Info()
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
