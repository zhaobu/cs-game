package main

import (
	"context"
	"cy/game/codec"
	"cy/game/logic/ddz/desk"
	"cy/game/pb/game"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (p *ddz) JoinDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbgame.JoinDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.JoinDeskReq")
		log.Error(err.Error())
		return
	}

	log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("recv %s %+v ", args.Name, *req)

	defer func() {
		r := recover()
		if r != nil {
			log.WithFields(logrus.Fields{"uid": args.UserID}).Warnf("r:%v stack:%s", r, string(debug.Stack()))
		}
	}()

	desk.JoinDesk(args.UserID, req)
	return nil
}
