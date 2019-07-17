package main

import (
	"context"
	"cy/other/im/cache"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	"cy/other/im/pb"
	"fmt"
	"runtime/debug"
	"sort"
	"strconv"

	"github.com/sirupsen/logrus"
)

func (p *logic) MsgNotifyAck(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *impb.MsgNotifyAck
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
			"req":    req,
		}).Info()
	}()

	pb, err := protobuf.Unmarshal(args.PayloadName, args.Payload)
	if err != nil {
		return err
	}
	req, ok = pb.(*impb.MsgNotifyAck)
	if !ok {
		return fmt.Errorf("not impb.MsgNotifyAck")
	}
	if len(req.MsgIDs) == 0 || req.OtherUID == 0 {
		return fmt.Errorf("arg bad")
	}

	// 按照请求中确认消息多少直接减少未读数量，不做任何判断
	cnt := make(map[uint64]int64)
	cnt[req.OtherUID] = (int64(len(req.MsgIDs)) * -1)
	if err2 := cache.ChangeUnreadCnt(args.FromUID, cnt); err2 != nil {
		logrus.Warn(err2)
	}

	// 排序后取最大的msgid
	sort.Slice(req.MsgIDs, func(i, j int) bool {
		return req.MsgIDs[i] > req.MsgIDs[i]
	})

	// 不判断消息ID合法（是否有效、是否对应此otherid）
	if lastid, err3 := strconv.ParseInt(req.MsgIDs[0], 10, 64); err3 != nil {
		logrus.Warn(err3)
	} else {
		cache.SetLastReadID(args.FromUID, req.OtherUID, lastid)
	}

	return nil
}
