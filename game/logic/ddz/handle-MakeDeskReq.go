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

func (p *ddz) MakeDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	makeDeskReq, ok := pb.(*pbgame.MakeDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.MakeDeskReq")
		logrus.Error(err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{"req": *makeDeskReq, "name": args.Name}).Info("recv")

	defer func() {
		r := recover()
		if r != nil {
			logrus.WithFields(logrus.Fields{
				"r":     r,
				"stack": string(debug.Stack()),
			}).Info(args.Name)
		}
	}()

	desk.MakeDesk(args.UserID, makeDeskReq)
	return nil
}
