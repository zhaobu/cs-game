package main

import (
	"context"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	"cy/other/im/inner"
	"cy/other/im/logic/db"
	"cy/other/im/pb"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/sirupsen/logrus"
)

func (p *logic) MsgRecordReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *impb.MsgRecordReq
	var ok bool
	rsp := &impb.MsgRecordRsp{}

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
	req, ok = pb.(*impb.MsgRecordReq)
	if !ok {
		return fmt.Errorf("not impb.MsgRecordReq")
	}
	if req.From != args.FromUID || req.To != args.ToUID {
		return fmt.Errorf("bad args from:%d-%d, to:%d-%d", req.From, args.FromUID, req.To, args.ToUID)
	}

	storeKey := inner.StoreKey(req.From)
	startMsgID, err := strconv.ParseInt(req.StartMsgID, 10, 64)
	if err != nil {
		return err
	}
	endMsgID, err := strconv.ParseInt(req.EndMsgID, 10, 64)
	if err != nil {
		return err
	}
	if startMsgID > endMsgID {
		return fmt.Errorf("start id > end id")
	}
	if req.Limit < 0 || req.Limit > 50 {
		req.Limit = 50
	}

	result, err := db.RangeGetMsgRecord(storeKey, startMsgID, endMsgID, req.Limit)
	if err != nil {
		return err
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
