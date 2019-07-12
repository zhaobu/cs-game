package main

import (
	"context"
	"cy/game/codec"
	"cy/game/db/mgo"
	"cy/game/pb/club"
	"cy/game/pb/common"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func (p *club) AckClubEmailReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.AckClubEmailReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.AckClubEmailReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.AckClubEmailRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}
	}()

	e, err := mgo.QueryEmail(req.EmailMsgID)
	if err != nil {
		rsp.Code = 2
		return
	}

	if e.Flag != 0 {
		rsp.Code = 3
		return
	}

	switch e.Typ {
	case emailTypJoinClub:
		rsp.Code = ackJoin(e, req)
	case emailTypInviteJoinClub:
		rsp.Code = ackInviteJoin(e, req)
	case emailTypTransferMaster:
		rsp.Code = ackTransferMaster(e, req)
	default:
		rsp.Code = 4
		return
	}
	mgo.SetEmailFlag(e.ID, 1)

	return
}

func ackJoin(e *mgo.ClubEmail, req *pbclub.AckClubEmailReq) int32 {
	joinUID := e.UserID1
	joinClubID := e.ClubID

	if req.Agree {
		cc := getClub(joinClubID)
		if cc == nil {
			return 5
		}

		cc.Lock()
		cc.Members[joinUID] = &mgo.ClubMember{
			UserID:   joinUID,
			Identity: identityNormal,
			Agree:    false,
			Relation:[]uint64{},
		}
		cc.noCommit = true
		cc.Unlock()

		addUserJoinClub(joinUID, joinClubID)
		sendClubChangeInfo(joinClubID, clubChangeTypJoin, joinUID)
	}

	ce := &mgo.ClubEmail{
		SendTime: time.Now().UTC().Unix(),
		Typ:      emailTypTitle,
		Flag:     0,
		ClubID:   joinClubID,
		UserID1:  joinUID,
	}
	if req.Agree {
		ce.Content = fmt.Sprintf(`您成功加入俱乐部[%d]`, joinClubID)
	} else {
		ce.Content = fmt.Sprintf(`您被拒绝加入俱乐部[%d]`, joinClubID)
	}
	mgo.AddClubEmail(ce, joinUID) // 发送给申请人
	sendClubEmail(ce, joinUID)


	if req.Agree {
		ce.Content = fmt.Sprintf(`您同意[%d]加入俱乐部[%d]`,joinUID,joinClubID)
	} else {
		ce.Content = fmt.Sprintf(`您拒绝[%d]加入俱乐部[%d]`,joinUID,joinClubID)
	}
	mgo.AddClubEmail(ce, req.Head.UserID) // 发送给操作用户
	sendClubEmail(ce, req.Head.UserID)
	return 1
}

func ackInviteJoin(e *mgo.ClubEmail, req *pbclub.AckClubEmailReq) int32 {
	inviter := e.UserID1
	invitee := e.UserID2
	joinClubID := e.ClubID

	if req.Agree {
		cc := getClub(joinClubID)
		if cc == nil {
			return 5
		}

		cc.Lock()
		cc.Members[invitee] = &mgo.ClubMember{
			UserID:   invitee,
			Identity: identityNormal,
			Agree:    false,
		}
		cc.noCommit = true
		cc.Unlock()

		addUserJoinClub(invitee, joinClubID)
		sendClubChangeInfo(joinClubID, clubChangeTypJoin, invitee)
	}

	ce := &mgo.ClubEmail{
		SendTime: time.Now().UTC().Unix(),
		Typ:      emailTypTitle,
		Flag:     0,
		ClubID:   joinClubID,
		UserID1:  inviter,
	}
	if req.Agree {
		ce.Content = fmt.Sprintf(`玩家[%d]同意加入俱乐部[%d]`, invitee, joinClubID)
	} else {
		ce.Content = fmt.Sprintf(`玩家[%d]拒绝加入俱乐部[%d]`, invitee, joinClubID)
	}
	mgo.AddClubEmail(ce, inviter) // 发送给邀请人
	sendClubEmail(ce, inviter)

	return 1
}

func ackTransferMaster(e *mgo.ClubEmail, req *pbclub.AckClubEmailReq) int32 {
	oldMaster := e.UserID1
	newMaster := e.UserID2
	clubID := e.ClubID

	if req.Agree {
		cc := getClub(clubID)
		if cc == nil {
			return 5
		}

		cc.Lock()
		if cc.MasterUserID != oldMaster {
			cc.Unlock()
			return 6 // 老的变了
		}

		_, findOld := cc.Members[oldMaster]
		_, findNew := cc.Members[newMaster]

		if !findNew || !findOld {
			cc.Unlock()
			return 7 // 找不到了
		}

		cc.Members[oldMaster].Identity = identityNormal
		cc.Members[newMaster].Identity = identityMaster

		cc.MasterUserID = newMaster
		cc.noCommit = true
		cc.Unlock()

		sendClubChangeInfoByuIds(clubID,clubChangeTypUpdate,newMaster,oldMaster,newMaster)
	}

	ce := &mgo.ClubEmail{
		SendTime: time.Now().UTC().Unix(),
		Typ:      emailTypTitle,
		Flag:     0,
		ClubID:   clubID,
		UserID1:  oldMaster,
	}
	if req.Agree {
		ce.Content = fmt.Sprintf(`玩家[%d]同意成为俱乐部[%d]群主`, newMaster, clubID)
	} else {
		ce.Content = fmt.Sprintf(`玩家[%d]拒绝成为俱乐部[%d]群主`, newMaster, clubID)
	}
	mgo.AddClubEmail(ce, oldMaster) // 发送给群主
	sendClubEmail(ce, oldMaster)

	return 1
}
