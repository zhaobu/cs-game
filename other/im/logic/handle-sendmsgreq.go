package main

import (
	"context"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	"cy/other/im/inner"

	impb "cy/other/im/pb"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"
)

func (p *logic) SendMsgReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *impb.SendMsgReq
	var ok bool
	rsp := &impb.SendMsgRsp{}

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
	req, ok = pb.(*impb.SendMsgReq)
	if !ok {
		return fmt.Errorf("not impb.SendMsgReq")
	}
	if req.From != args.FromUID {
		return fmt.Errorf("bad args from:%d-%d", req.From, args.FromUID)
	}
	if req.To != args.ToUID {
		return fmt.Errorf("bad args to:%d-%d", req.To, args.ToUID)
	}
	if req.To == req.From {
		return fmt.Errorf("to == from", req.To, req.From)
	}
	if req.To == 0 {
		return fmt.Errorf("bad args to:%d", req.To)
	}

	ut := time.Now().UTC().UnixNano()

	// 这2种类型 就不reply了
	if args.Flag.IsBroadCast() || args.Flag.IsMultiCast() {
		mn := &impb.MsgNotify{}
		mn.MsgID = strconv.FormatInt(ut, 10)
		mn.To = req.To
		mn.From = req.From
		mn.Content = req.Content
		mn.Ct = req.Ct
		mn.SendTime = ut

		pay := codec.NewMsgPayload()
		pay.Seq = args.Seq
		pay.FromUID = args.FromUID
		pay.ToUID = args.ToUID
		pay.Flag = args.Flag
		var err2 error
		pay.PayloadName, pay.Payload, err2 = protobuf.Marshal(mn)
		if err2 == nil {
			serviceMethod := "BroadCast"
			if args.Flag.IsMultiCast() {
				serviceMethod = "MultiCast"
			}
			cliGate.Broadcast(context.Background(), serviceMethod, pay, nil)
		}
		return
	}

	storeKeys := []string{inner.StoreKey(req.To), inner.StoreKey(req.From)}
	sessID := inner.SessionID(req.From, req.To, args.IsBroadCast(), args.IsMultiCast())

	for _, sk := range storeKeys {
		req := &ChatMsg{
			StoreKey:   sk,
			MsgID:      ut,
			SessionKey: sessID,
			To:         req.To,
			From:       req.From,
			GroupID:    0, // 1对1 消息
			Content:    req.Content,
			Ct:         int64(req.Ct),
			SentTime:   ut,
		}

		b := getBatcher(sk, time.Millisecond*500)
		b.Batch(req)
	}

	rsp.Seq = req.Seq
	rsp.Code = 1
	rsp.MsgID = strconv.FormatInt(ut, 10)

	reply.Seq = args.Seq
	reply.Flag = args.Flag
	reply.PayloadName, reply.Payload, err = protobuf.Marshal(rsp)
	if err != nil {
		return err
	}

	return nil
}
