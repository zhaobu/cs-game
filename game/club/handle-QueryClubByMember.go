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

func (p *club) QueryClubByMemberReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.QueryClubByMemberReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.QueryClubByMemberReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.QueryClubByMemberRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}
	}()

	clubDbs, err := mgo.QueryClubByMember(req.UserID)
	if err != nil {
		return
	}
	rsp.Infos = &pbclub.ClubList{}

	for _, v := range clubDbs {
		rsp.Infos.List = append(rsp.Infos.List, clubDb2ClubInfo(v))
	}

	return
}
