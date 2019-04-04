package main

import (
	"context"
	"cy/game/codec"
	"cy/game/db/mgo"
	pbclub "cy/game/pb/club"
	pbcommon "cy/game/pb/common"
	"fmt"
	"runtime/debug"

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

	rsp := &pbclub.CreateClubRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	createUserID := args.UserID
	var newID int64

	defer func() {
		r := recover()
		if r != nil {
			logrus.Errorf("%v %s", r, string(debug.Stack()))
		}
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}

		if rsp.Code == 1 {
			sendClubChangeInfo(newID, clubChangeTypJoin, createUserID)
		}
	}()

	if checkClubLimit(createUserID) {
		rsp.Code = 2
		return nil
	}

	if req.Base.Name == "" {
		rsp.Code = 3
		return
	}

	if req.Base.IsAutoCreate && req.Base.IsCustomGameArg {
		req.Base.IsCustomGameArg = false
	}

	newID, err = mgo.IncClubID()
	if err != nil {
		rsp.Code = 4
		return nil
	}

	cc := newCacheClub()
	cc.ID = newID
	cc.MasterUserID = createUserID
	if uinfo, err := mgo.QueryUserInfo(createUserID); err == nil {
		cc.Profile = uinfo.Profile
	}
	cc.Name = req.Base.Name
	cc.IsAutoCreate = req.Base.IsAutoCreate
	cc.IsCustomGameArg = req.Base.IsCustomGameArg
	cc.IsMasterPay = req.Base.IsMasterPay
	// 创建人默认加入且同意
	cc.Members[createUserID] = &mgo.ClubMember{
		UserID:   createUserID,
		Identity: identityMaster,
		Agree:    true,
	}
	for _, v := range req.GameArgs {
		cc.GameArgs = append(cc.GameArgs, &mgo.DeskSetting{
			GameName:        v.GameName,
			GameArgMsgName:  v.GameArgMsgName,
			GameArgMsgValue: v.GameArgMsgValue,
			Enable:          v.Enable,
		})
	}

	cc.f = func() { checkAutoCreate(newID) }
	cc.noCommit = true

	addClub(cc)
	addUserJoinClub(createUserID, cc.ID)

	rsp.Code = 1

	return
}

func checkClubLimit(userID uint64) bool {
	if len(getUserJoinedClubs(userID)) > 30 {
		return true
	}
	return false
}
