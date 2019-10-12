package main

import (
	"context"
	"game/codec"
	pbclub "game/pb/club"
	pbcommon "game/pb/common"
	pbgame "game/pb/game"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

func (p *club)DestroyDeskReq(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	req, ok := pb.(*pbclub.DestroyDeskReq)
	if !ok {
		err = fmt.Errorf("not *pbclub.UpdateClubNoticeReq")
		logrus.Error(err.Error())
		return
	}
	rsp := &pbgame.DestroyDeskRsp{}
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

	m, find := cc.Members[args.UserID]
	if !find || (m.Identity != identityMaster) {
		rsp.Code = 8			//权限不够
		cc.Unlock()
		return
	}
	lasttime := time.Unix(cc.LastDestoryDeskNumTime,0)
	if lasttime.Month() != time.Now().Month() || lasttime.Day() != time.Now().Day(){//重置次数
		cc.CurrDayDestoryDeskNum = 0
	}
	if cc.CurrDayDestoryDeskNum >= 3  {
		rsp.Code = 11					//解散次数已达上限
		cc.Unlock()
		return
	}
	var tdesk *pbcommon.DeskInfo
	for _,v := range cc.desks{
		if v.ID == req.DeskID{
			tdesk = v
		}
	}
	if tdesk == nil{
		rsp.Code = 5			//桌子不存在
		cc.Unlock()
		return
	}
	rsp.Code = 1
	cc.CurrDayDestoryDeskNum ++
	cc.LastDestoryDeskNumTime = time.Now().Unix()
	cc.noCommit = true
	destoryDesk(req.Head.UserID,tdesk)
	cc.Unlock()
	return
}
