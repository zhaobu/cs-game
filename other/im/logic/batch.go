package main

import (
	"context"
	"cy/other/im/cache"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	"cy/other/im/inner"
	impb "cy/other/im/pb"
	"cy/other/im/pb/misc"
	"strconv"
	"sync"
	"time"

	"github.com/aperdana/batcher"
)

var (
	mubat sync.Mutex
	bat   = make(map[string]*batcher.Batcher) // key: storeKey
)

func getBatcher(storeKey string, timeout time.Duration) *batcher.Batcher {
	mubat.Lock()
	defer mubat.Unlock()

	b, ok := bat[storeKey]
	if ok {
		return b
	}

	b = batcher.New(batchOperator, timeout, batcher.SetMaxBatchSize(10))
	b.Listen()
	bat[storeKey] = b
	return b
}

func batchOperator(reqs []interface{}) {
	if len(reqs) == 0 {
		return
	}

	var batchReqs []*ChatMsg

	unreadCnt := make(map[uint64]int64)
	gm := &misc.GroupMsg{}
	var toUID uint64

	for _, req := range reqs {
		batchReq, ok := req.(*ChatMsg)
		if !ok {
			continue
		}
		batchReqs = append(batchReqs, batchReq)

		// 来自发送者的不用通知了
		if batchReq.StoreKey == inner.StoreKey(batchReq.From) {
			continue
		}

		toUID = batchReq.To

		unreadCnt[batchReq.From]++

		mn := &impb.MsgNotify{}
		mn.MsgID = strconv.FormatInt(batchReq.MsgID, 10)
		mn.To = batchReq.To
		mn.From = batchReq.From
		mn.Content = batchReq.Content
		mn.Ct = impb.ContentType(batchReq.Ct)
		mn.SendTime = batchReq.SentTime

		gm = protobuf.GroupAppend(gm, mn).(*misc.GroupMsg)
	}

	BatchWriteChatMsg(batchReqs)

	if toUID != 0 {
		cache.ChangeUnreadCnt(toUID, unreadCnt)
		log.Infof("change unread cnt,uid=%d,unreadCnt=%v", toUID, unreadCnt)

		if queryPlace(toUID) != "" {
			log.Infof("online,touid=%d", toUID)
			pntf := codec.NewMsgPayload()
			pntf.ToUID = toUID
			var err error
			pntf.PayloadName, pntf.Payload, err = protobuf.Marshal(gm)
			if err == nil {
				cliGate.Go(context.Background(), "BackEnd", pntf, nil, nil)
			} else {
				log.Infof("offline,touid=%d", toUID)
			}
		}
	}
}
