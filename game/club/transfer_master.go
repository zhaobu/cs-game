package main

import (
	"context"
	"cy/game/codec"
	"cy/game/db/mgo"
	"cy/game/pb/club"
	"cy/game/pb/common"
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
)

func (p *club) TransferMasterReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.TransferMasterReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.TransferMasterReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.TransferMasterRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}
	}()

	// 不能是自己
	if req.NewMasterUserID == args.UserID {
		rsp.Code = 2
		return
	}

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 3
		return
	}

	cc.Lock()

	if v, find := cc.Members[args.UserID]; !find {
		cc.Unlock()
		rsp.Code = 2
		return
	} else if v.Identity != identityMaster {
		cc.Unlock()
		rsp.Code = 4
		return
	}

	if _, find := cc.Members[req.NewMasterUserID]; !find {
		cc.Unlock()
		rsp.Code = 5
		return
	}

	cc.Unlock()

	if exist, _ := mgo.ExistEmail(bson.M{
		"clubid":  req.ClubID,
		"typ":     emailTypTransferMaster,
		"flag":    0,
		"userid1": args.UserID,
		"userid2": req.NewMasterUserID,
	}); exist {
		rsp.Code = 6
		return
	}

	rsp.Code = 1

	// 给新群主发送邮件
	ce := &mgo.ClubEmail{
		SendTime: time.Now().UTC().Unix(),
		Typ:      emailTypTransferMaster,
		Content:  fmt.Sprintf(`玩家[%d]邀请您成为俱乐部[%d]群主`, args.UserID, req.ClubID),
		Flag:     0,
		ClubID:   req.ClubID,
		UserID1:  args.UserID,
		UserID2:  req.NewMasterUserID,
	}
	mgo.AddClubEmail(ce, req.NewMasterUserID)
	sendClubEmail(ce, req.NewMasterUserID)

	return
}
