package main

import (
	"context"
	"cy/im/cache"
	"cy/im/codec"
	"cy/im/codec/protobuf"
	"cy/im/friend/db/notif"
	"cy/im/friend/pb"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

func (p *friend) AddFriendReq(ctx context.Context, args *codec.MsgPayload, reply *codec.MsgPayload) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			logrus.WithFields(logrus.Fields{
				"stack": string(debug.Stack()),
			}).Error()
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

	logrus.WithFields(logrus.Fields{
		"name":   args.PayloadName,
		"fromid": args.FromUID,
		"detail": req,
	}).Info("handle")

	if req.Source == req.Target || req.Source != args.FromUID {
		logrus.WithFields(logrus.Fields{
			"name":   args.PayloadName,
			"fromid": args.FromUID,
			"detail": req,
		}).Info("bad args")
		return
	}

	pending, err := cache.AddFriendPending(req.Source, req.Target)
	if err != nil || pending {
		logrus.WithFields(logrus.Fields{
			"name":   args.PayloadName,
			"detail": req,
			"err":    err,
		}).Info("AddFriendPending")
		return
	}

	fs, err := cache.UserFriend(req.Source)
	if err == nil {
		for _, fid := range fs {
			if fid == req.Target {
				logrus.WithFields(logrus.Fields{"source": req.Source, "target": req.Target}).Info("areadly friend")
				return
			}
		}
	}

	{
		addFriendNotif := &notif.AddFriendNotif{
			StoreKey:   fmt.Sprintf("uid:%d", req.Target),
			MsgID:      time.Now().UTC().UnixNano(),
			Target:     req.Target,
			Source:     req.Source,
			Msg:        req.Msg,
			InviteTime: time.Now().UTC().UnixNano(),
		}
		notif.BatchWriteAddFriendNotif(tsdbCli, []*notif.AddFriendNotif{addFriendNotif})
	}

	{
		if queryPlace(req.Target) == "" {
			logrus.WithFields(logrus.Fields{"target": req.Target}).Info("not online")
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
