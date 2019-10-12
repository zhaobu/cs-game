package main

import (
	"context"
	"game/codec"
	"game/logic/ddz/desk"
	"game/pb/game"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (p *ddz) QueryDeskInfoReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbgame.QueryDeskInfoReq)
	if !ok {
		err = fmt.Errorf("not *pbgame.QueryDeskInfoReq")
		log.Error(err.Error())
		return
	}

	log.WithFields(logrus.Fields{"uid": args.UserID}).Infof("recv %s %+v ", args.Name, *req)

	desk.QueryDeskInfo(args.UserID, req)
	return nil
}
