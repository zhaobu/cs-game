package main
import (
	"context"
	"cy/other/im/cache"
	"cy/other/im/codec"
	"cy/other/im/codec/protobuf"
	. "cy/other/im/common/logger"
	"cy/other/im/friend/db"
	friendpb "cy/other/im/pb/friend"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"
)

func (p *friend) AddFriendNotifAck(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	var req *friendpb.AddFriendNotifAck
	var ok bool

	defer func() {
		r := recover()
		if r != nil {
			Log.Errorf("recover info,fromid=%d,toid=%d,flag=%v,req=%v,err=%v,stack=%s", args.FromUID, args.ToUID, args.Flag, req, err, string(debug.Stack()))
		}
	}()

	pb, err := protobuf.Unmarshal(args.PayloadName, args.Payload)
	if err != nil {
		return err
	}
	req, ok = pb.(*friendpb.AddFriendNotifAck)
	if !ok {
		return fmt.Errorf("not friendpb.AddFriendNotifAck")
	}

	if req.Code == friendpb.FriendAction_Agree {
		cache.AddFriend(req.Source, req.Target)
		cache.AddFriend(req.Target, req.Source)
	}

	cache.DeleteFriendPending(req.Source, req.Target)

	// 回应
	if req.Code != friendpb.FriendAction_Readed {
		addFriendResult := &friendpb.AddFriendResult{}
		addFriendResult.Source = req.Source
		addFriendResult.Target = req.Target
		addFriendResult.Msg = req.Msg
		addFriendResult.Code = req.Code
		addFriendResult.MsgID = req.MsgID

		reply.Seq = args.Seq
		reply.ToUID = req.Target // 回应给被加人
		reply.PayloadName, reply.Payload, err = protobuf.Marshal(addFriendResult)
		if err != nil {

		}
	}

	// 被加人删除 AddFriendNotif
	msgid, _ := strconv.ParseInt(req.MsgID, 10, 64)
	db.DeleteAddFriendNotif(tsdbCli, fmt.Sprintf("uid:%d", req.Target), msgid)

	// 邀请人添加 AddFriendResult
	addFriendResultDB := &db.AddFriendResult{
		StoreKey:   fmt.Sprintf("uid:%d", req.Source),
		MsgID:      time.Now().UTC().UnixNano(), // 用新的ID
		Target:     req.Target,
		Source:     req.Source,
		Msg:        req.Msg,
		InviteTime: time.Now().UTC().UnixNano(),
		Code:       int64(req.Code),
	}

	db.BatchWriteAddFriendResult(tsdbCli, []*db.AddFriendResult{addFriendResultDB})

	// 通知邀请人
	{
		if queryPlace(req.Source) == "" {
			Log.Infof("not online:source=%d", req.Source)
			return nil
		}

		addFriendResult := &friendpb.AddFriendResult{}
		addFriendResult.Source = req.Source
		addFriendResult.Target = req.Target
		addFriendResult.Msg = req.Msg
		addFriendResult.Code = req.Code
		addFriendResult.MsgID = strconv.FormatInt(addFriendResultDB.MsgID, 10)

		pay := codec.NewMsgPayload()
		pay.ToUID = req.Source // 邀请人
		var err error
		pay.PayloadName, pay.Payload, err = protobuf.Marshal(addFriendResult)
		if err != nil {
			return err
		}

		cliGate.Go(context.Background(), "BackEnd", pay, nil, nil)
	}

	return nil
}

