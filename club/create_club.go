package main

import (
	"context"
	"fmt"
	"game/codec"
	"game/db/mgo"
	pbclub "game/pb/club"
	pbcommon "game/pb/common"
	"runtime/debug"
	"time"
)

func (p *club) CreateClubReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.CreateClubReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.CreateClubReq")
		tlog.Error(err.Error())
		return
	}

	log.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.CreateClubRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	createUserID := args.UserID
	var newID int64

	defer func() {
		r := recover()
		if r != nil {
			log.Errorf("%v %s", r, string(debug.Stack()))
		}
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			tlog.Error(err.Error())
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

	newID, err = mgo.AllocClubID()
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
	cc.IsProofe = false
	cc.CurrDayDestoryDeskNum = 0
	cc.LastDestoryDeskNumTime = time.Now().Unix()
	// 创建人默认加入且同意
	cc.Members[createUserID] = &mgo.ClubMember{
		UserID:   createUserID,
		Identity: identityMaster,
		Agree:    true,
		Relation: []uint64{},
	}
	cc.GameArgs = make([]*mgo.DeskSetting, 0)
	for _, v := range req.GameArgs {
		cc.GameArgs = append(cc.GameArgs, &mgo.DeskSetting{
			GameName:        v.GameName,
			GameArgMsgName:  v.GameArgMsgName,
			GameArgMsgValue: v.GameArgMsgValue,
			Enable:          v.Enable,
		})
	}
	if cc.IsAutoCreate {
		cc.f = func() {
			setting, cid, masterUserID := checkAutoCreate(newID)
			if len(setting) > 0 {
				createDesk(setting, cid, masterUserID)
			}
		}
	}
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
