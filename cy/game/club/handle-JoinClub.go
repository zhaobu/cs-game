package main

import (
	"context"
	"cy/game/codec"
	"cy/game/db/mgo"
	"cy/game/pb/club"
	"cy/game/pb/common"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (p *club) JoinClubReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.JoinClubReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.JoinClubReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.JoinClubRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	rsp.Code = 2

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}
	}()

	_, err = mgo.JoinClub(req.ClubID, args.UserID)
	if err != nil {
		return
	}
	rsp.Code = 1

	return
}
