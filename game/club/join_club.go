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

	var toOpers []uint64

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}

		if rsp.Code == 1 {
			// 给申请人发邮件
			ce := &mgo.ClubEmail{
				SendTime: time.Now().UTC().Unix(),
				Typ:      emailTypTitle,
				Content:  fmt.Sprintf(`您已提交申请加入俱乐部[%d]`, req.ClubID),
				Flag:     0,
				ClubID:   req.ClubID,
				UserID1:  args.UserID,
			}
			mgo.AddClubEmail(ce, args.UserID)
			sendClubEmail(ce, args.UserID)
		}
	}()

	// 加入数量限制
	if checkClubLimit(args.UserID) {
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
	if len(cc.Members) > 1000 {
		cc.Unlock()
		rsp.Code = 6
		return
	}

	// 已经是成员或者在黑名单中
	if _, find := cc.Members[args.UserID]; find {
		cc.Unlock()
		rsp.Code = 5
		return
	}

	for _, m := range cc.Members {
		if m.Identity == 1 || m.Identity == 2 {
			toOpers = append(toOpers, m.UserID)
		}
	}

	cc.Unlock()

	if exist, _ := mgo.ExistEmail(bson.M{
		"clubid":  req.ClubID,
		"typ":     emailTypJoinClub,
		"flag":    0,
		"userid1": args.UserID,
	}); exist {
		rsp.Code = 4
		return
	}

	rsp.Code = 1

	// 给管理员发送邮件
	ce := &mgo.ClubEmail{
		SendTime: time.Now().UTC().Unix(),
		Typ:      emailTypJoinClub,
		Content:  fmt.Sprintf(`[%d]申请加入俱乐部[%d]`, args.UserID, req.ClubID),
		Flag:     0,
		ClubID:   req.ClubID,
		UserID1:  args.UserID,
	}
	mgo.AddClubEmail(ce, toOpers...)
	sendClubEmail(ce, toOpers...)

	return
}
