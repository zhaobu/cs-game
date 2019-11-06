package main

import (
	"context"
	"fmt"
	"game/codec"
	pbclub "game/pb/club"
	pbcommon "game/pb/common"
)

func (p *club) UpdateClubNoticeReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.UpdateClubNoticeReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.UpdateClubNoticeReq")
		tlog.Error(err.Error())
		return
	}

	log.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.UpdateClubNoticeRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			tlog.Error(err.Error())
		}

		//if rsp.Code == 1 {
		//	sendClubChangeInfo(req.ClubID, clubChangeTypUpdate, args.UserID)
		//}
	}()

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 2
		return
	}

	cc.Lock()

	m, find := cc.Members[args.UserID]
	if !find || (m.Identity != identityMaster && m.Identity != identityAdmin) {
		cc.Unlock()
		rsp.Code = 3
		return
	}

	cc.Notice = req.Notice
	cc.noCommit = true
	cc.Unlock()

	rsp.Code = 1

	return
}
