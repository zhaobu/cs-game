package main

import (
	"context"
	"fmt"
	"game/codec"
	pbclub "game/pb/club"
	pbcommon "game/pb/common"
)

func (p *club) ExitClubReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.ExitClubReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.ExitClubReq")
		tlog.Error(err.Error())
		return
	}

	log.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.ExitClubRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			tlog.Error(err.Error())
		}

	}()

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 2
		return
	}

	clubChangeInfo := &pbclub.ClubChangeInfo{}
	clubChangeInfo.Typ = int32(clubChangeTypExit)
	clubChangeInfo.UserID = args.UserID

	cc.Lock()
	delete(cc.Members, args.UserID)
	cc.noCommit = true

	clubChangeInfo.Info = &pbclub.BriefInfo{
		ID:           req.ClubID,
		Name:         cc.Name,
		Profile:      cc.Profile,
		MasterUserID: cc.MasterUserID,
	}
	cc.Unlock()

	rsp.Code = 1

	delUserJoinClub(args.UserID, req.ClubID)

	toGateNormal(clubChangeInfo, args.UserID)

	return
}
