package main

import (
	"context"
	"cy/game/codec"
	pbclub "cy/game/pb/club"
	pbcommon "cy/game/pb/common"
	"fmt"
	"runtime/debug"
	"time"

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
	mastercu := mustGetUserOther(cc.MasterUserID)
	rsp.Info = &pbclub.ClubInfo{
		ID:           cc.ID,
		MasterUserID: cc.MasterUserID,
		MasterName:   mastercu.UserName,
		Profile:      cc.Profile,
		Base: &pbclub.BaseInfo{
			Name:            cc.Name,
			IsMasterPay:     cc.IsMasterPay,
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

	if time.Now().Sub(cc.lastquerytime).Seconds() > 3 { //进行缓存同步
		synchroClubdeskinfo(cc.ID)
		cc.lastquerytime = time.Now()
	}

	//在查询时 做一下俱乐部桌子校验 防止游戏服务器重启 自动开放俱乐部的桌子不存在的情况
	if cc.IsAutoCreate && cc.f == nil { //自动创建桌子 但是当前不存在桌子
		haveEmptyTable := false
		for _, v := range cc.desks {
			if v.Status == "1" { //有空桌子
				haveEmptyTable = true
				break
			}
		}
		if !haveEmptyTable { //不存在空桌子
			checkAutoCreate(cc.ID) //自动创建房间
		}
	}
	for _, d := range cc.desks {
		rsp.Info.Desks = append(rsp.Info.Desks, d)
	}
	rsp.Code = 1
	return
}
