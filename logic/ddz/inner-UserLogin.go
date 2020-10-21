package main

import (
	"context"
	"game/codec"
	"game/logic/ddz/desk"
	"game/pb/inner"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

// gate通知玩家登陆
func (p *ddz) UserLogin(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbinner.UserLogin)
	if !ok {
		err = fmt.Errorf("not *pbinner.UserLogin")
		log.Error(err.Error())
		return
	}

	log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("recv %s %+v ", args.Name, *req)

	defer func() {
		if r := recover(); r != nil {
			log.WithFields(logrus.Fields{"uid": args.UserID}).Warnf("r:%v stack:%s", r, string(debug.Stack()))
		}
	}()

	desk.UserLogin(req.UserID)

	return nil
}
