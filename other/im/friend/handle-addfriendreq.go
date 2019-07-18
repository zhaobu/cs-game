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
	"time"
)

func (p *friend) AddFriendReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			Log.Errorf("recover info, err=%s,stack info:%s", err, string(debug.Stack()))
		}
	}()

	pb, err := protobuf.Unmarshal(args.PayloadName, args.Payload)
	if err != nil {
		return err
	}
	req, ok := pb.(*friendpb.AddFriendReq)
	if !ok {
		return fmt.Errorf("not friendpb.AddFriendReq")
	}

	Log.Infof("args info:name=%s,fromid=%d,detail=%v", args.PayloadName, args.FromUID, req)

	if req.Source == req.Target || req.Source != args.FromUID {
		Log.Infof("bad args info:name=%s,fromid=%d,detail=%v", args.PayloadName, args.FromUID, req)
		return
	}

	pending, err := cache.AddFriendPending(req.Source, req.Target)
	if err != nil || pending {
		Log.Infof("AddFriendPending info:name=%s,detail=%v,err=%s", args.PayloadName, req, err)
		return
	}

	fs, err := cache.UserFriend(req.Source)
	if err == nil {
		for _, fid := range fs {
			if fid == req.Target {
				Log.Infof("areadly friend:source=%d,target=%d", req.Source, req.Target)
				return
			}
		}
	}

	{
		addFriendNotif := &db.AddFriendNotif{
			StoreKey:   fmt.Sprintf("uid:%d", req.Target),
			MsgID:      time.Now().UTC().UnixNano(),
			Target:     req.Target,
			Source:     req.Source,
			Msg:        req.Msg,
			InviteTime: time.Now().UTC().UnixNano(),
		}
		db.BatchWriteAddFriendNotif(tsdbCli, []*db.AddFriendNotif{addFriendNotif})
	}

	{
		if queryPlace(req.Target) == "" {
			Log.Infof("not online:target=%d", req.Target)
			return
		}

		addFriendNotif := &friendpb.AddFriendNotif{}
		addFriendNotif.Source = req.Source
		addFriendNotif.Target = req.Target
		addFriendNotif.Msg = req.Msg

		pay := codec.NewMsgPayload()
		pay.ToUID = req.Target
		var err error
		pay.PayloadName, pay.Payload, err = protobuf.Marshal(addFriendNotif)
		if err == nil {
			cliGate.Go(context.Background(), "BackEnd", pay, nil, nil)
		}
	}

	return nil
}
