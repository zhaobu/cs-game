package main

import (
	"context"
	"cy/game/codec"
	"cy/game/pb/club"
	"cy/game/pb/common"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func (p *club) QueryClubByIDReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.QueryClubByIDReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.QueryClubByIDReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.QueryClubByIDRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		r := recover()
		if r != nil {
			logrus.Errorf("%v %s", r, string(debug.Stack()))
		}
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}
	}()

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 2
		return
	}

	cc.RLock()
	defer cc.RUnlock()
	rsp.Info = &pbclub.ClubInfo{
		ID:           cc.ID,
		MasterUserID: cc.MasterUserID,
		MasterName:   "", // TODO
		Profile:      cc.Profile,
		Base: &pbclub.BaseInfo{
			Name:            cc.Name,
			IsAutoCreate:    cc.IsAutoCreate,
			IsCustomGameArg: cc.IsCustomGameArg,
		},
		Notice: cc.Notice,
	}

	for _, v := range cc.GameArgs {
		rsp.Info.GameArgs = append(rsp.Info.GameArgs, &pbclub.DeskSetting{
			GameName:        v.GameName,
			GameArgMsgName:  v.GameArgMsgName,
			GameArgMsgValue: v.GameArgMsgValue,
			Enable:          v.Enable,
		})
	}

	for _, m := range cc.Members {
		if m.UserID == args.UserID {
			rsp.Agree = m.Agree
		}

		if m.Identity == identityMaster || m.Identity == identityAdmin || m.Identity == identityNormal {
			rsp.Info.MemberCnt++
		}

		uc := mustGetUserOther(m.UserID)
		uc.RLock()
		if uc.Online == 1 { // TODO
			rsp.Info.OnlineCnt++
		}
		uc.RUnlock()
	}

	if m, find := cc.Members[args.UserID]; find {
		rsp.Identity = m.Identity
	}

	for _, d := range cc.desks {
		rsp.Info.Desks = append(rsp.Info.Desks, d)
	}

	rsp.Code = 1
	return
}
