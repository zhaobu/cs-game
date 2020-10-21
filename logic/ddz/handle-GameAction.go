package main

import (
	"context"
	"game/codec"
	"game/logic/ddz/desk"
	pbgame "game/pb/game"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (p *ddz) GameAction(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbgame.GameAction)
	if !ok {
		err = fmt.Errorf("not *pbgame.GameAction")
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

	desk.Action(args.UserID, req.ActionName, req.ActionValue)
	return nil
}
