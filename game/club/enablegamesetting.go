package main

import (
	"context"
	"game/codec"
	"game/db/mgo"
	"game/pb/club"
	"game/pb/common"
	"fmt"

	"github.com/sirupsen/logrus"
	)

func (p *club) EnableGameSettingReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.EnableGameSettingReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.UpdateClubReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.EnableGameSettingRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}
	
	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}

		if rsp.Code == 1 {
			sendClubChangeInfo(req.ClubID, clubChangeTypUpdateNoTips, args.UserID)
		}
	}()

	if req.Index < 0 || req.Index > 2 {
		rsp.Code = 2
		return
	}

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 3
		return
	}
	cc.Lock()
	permisOK := false
	if m, find := cc.Members[args.UserID]; find && (m.Identity == identityMaster || m.Identity == identityAdmin) {
		permisOK = true
	}
	if !permisOK {	//操作用户权限够
		rsp.Code = 4
		return
	}
	cc.GameArgs[req.Index] =  &mgo.DeskSetting{
		GameName:        req.GameArgs.GameName,
		GameArgMsgName:  req.GameArgs.GameArgMsgName,
		GameArgMsgValue: req.GameArgs.GameArgMsgValue,
		Enable:          req.GameArgs.Enable,
	}
	cc.noCommit = true
	IsAutoCreate := cc.IsAutoCreate
	IsProofe := cc.IsProofe
	f := cc.f
	cc.Unlock()
	rsp.Code = 1
	//更新房间设置时检查是否需要重新创建房间
	if IsAutoCreate&& !IsProofe && f == nil {		//自动创建桌子 但是当前不存在桌子
		setting,cid,masterUserID := checkAutoCreate(cc.ID)
		if len(setting) > 0 {
			defer createDesk(setting, cid, masterUserID)
		}
	}
	return
}
