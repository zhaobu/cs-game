package main

import (
	"context"
	"game/codec"
	"game/db/mgo"
	"game/pb/club"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (p *club) BatchAckClubEmail(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.BatchAckClubEmail)
	if !ok {
		err = fmt.Errorf("not *pbclub.BatchAckClubEmail")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	ids := make([]int64, 0)
	for _, v := range req.Ids {
		ids = append(ids, v.EmailMsgID)
	}
	mgo.BatchSetEmailFlag(ids...)

	return
}
