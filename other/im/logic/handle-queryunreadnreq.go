package main

import (
	"context"
	"cy/other/im/cache"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	. "cy/other/im/common/logger"
	"cy/other/im/inner"

	impb "cy/other/im/pb"
	"fmt"
	"runtime/debug"
	"strconv"
)

func (p *logic) QueryUnreadNReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *impb.QueryUnreadNReq
	var ok bool
	rsp := &impb.QueryUnreadNRsp{}

	defer func() {
		r := recover()
		if r != nil {
			Log.Errorf("recover info,fromid=%d,toid=%d,flag=%v,plname=%s,req=%v,rsp=%v,err=%s,r=%s,stack=%s", args.FromUID, args.ToUID, args.Flag, args.PayloadName, req, rsp, err, r, string(debug.Stack()))
		}
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

	result, err := RangeGetBySessionKey(storeKey, sessID, startMsgID, int32(req.LastN))

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
