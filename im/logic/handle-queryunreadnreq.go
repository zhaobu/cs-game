package main

import (
	"context"
	"cy/im/cache"
	"cy/im/codec"
	"cy/im/codec/protobuf"
	"cy/im/inner"
	"cy/im/logic/db"
	"cy/im/pb"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/sirupsen/logrus"
)

func (p *logic) QueryUnreadNReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *impb.QueryUnreadNReq
	var ok bool
	rsp := &impb.QueryUnreadNRsp{}

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
	req, ok = pb.(*impb.QueryUnreadNReq)
	if !ok {
		return fmt.Errorf("not impb.QueryUnreadNReq")
	}

	if req.LastN < 0 || req.LastN > 30 {
		req.LastN = 30
	}

	storeKey := inner.StoreKey(args.FromUID)
	sessID := inner.SessionID(args.FromUID, req.OtherUID, args.IsBroadCast(), args.IsMultiCast())
	startMsgID, _ := cache.LastReadID(args.FromUID, req.OtherUID)

	result, err := db.RangeGetBySessionKey(storeKey, sessID, startMsgID, int32(req.LastN))

	if err != nil || len(result) == 0 {
		return
	}

	rsp.Seq = req.Seq
	for _, v := range result {
		mn := &impb.MsgNotify{}
		mn.MsgID = strconv.FormatInt(v.MsgID, 10)
		mn.To = v.To
		mn.From = v.From
		mn.Content = v.Content
		mn.Ct = impb.ContentType(v.Ct)
		mn.SendTime = v.SentTime

		rsp.Msgs = append(rsp.Msgs, mn)
	}

	reply.Seq = args.Seq
	reply.Flag = args.Flag
	reply.PayloadName, reply.Payload, err = protobuf.Marshal(rsp)
	if err != nil {
		return err
	}

	return nil

}
