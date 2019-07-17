package main

import (
	"context"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	"cy/other/im/pb"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (p *logic) LoginReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {

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
		}).Info()
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
