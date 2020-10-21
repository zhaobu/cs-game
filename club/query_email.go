package main

import (
	"context"
	"fmt"
	"game/codec"
	"game/db/mgo"
	pbclub "game/pb/club"
	pbcommon "game/pb/common"
)

func (p *club) ClubEmailReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		tlog.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbclub.ClubEmailReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.ClubEmailReq")
		tlog.Error(err.Error())
		return
	}

	log.Infof("recv %s %+v", args.Name, req)

	rsp := &pbclub.ClubEmailRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		err = toGateNormal(rsp, args.UserID)
		if err != nil {
			tlog.Error(err.Error())
		}

	}()

	ces, err := mgo.QueryUserEmail(args.UserID)
	if err != nil {
		return nil
	}

	for _, ce := range ces {
		rsp.Emails = append(rsp.Emails, &pbclub.ClubEmail{
			ID:       ce.ID,
			SendTime: ce.SendTime,
			Typ:      ce.Typ,
			Content:  ce.Content,
			Flag:     ce.Flag,
			ClubID:   ce.ClubID,
		})
	}

	return
}
