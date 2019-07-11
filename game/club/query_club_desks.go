package main

import (
	"cy/game/codec"
	"context"
	pbclub "cy/game/pb/club"
	pbcommon "cy/game/pb/common"
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime/debug"
	"time"
)

//查询俱乐部桌子信息
func (p *club) RefreshClubDesks(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.RefreshClubDesks)
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

	rsp.Info = &pbclub.ClubInfo{
		ID:           cc.ID,
		MasterUserID: cc.MasterUserID,
		MasterName:   "", // TODO
		Profile:      cc.Profile,
		IsProofe:cc.IsProofe,
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

	if time.Now().Sub(cc.lastquerytime).Seconds() > 3 {//进行缓存同步
		synchroClubdeskinfo(cc.ID)
		cc.lastquerytime = time.Now()
	}
	for _, d := range cc.desks {
		rsp.Info.Desks = append(rsp.Info.Desks, d)
	}
	rsp.Code = 1
	IsAutoCreate := cc.IsAutoCreate
	IsProofe := cc.IsProofe
	f := cc.f
	cc.RUnlock()

	//在查询时 做一下俱乐部桌子校验 防止游戏服务器重启 自动开放俱乐部的桌子不存在的情况
	if IsAutoCreate && !IsProofe && f == nil {		//自动创建桌子 但是当前不存在桌子
		setting,cid,masterUserID := checkAutoCreate(cc.ID)
		if len(setting) > 0 {
			defer createDesk(setting, cid, masterUserID)
		}
	}
	return
}