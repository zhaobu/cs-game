package main

import (
	"context"
	"cy/game/codec"
	"cy/game/pb/club"
	"cy/game/pb/common"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (p *club) QueryClubMemberReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.QueryClubMemberReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.QueryClubMemberReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.QueryClubMemberRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}
	}()

	cc := getClub(req.ClubID)
	if cc == nil {
		return
	}

	cc.RLock()
	for _, m := range cc.Members {
		cu := mustGetUserOther(m.UserID)
		cu.RLock()
		rsp.Members = append(rsp.Members, &pbclub.MemberInfo{
			UserID:   m.UserID,
			Identity: m.Identity,
			Agree:    m.Agree,
			UserName: cu.UserName,
			Profile:  cu.Profile,
		})
		cu.RUnlock()
	}
	cc.RUnlock()

	return
}
