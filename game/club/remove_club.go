package main

import (
	"context"
	"cy/game/codec"
	"cy/game/db/mgo"
	"cy/game/pb/club"
	"cy/game/pb/common"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (p *club) RemoveClubReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.RemoveClubReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.RemoveClubReq")
		logrus.Error(err.Error())
		return
	}

	logrus.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.RemoveClubRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	var notifyUids []uint64
	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			logrus.Error(err.Error())
		}

		if rsp.Code == 1 {
			sendClubRemove(args.UserID, req.ClubID, notifyUids...)
		}
		delClub(req.ClubID)
	}()

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 3
		return
	}

	cc.Lock()
	if cc.MasterUserID != args.UserID {
		cc.Unlock()
		rsp.Code = 2
		return
	}

	for _, m := range cc.Members {
		notifyUids = append(notifyUids, m.UserID)
	}
	cc.Unlock()

	mgo.RemoveClub(req.ClubID) // db 直接删除
	for _, uid := range notifyUids {
		delUserJoinClub(uid, req.ClubID)
	}
	rsp.Code = 1
	return
}

func sendClubRemove(operUid uint64, cid int64, uids ...uint64) {
	if len(uids) == 0 {
		return
	}
	clubChangeInfo := &pbclub.ClubChangeInfo{}
	clubChangeInfo.Typ = int32(clubChangeTypRemove)
	clubChangeInfo.UserID = operUid
	cc := getClub(cid)
	cc.RLock()
	clubChangeInfo.Info = &pbclub.BriefInfo{
		ID:           cid,
		Name:         cc.Name,
		Profile:      cc.Profile,
		MasterUserID: cc.MasterUserID,
	}
	cc.RUnlock()

	toGateNormal(clubChangeInfo, uids...)
}
