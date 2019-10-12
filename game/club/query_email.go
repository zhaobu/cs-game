package main

import (
	"context"
	"game/codec"
	"game/db/mgo"
	"game/pb/club"
	"game/pb/common"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (p *club) ClubEmailReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.ClubEmailReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.ClubEmailReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.ClubEmailRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}

	}()

	ces, err := mgo.QueryUserEmail(args.UserID)
	if err != nil {
		return nil
	}

	for _, ce := range ces {
		rsp.Emails = append(rsp.Emails, &pbclub.ClubEmail{
			ID:       ce.ID,
			SendTime: ce.SendTime,
			Typ:      ce.Typ,
			Content:  ce.Content,
			Flag:     ce.Flag,
			ClubID:   ce.ClubID,
		})
	}

	return
}
