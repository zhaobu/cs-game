package main

import (
	"context"
	"cy/game/codec"
	pbclub "cy/game/pb/club"
	pbcommon "cy/game/pb/common"
	pbgame "cy/game/pb/game"
	"fmt"
	"github.com/sirupsen/logrus"
)

func (p *club) MakeDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	req, ok := pb.(*pbclub.MakeDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.UpdateClubNoticeReq")
		logrus.Error(err.Error())
		return
	}
	rsp := &pbgame.MakeDeskRsp{}
	if req.Head != nil {
		rsp.Head = &pbcommon.RspHead{Seq: req.Head.Seq}
	}

	defer func() {
		if rsp.Code != 1 {
			err = toGateNormal(rsp, args.UserID)
			if err != nil {
				logrus.Error(err.Error())
			}
		}
	}()

	cc := getClub(req.ClubID)
	if cc == nil {
		rsp.Code = 7
		return
	}
	cc.Lock()

	if cc.IsProofe {
		rsp.Code = 12
		cc.Unlock()
		return
	}

	m, find := cc.Members[args.UserID]
	if !find || (m.Identity == identityBlack ) {
		cc.Unlock()
		rsp.Code = 11
		return
	}
	desk := cc.desks
	cc.Unlock()
	destorydesks := []*pbcommon.DeskInfo{}
	for _, v := range desk {
		if v.Status == "1" {
			destorydesks = append(destorydesks, v)
		}
	}
	if len(destorydesks) >= 10 {			//空桌子已达上限
		rsp.Code = 8
		return
	}
	cli, err := getGameCli(req.GameName)
	if err != nil {
		rsp.Code = 2					//创建桌子参数错误
		return
	}else{
		rsp.Code = 1
	}
	reqRCall := &codec.Message{}
	codec.Pb2Msg(&pbgame.MakeDeskReq{
		Head:            &pbcommon.ReqHead{UserID: req.Head.UserID},
		GameName:        req.GameName,
		GameArgMsgName:  req.GameArgMsgName,
		GameArgMsgValue: req.GameArgMsgValue,
		ClubID:          req.ClubID,
	}, reqRCall)
	reqRCall.UserID = req.Head.UserID
	rspRCall := &codec.Message{}
	cli.Go(context.Background(), "MakeDeskReq", reqRCall, rspRCall, nil)
	return
}
