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

func (p *club) InviteJoinClubReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.InviteJoinClubReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.InviteJoinClubReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.InviteJoinClubRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}
	}()

	// 加入数量限制
	if checkClubLimit(req.Invitee) {
		rsp.Code = 2
		return
	}

	// 存在
	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 3
		return
	}

	cc.Lock()
	// 限制人数
	if len(cc.Members) > 30 {
		cc.Unlock()
		rsp.Code = 4
		return
	}

	// 已经是成员或者在黑名单中
	if _, find := cc.Members[req.Invitee]; find {
		cc.Unlock()
		rsp.Code = 5
		return
	}

	cc.Unlock()

	if exist, _ := mgo.ExistEmail(bson.M{
		"clubid":  req.ClubID,
		"typ":     emailTypInviteJoinClub,
		"flag":    0,
		"userid2": req.Invitee,
	}); exist {
		rsp.Code = 6
		return
	}

	rsp.Code = 1

	// 给被邀请人发送邮件
	ce := &mgo.ClubEmail{
		SendTime: time.Now().UTC().Unix(),
		Typ:      emailTypInviteJoinClub,
		Content:  fmt.Sprintf(`玩家[%d]邀请您加入俱乐部[%d]`, args.UserID, req.ClubID),
		Flag:     0,
		ClubID:   req.ClubID,
		UserID1:  args.UserID,
		UserID2:  req.Invitee,
	}
	mgo.AddClubEmail(ce, req.Invitee)
	sendClubEmail(ce, req.Invitee)

	return
}
