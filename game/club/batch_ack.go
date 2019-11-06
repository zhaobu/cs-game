package main

import (
	"context"
	"fmt"
	"game/codec"
	"game/db/mgo"
	pbclub "game/pb/club"
)

func (p *club) BatchAckClubEmail(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.BatchAckClubEmail)
	if !ok {
		err = fmt.Errorf("not *pbclub.BatchAckClubEmail")
		tlog.Error(err.Error())
		return
	}

	log.Infof("recv %s %+v", args.Name, req)

	ids := make([]int64, 0)
	for _, v := range req.Ids {
		ids = append(ids, v.EmailMsgID)
	}
	mgo.BatchSetEmailFlag(ids...)

	return
}
