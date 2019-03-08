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

func (p *logic) EnterExitRoom(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *impb.EnterExitRoom
	var ok bool

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
	req, ok = pb.(*impb.EnterExitRoom)
	if !ok {
		return fmt.Errorf("not impb.EnterExitRoom")
	}

	if req.EnterOrExit == 1 {
		cache.UserEnterRoom(req.UID, req.RoomID)
	} else if req.EnterOrExit == 2 {
		cache.UserExitRoom(req.UID, req.RoomID)
	}

	return nil
}
