package main

import (
	"context"
	"fmt"
	"game/codec"
	pbclub "game/pb/club"
	pbcommon "game/pb/common"
)

func (p *club) AgreeClubLawReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.AgreeClubLawReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.AgreeClubLawReq")
		tlog.Error(err.Error())
		return
	}

	log.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.AgreeClubLawRsp{}
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
		return
	}

	cc.Lock()
	if v, find := cc.Members[args.UserID]; find {
		v.Agree = true
	}
	cc.noCommit = true
	cc.Unlock()

	return
}
