package main

import (
	"context"
	"cy/game/codec"
	"cy/game/pb/club"
	"cy/game/pb/common"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (p *club) DealMemberIdentityReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.DealMemberIdentityReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.DealMemberIdentityReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.DealMemberIdentityRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}
	}()

	// 有效身份
	if req.Identity != identityAdmin &&
		req.Identity != identityNormal &&
		req.Identity != identityBlack {
		rsp.Code = 4
		return
	}

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 2
		return
	}

	cc.Lock()

	// 权限检查
	var permisOK bool
	if m, find := cc.Members[args.UserID]; find && (m.Identity == identityMaster || m.Identity == identityAdmin) {
		permisOK = true
	}

	if !permisOK {
		rsp.Code = 5
		cc.Unlock()
		return
	}

	// 操作的是群里面的
	if _, find := cc.Members[req.UserID]; !find {
		rsp.Code = 3
		cc.Unlock()
		return
	}

	// 删除
	if req.Del {
		delete(cc.Members, req.UserID)
	} else { // 身份变更
		cc.Members[req.UserID].Identity = req.Identity
	}

	cc.noCommit = true
	cc.Unlock()

	rsp.Code = 1

	if req.Del {
		delUserJoinClub(req.UserID, req.ClubID)
		//sendClubChangeInfo(cc.ID, clubChangeTypExit, req.UserID)
		sendClubChangeInfoByuIds(cc.ID, clubChangeTypExit, req.UserID,[]uint64{req.Head.UserID,req.UserID})
	} else {
		//sendClubChangeInfoByuIds(cc.ID, clubChangeTypUpdate, req.UserID,[]uint64{req.Head.UserID,req.UserID})
	}
	return
}
