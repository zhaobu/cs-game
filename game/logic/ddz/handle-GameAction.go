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

func (p *ddz) GameAction(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	gameAction, ok := pb.(*pbgame.GameAction)
	if !ok {
		err = fmt.Errorf("not *pbgame.GameAction")
		logrus.Error(err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{"req": *gameAction, "name": args.Name}).Info("recv")

	defer func() {
		r := recover()
		if r != nil {
			logrus.WithFields(logrus.Fields{
				"r":     r,
				"stack": string(debug.Stack()),
			}).Info(args.Name)
		}
	}()

	desk.Action(args.UserID, gameAction.ActionName, gameAction.ActionValue)
	return nil
}
