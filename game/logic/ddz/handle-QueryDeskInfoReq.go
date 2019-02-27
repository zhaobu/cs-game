package main

import (
	"context"
	"cy/game/codec"
	"cy/game/logic/ddz/desk"
	"cy/game/pb/game"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (p *ddz) QueryDeskInfoReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbgame.QueryDeskInfoReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.QueryDeskInfoReq")
		logrus.Error(err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{"req": *req, "name": args.Name}).Info("recv")

	desk.QueryDeskInfo(args.UserID, req)
	return nil
}
