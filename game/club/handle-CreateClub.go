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

func (p *club) CreateClubReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.CreateClubReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.CreateClubReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	// TODO 限制创建个数

	rsp := &pbclub.CreateClubRsp{}
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

	var clubDb *mgo.ClubDb
	clubDb, err = mgo.CreateClub(req.Name, args.UserID, req.Notice, req.Arg)
	if err != nil {
		return
	}

	// 创建人需加入
	clubDb, err = mgo.JoinClub(clubDb.ID, args.UserID)
	if err != nil {
		return
	}

	rsp.Code = 1

	rsp.Info = clubDb2ClubInfo(clubDb)

	return
}

func clubDb2ClubInfo(clubDb *mgo.ClubDb) *pbclub.ClubInfo {
	info := &pbclub.ClubInfo{
		ID:           clubDb.ID,
		Name:         clubDb.Name,
		CreateUserID: clubDb.CreateUserID,
		Notice:       clubDb.Notice,
		Arg:          clubDb.Arg,
	}

	for _, v := range clubDb.Members {
		// TODO 取其他属性 需优化性能
		cmi := &pbclub.ClubMemberInfo{
			UserID: v,
		}
		info.Members = append(info.Members, cmi)
	}
	return info
}
