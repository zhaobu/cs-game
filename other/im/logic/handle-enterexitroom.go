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

func (p *logic) EnterExitRoom(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *impb.EnterExitRoom
	var ok bool

	defer func() {
		r := recover()
		if r != nil {
			log.Errorf("recover info,fromid=%d,toid=%d,flag=%v,plname=%s,req=%v,err=%s,r=%s,stack=%s", args.FromUID, args.ToUID, args.Flag, args.PayloadName, req, err, r, string(debug.Stack()))
		}
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
