package main

import (
	"context"
	"fmt"
	"game/codec"
	"game/db/mgo"
	pbclub "game/pb/club"
	pbcommon "game/pb/common"
)

func (p *club) UpdateClubReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.UpdateClubReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.UpdateClubReq")
		tlog.Error(err.Error())
		return
	}

	log.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.UpdateClubRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			tlog.Error(err.Error())
		}

		if rsp.Code == 1 {
			sendClubChangeInfo(req.ClubID, clubChangeTypUpdate, args.UserID)
		}
	}()

	if req.Base.Name == "" {
		rsp.Code = 2
		return
	}

	if req.Base.IsAutoCreate && req.Base.IsCustomGameArg {
		req.Base.IsCustomGameArg = false
	}

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 3
		return
	}

	cc.Lock()
	//if cc.MasterUserID != args.UserID {
	//	cc.Unlock()
	//	rsp.Code = 4
	//	return
	//}
	// 权限检查
	permisOK := false
	if m, find := cc.Members[args.UserID]; find && (m.Identity == identityMaster || m.Identity == identityAdmin) {
		permisOK = true
	}
	if !permisOK { //操作用户权限够
		rsp.Code = 4
		return
	}
	cc.Name = req.Base.Name
	cc.IsAutoCreate = req.Base.IsAutoCreate
	cc.IsCustomGameArg = req.Base.IsCustomGameArg
	cc.IsMasterPay = req.Base.IsMasterPay
	cc.GameArgs = make([]*mgo.DeskSetting, 0)
	for _, v := range req.GameArgs {
		cc.GameArgs = append(cc.GameArgs, &mgo.DeskSetting{
			GameName:        v.GameName,
			GameArgMsgName:  v.GameArgMsgName,
			GameArgMsgValue: v.GameArgMsgValue,
			Enable:          v.Enable,
		})
	}
	cc.noCommit = true
	IsProofe := cc.IsProofe
	f := cc.f
	cc.Unlock()
	rsp.Code = 1
	//更新房间设置时检查是否需要重新创建房间
	if req.Base.IsAutoCreate && !IsProofe && f == nil { //自动创建桌子 但是当前不存在桌子
		setting, cid, masterUserID := checkAutoCreate(cc.ID)
		if len(setting) > 0 {
			defer createDesk(setting, cid, masterUserID)
		}
	}
	return
}
